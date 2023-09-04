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
			go func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				s.send(timeout, &msg)
			}()
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
		s.errsCh <- fmt.Errorf("failed to send message: %w", err)
		return
	}

	zap.S().Debugf("message %s was delivered successfully", msg.ID())
}
