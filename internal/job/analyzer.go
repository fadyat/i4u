package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
)

type MessageAnalyzerJob struct {
	client api.Analyzer

	in  <-chan entity.Message
	out []chan<- entity.Message

	errsCh chan<- error
}

func NewAnalyzerJob(
	analyzer api.Analyzer,
	errsCh chan<- error,
	in <-chan entity.Message,
	out []chan<- entity.Message,
) Job {
	return &MessageAnalyzerJob{
		client: analyzer,
		in:     in,
		out:    out,
		errsCh: errsCh,
	}
}

func (m *MessageAnalyzerJob) Run(ctx context.Context) {
	for {
		select {
		case msg := <-m.in:
			// todo: add cancellation of context?
			m.analyze(ctx, msg)
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

	isIntern, err := m.client.IsInternshipRequest(ctx, msg)
	if err != nil {
		m.errsCh <- fmt.Errorf("failed to analyze message: %s", err)
		return
	}

	_, ok := msg.(*entity.Msg)
	if !ok {
		m.errsCh <- fmt.Errorf("unknown message type: %T", msg)
		return
	}

	// todo: get from config
	var isInternLabel = "Label_11"
	if isIntern {
		isInternLabel = "Label_13"
	}

	msg = msg.(*entity.Msg).Copy().WithIsIntern(isIntern).
		WithLabel(isInternLabel)

	zap.S().Debugf("analyzed message: %s, isIntern: %v", msg.ID(), isIntern)
	for _, o := range m.out {
		go func(o chan<- entity.Message) { o <- msg }(o)
	}
}
