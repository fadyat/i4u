package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/mocks"
	"github.com/fadyat/i4u/pkg/syncs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type senderJobTestcase struct {
	name          string
	in            []entity.SummaryMsg
	pre           func(t *testing.T, c api.Sender, tc senderJobTestcase)
	expectedError error
}

func TestSenderJob_Run(t *testing.T) {
	testCases := []senderJobTestcase{
		{
			name: "context deadline",
			in: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("0", "i4u", "kek", true),
					"summary",
				),
			},
			pre: func(t *testing.T, c api.Sender, tc senderJobTestcase) {
				c.(*mocks.Sender).On("Send", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
					}).Return(tc.expectedError)
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name: "send success",
			in: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("0", "i4u", "kek", true),
					"summary",
				),
			},
			pre: func(t *testing.T, c api.Sender, tc senderJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Sender).On("Send", mock.Anything, &msg).
						Return(nil)
				}
			},
		},
		{
			name: "send error",
			in: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("0", "i4u", "kek", true),
					"summary",
				),
			},
			pre: func(t *testing.T, c api.Sender, tc senderJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Sender).On("Send", mock.Anything, &msg).
						Return(tc.expectedError)
				}
			},
			expectedError: fmt.Errorf("send error"),
		},
		{
			name: "send for multiple messages",
			in: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("0", "i4u", "kek", true),
					"summary",
				),
				*entity.NewSummaryMsg(
					entity.NewMsg("1", "i4u", "aboba", true),
					"summary",
				),
				*entity.NewSummaryMsg(
					entity.NewMsg("2", "i4u", "kekis", true),
					"summary",
				),
			},
			pre: func(t *testing.T, c api.Sender, tc senderJobTestcase) {
				c.(*mocks.Sender).On("Send", mock.Anything, mock.Anything).
					Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sender := mocks.NewSender(t)
			tc.pre(t, sender, tc)

			errCh, in := make(chan error), make(chan entity.SummaryMsg)
			defer close(in)

			var (
				jobWg              syncs.WaitGroup
				inputWg            syncs.WaitGroup
				jobContext, cancel = context.WithCancel(context.Background())
			)
			jobWg.Go(func() {
				defer close(errCh)

				NewSenderJob(sender, errCh, in).Run(jobContext)
			})

			for _, msg := range tc.in {
				m := msg
				inputWg.Go(func() { in <- m })
			}

			go func() {
				for err := range errCh {
					if !errors.Is(err, tc.expectedError) {
						assert.NoError(t, err)
					}
				}
			}()

			inputWg.Wait()
			cancel()
			jobWg.Wait()

			sender.AssertExpectations(t)
		})
	}
}
