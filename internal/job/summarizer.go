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

type SummarizerJob struct {
	client api.Summarizer

	in     <-chan entity.Message
	out    chan<- entity.SummaryMsg
	errsCh chan<- error
}

func NewSummarizerJob(
	client api.Summarizer,
	errsCh chan<- error,
	in <-chan entity.Message,
	out chan<- entity.SummaryMsg,
) Job {
	return &SummarizerJob{
		client: client,
		in:     in,
		out:    out,
		errsCh: errsCh,
	}
}

func (s *SummarizerJob) Run(ctx context.Context) {
	var wg syncs.WaitGroup

	for {
		select {
		case msg := <-s.in:
			wg.Go(func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				s.summary(timeout, msg)
			})
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

// summary gets the main information from the message and sends it to the
// summarizer API. It then sends the summary to the output channel.
func (s *SummarizerJob) summary(ctx context.Context, msg entity.Message) {
	if !config.FeatureFlags.IsSummarizerJobEnabled {
		zap.S().Debugf("got message %s, but summarizer job is disabled", msg.ID())
		return
	}

	// other jobs don't make filtering, because they don't care about the
	// message type. But the summarizer job does, because it only cares
	// about internship requests.
	if !msg.IsInternshipRequest() {
		zap.S().Debugf("got message %s, but it is not an internship request", msg.ID())
		return
	}

	summary, err := s.client.GetMsgSummary(ctx, msg)
	if err != nil {
		s.errsCh <- fmt.Errorf("failed to get msg summary: %w", err)
		return
	}

	zap.S().Debugf("got summary for message %s", msg.ID())
	s.out <- *entity.NewSummaryMsg(msg, summary)
}
