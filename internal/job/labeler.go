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
	var wg syncs.WaitGroup

	for {
		select {
		case msg := <-l.in:
			wg.Go(func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				l.labeling(timeout, msg)
			})
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

// labeling making api call to mail provider and labels the message
// with the appropriate label.
// launched when want to mark the message as read or with result
// of the message analysis.
func (l *LabelerJob) labeling(ctx context.Context, msg entity.Message) {
	if !config.FeatureFlags.IsLabelerJobEnabled {
		zap.S().Debugf("got message %s, but labeler job is disabled", msg.ID())
		return
	}

	if err := l.client.LabelMsg(ctx, msg); err != nil {
		l.errsCh <- fmt.Errorf("labeling failed with: %w", err)
		return
	}

	zap.S().Debugf("labeled message: %s with label: %s", msg.ID(), msg.Label())
}
