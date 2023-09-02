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

type SummarizerJob struct {
	client api.Summarizer

	in  <-chan entity.Message
	out chan<- entity.SummaryMsg

	errsCh chan<- error
}

func NewSummarizerJob(
	summarizer api.Summarizer,
	errsCh chan<- error,
	in <-chan entity.Message,
	out chan<- entity.SummaryMsg,
) Job {
	return &SummarizerJob{
		client: summarizer,
		in:     in,
		out:    out,
		errsCh: errsCh,
	}
}

func (s *SummarizerJob) Run(ctx context.Context) {
	for {
		select {
		case msg := <-s.in:
			go func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				s.summary(timeout, msg)
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (s *SummarizerJob) summary(ctx context.Context, msg entity.Message) {
	if !config.FeatureFlags.IsSummarizerJobEnabled {
		zap.S().Debugf("got message %v, but summarizer job is disabled", msg)
		return
	}

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
