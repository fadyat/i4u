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

type summaryJobTestcase struct {
	name          string
	pre           func(t *testing.T, c api.Summarizer, tc summaryJobTestcase)
	in            []entity.Message
	expected      []entity.SummaryMsg
	expectedError error
}

func TestSummarizerJob_Run(t *testing.T) {
	testCases := []summaryJobTestcase{
		{
			name: "context deadline",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, tc summaryJobTestcase) {
				c.(*mocks.Summarizer).On("GetMsgSummary", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
					}).Return("", tc.expectedError)
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name: "summary success",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, tc summaryJobTestcase) {
				c.(*mocks.Summarizer).On("GetMsgSummary", mock.Anything, mock.Anything).
					Return("summary", nil)
			},
			expected: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("0", "i4u", "kek", true),
					"summary",
				),
			},
		},
		{
			name: "not internship request",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", false,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, tc summaryJobTestcase) {},
		},
		{
			name: "summary for multiple messages",
			in: func() []entity.Message {
				var size = 100
				var msgs = make([]entity.Message, size)

				for i := 0; i < size; i++ {
					msgs[i] = entity.NewMsg(
						fmt.Sprintf("%d", i), "i4u", fmt.Sprintf("%d", i), true,
					)
				}

				return msgs
			}(),
			pre: func(t *testing.T, c api.Summarizer, tc summaryJobTestcase) {
				for i := 0; i < len(tc.in); i++ {
					c.(*mocks.Summarizer).On("GetMsgSummary", mock.Anything, tc.in[i]).
						Return(fmt.Sprintf("summary%d", i), nil)
				}
			},
			expected: func() []entity.SummaryMsg {
				var size = 100
				var msgs = make([]entity.SummaryMsg, size)

				for i := 0; i < size; i++ {
					msgs[i] = *entity.NewSummaryMsg(
						entity.NewMsg(
							fmt.Sprintf("%d", i), "i4u", fmt.Sprintf("%d", i), true,
						),
						fmt.Sprintf("summary%d", i),
					)
				}

				return msgs
			}(),
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errsCh, in, out := make(chan error), make(chan entity.Message), make(chan entity.SummaryMsg)
			defer close(in)

			summarizer := mocks.NewSummarizer(t)
			tc.pre(t, summarizer, tc)

			var (
				jobWg                 syncs.WaitGroup
				inputWg               syncs.WaitGroup
				jobContext, cancelJob = context.WithCancel(context.Background())
			)
			jobWg.Go(func() {
				defer close(out)
				defer close(errsCh)

				NewSummarizerJob(summarizer, errsCh, in, out).Run(jobContext)
			})

			for _, msg := range tc.in {
				m := msg
				inputWg.Go(func() { in <- m })
			}

			verifyContext, cancelVerify := context.WithCancel(context.Background())
			go func() {
				for {
					select {
					case err, ok := <-errsCh:
						if !ok {
							continue
						}

						if !errors.Is(err, tc.expectedError) {
							assert.Equal(t, tc.expectedError, err)
						}
					case msg, ok := <-out:
						if !ok {
							continue
						}

						assert.Equal(t, tc.expected[parseInt(t, msg.ID())], msg)
					case <-verifyContext.Done():
						return
					}
				}
			}()

			inputWg.Wait()
			cancelJob()
			jobWg.Wait()
			cancelVerify()
		})
	}
}
