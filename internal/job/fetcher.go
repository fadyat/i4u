package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
)

// MessageFetcherJob is the job responsible for fetching and
// parsing messages from the mail client.
//
// After parsing, the job will push the message to the next
// jobs in the pipeline.
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

func (m *MessageFetcherJob) Run(ctx context.Context) {
	for wrap := range m.client.GetUnreadMsgs(ctx) {
		if wrap.Err != nil {
			m.errsCh <- fmt.Errorf("failed to fetch message: %w", wrap.Err)
			continue
		}

		m.pushSingle(wrap.Msg)
	}

	zap.S().Debugf("current stage for fetching messages is done")
}

func (m *MessageFetcherJob) pushSingle(msg entity.Message) {
	for _, o := range m.out {
		go func(o chan<- entity.Message) { o <- msg }(o)
	}
}
