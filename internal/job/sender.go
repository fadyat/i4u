package job

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
)

type SenderJob struct {
	client api.Sender

	in     <-chan entity.SummaryMsg
	errsCh chan<- error
}

func NewSenderJob(
	sender api.Sender,
	errsCh chan<- error,
	in <-chan entity.SummaryMsg,
) Job {
	return &SenderJob{
		client: sender,
		in:     in,
		errsCh: errsCh,
	}
}

func (s *SenderJob) Run(ctx context.Context) {
	for {
		select {
		case msg := <-s.in:
			// todo: add cancellation for context
			s.send(ctx, &msg)
		case <-ctx.Done():
			return
		}
	}
}

func (s *SenderJob) send(ctx context.Context, msg *entity.SummaryMsg) {
	if !config.FeatureFlags.IsSenderJobEnabled {
		zap.S().Debugf("got message %v, but sender job is disabled", msg)
		return
	}

	if err := s.client.Send(ctx, msg); err != nil {
		s.errsCh <- err
	}
}
