package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/pkg/syncs"
	"go.uber.org/zap"
	"time"
)

type MessageAnalyzerJob struct {
	client       api.Analyzer
	labelsMapper *config.LabelsMapper

	in     <-chan entity.Message
	out    []chan<- entity.Message
	errsCh chan<- error
}

func NewAnalyzerJob(
	analyzer api.Analyzer,
	labels *config.LabelsMapper,
	errsCh chan<- error,
	in <-chan entity.Message,
	out []chan<- entity.Message,
) Job {
	return &MessageAnalyzerJob{
		client:       analyzer,
		labelsMapper: labels,
		in:           in,
		out:          out,
		errsCh:       errsCh,
	}
}

func (m *MessageAnalyzerJob) Run(ctx context.Context) {
	var wg syncs.WaitGroup

	for {
		select {
		case msg := <-m.in:
			wg.Go(func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				m.analyze(timeout, &wg, msg)
			})
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

// analyze gets the message and sends it to the analyzer API
// to determine whether the message is an internship request or not.
func (m *MessageAnalyzerJob) analyze(
	ctx context.Context, wg *syncs.WaitGroup, msg entity.Message,
) {
	if !config.FeatureFlags.IsAnalyzerJobEnabled {
		zap.S().Debugf("got message %s, but analyzer job is disabled", msg.ID())
		return
	}

	// notifying the user that the message is empty, and we can't analyze it
	// error may happen, when you have dialog with someone, and you reply to the message.
	// because, parsing don't work well with that.
	if msg.Body() == "" {
		m.errsCh <- fmt.Errorf("got empty body for message: %s", msg.ID())
		return
	}

	isIntern, err := m.client.IsInternshipRequest(ctx, msg)
	if err != nil {
		m.errsCh <- fmt.Errorf("failed to analyze message: %w", err)
		return
	}

	if _, ok := msg.(*entity.Msg); !ok {
		m.errsCh <- fmt.Errorf("unknown message type: %T", msg)
		return
	}

	// todo: think about the better way to do this
	msg = msg.(*entity.Msg).Copy().WithIsIntern(isIntern).
		WithLabel(m.labelsMapper.GetInternLabel(isIntern))

	for _, o := range m.out {
		out := o
		wg.Go(func() { out <- msg })
	}

	zap.S().Debugf("analyzed message: %s, isIntern: %v", msg.ID(), isIntern)
}
