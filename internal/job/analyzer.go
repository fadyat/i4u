package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
	"time"
)

type MessageAnalyzerJob struct {
	client       api.Analyzer
	labelsMapper *config.LabelsMapper

	in  <-chan entity.Message
	out []chan<- entity.Message

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
	for {
		select {
		case msg := <-m.in:
			go func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				m.analyze(timeout, msg)
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (m *MessageAnalyzerJob) analyze(ctx context.Context, msg entity.Message) {
	if !config.FeatureFlags.IsAnalyzerJobEnabled {
		zap.S().Debugf("got message %v, but analyzer job is disabled", msg)
		return
	}

	// notifying the user that the message is empty, and we can't analyze it
	// error may happen, when you have dialog with someone, and you reply to
	// the message.
	// because, parsing don't work well with that.
	if msg.Body() == "" {
		m.errsCh <- fmt.Errorf("got empty body for message: %s", msg.Link())
		return
	}

	isIntern, err := m.client.IsInternshipRequest(ctx, msg)
	if err != nil {
		m.errsCh <- fmt.Errorf("failed to analyze message: %w", err)
		return
	}

	_, ok := msg.(*entity.Msg)
	if !ok {
		m.errsCh <- fmt.Errorf("unknown message type: %T", msg)
		return
	}

	msg = msg.(*entity.Msg).Copy().WithIsIntern(isIntern).
		WithLabel(m.labelsMapper.GetInternLabel(isIntern))

	zap.S().Debugf("analyzed message: %s, isIntern: %v", msg.ID(), isIntern)
	for _, o := range m.out {
		go func(o chan<- entity.Message) { o <- msg }(o)
	}
}
