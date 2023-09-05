package job

import (
	"context"
	"errors"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/mocks"
	"github.com/fadyat/i4u/pkg/syncs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type labelerJobTestcase struct {
	name          string
	in            []entity.Message
	pre           func(t *testing.T, c api.Mail, tc labelerJobTestcase)
	expectedError error
}

func TestLabelerJob_Run(t *testing.T) {
	testCases := []labelerJobTestcase{
		{
			name: "context deadline",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Mail, tc labelerJobTestcase) {
				c.(*mocks.Mail).On("LabelMsg", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
					}).Return(tc.expectedError)
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name: "label success",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Mail, tc labelerJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Mail).On("LabelMsg", mock.Anything, msg).
						Return(nil)
				}
			},
		},
		{
			name: "label error",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Mail, tc labelerJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Mail).On("LabelMsg", mock.Anything, msg).
						Return(tc.expectedError)
				}
			},
			expectedError: errors.New("labeling failed with: label error"),
		},
		{
			name: "labeling multiple messages",
			in: func() []entity.Message {
				size := 100
				msgs := make([]entity.Message, size)
				for i := 0; i < size; i++ {
					msgs[i] = entity.NewMsg(
						"1", "i4u", "kek", true,
					)
				}

				return msgs
			}(),
			pre: func(t *testing.T, c api.Mail, tc labelerJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Mail).On("LabelMsg", mock.Anything, msg).
						Return(nil)
				}
			},
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			labeler := mocks.NewMail(t)
			tc.pre(t, labeler, tc)

			errCh, in := make(chan error), make(chan entity.Message)
			defer close(in)

			var (
				jobWg              syncs.WaitGroup
				inputWg            syncs.WaitGroup
				jobContext, cancel = context.WithCancel(context.Background())
			)

			jobWg.Go(func() {
				defer close(errCh)

				NewLabelerJob(labeler, errCh, in).Run(jobContext)
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
		})
	}
}
