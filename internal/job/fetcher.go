package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
	"time"
)

type MessageFetcherJob struct {
	client api.Mail
	out    []chan<- entity.Message
	errsCh chan<- error
}

func NewFetcherJob(
	mailClient api.Mail,
	errsCh chan<- error,
	out []chan<- entity.Message,
) Job {
	return &MessageFetcherJob{
		client: mailClient,
		errsCh: errsCh,
		out:    out,
	}
}

// Run launches the stage of fetching messages from the mail client
// with parsing to the internal message format.
func (m *MessageFetcherJob) Run(ctx context.Context) {
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for wrap := range m.client.GetUnreadMsgs(timeout) {
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
