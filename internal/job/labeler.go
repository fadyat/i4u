package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
)

type LabelerJob struct {
	client api.Mail

	in     <-chan entity.Message
	errsCh chan<- error
}

func NewLabelerJob(
	client api.Mail,
	errsCh chan<- error,
	in <-chan entity.Message,
) Job {
	return &LabelerJob{
		client: client,
		in:     in,
		errsCh: errsCh,
	}
}

func (l *LabelerJob) Run(ctx context.Context) {
	for {
		select {
		case msg := <-l.in:
			// todo: add cancellation for context
			l.labeling(ctx, msg)
		case <-ctx.Done():
			return
		}
	}
}

func (l *LabelerJob) labeling(ctx context.Context, msg entity.Message) {
	if !config.FeatureFlags.IsLabelerJobEnabled {
		zap.S().Debugf("got message, but labeler job is disabled")
		return
	}

	if err := l.client.LabelMsg(ctx, msg); err != nil {
		l.errsCh <- fmt.Errorf("labeling failed with: %s", err)
		return
	}

	zap.S().Debugf("labeled message: %s with label: %s", msg.ID(), msg.Label())
}
