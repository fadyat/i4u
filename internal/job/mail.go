package job

import (
	"context"
	"github.com/fadyat/i4u/api"
	"time"
)

type Job interface {
	Run(context.Context) error
}

type Mail struct {
	client api.Mail
}

func NewMailJob(client api.Mail) Job {
	return &Mail{client: client}
}

func (m *Mail) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.client.GetUnreadMsgs(ctx)
		}
	}
}
