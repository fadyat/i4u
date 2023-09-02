package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
)

// MessageFetcherJob is the job responsible for fetching new messages
// from mail provider, gmail in current case.
type MessageFetcherJob struct {
	client api.Mail

	// After message get processed it pushing it to few channels:
	// - labeler channel
	// - analyzer channel
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
	msgs, err := m.run(ctx)
	if err != nil {
		m.errsCh <- err
		return
	}

	for _, msg := range msgs {
		m.pushSingle(msg)
	}
}

func (m *MessageFetcherJob) parseMsg(msg *gmail.Message) (entity.Message, error) {
	parsed, err := entity.NewMsgFromGmailMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %s", err)
	}

	zap.S().Debugf("parsed message: %s", parsed.ID())
	return parsed, nil
}

func (m *MessageFetcherJob) pushSingle(msg *gmail.Message) {
	parsed, err := m.parseMsg(msg)
	if err != nil {
		m.errsCh <- err
		return
	}

	for _, o := range m.out {
		go func(o chan<- entity.Message) { o <- parsed }(o)
	}
}

func (m *MessageFetcherJob) run(ctx context.Context) ([]*gmail.Message, error) {
	msgs, err := m.client.GetUnreadMsgs(ctx)
	zap.S().Debugf("got %d messages", len(msgs))
	return msgs, err
}
