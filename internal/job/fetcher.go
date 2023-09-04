package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/pkg/syncs"
	"go.uber.org/zap"
	"time"
)

type MessageFetcherJob struct {
	client api.Mail
	period time.Duration

	out    []chan<- entity.Message
	errsCh chan<- error
}

func NewFetcherJob(
	mailClient api.Mail,
	period time.Duration,
	errsCh chan<- error,
	out []chan<- entity.Message,
) Job {
	return &MessageFetcherJob{
		client: mailClient,
		period: period,
		errsCh: errsCh,
		out:    out,
	}
}

func (m *MessageFetcherJob) Run(ctx context.Context) {
	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	var wg syncs.WaitGroup
	for {
		select {
		case <-ticker.C:
			wg.Go(func() {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				m.fetch(timeout)
			})
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

// fetch getting unread messages from mail provider and push them to the next stage
// with parsing to the internal message format.
func (m *MessageFetcherJob) fetch(ctx context.Context) {
	for wrap := range m.client.GetUnreadMsgs(ctx) {
		if wrap.Err != nil {
			m.errsCh <- fmt.Errorf("failed to fetch message: %w", wrap.Err)
			continue
		}

		m.pushSingle(wrap.Msg)
	}

	zap.S().Debug("message fetcher job finished")
}

// pushSingle pushes the message to all channels in separate goroutines,
// to avoid blocking the main thread.
func (m *MessageFetcherJob) pushSingle(msg entity.Message) {
	for _, o := range m.out {
		go func(o chan<- entity.Message) { o <- msg }(o)
	}

	zap.S().Debugf("message %s was pushed to the next stage", msg.ID())
}
