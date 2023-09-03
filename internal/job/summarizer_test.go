package job

import (
	"context"
	"errors"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
)

func setup() {
	lg, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(lg)

	config.FeatureFlags = config.Flags{
		IsSummarizerJobEnabled: true,
	}
}

func TestSummarizerJob_Run(t *testing.T) {
	setup()

	testCases := []struct {
		name           string
		pre            func(t *testing.T, c api.Summarizer, expErr error)
		in             []entity.Message
		expected       []entity.SummaryMsg
		expectedErrors []error
	}{
		{
			name: "context deadline",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, expErr error) {
				c.(*mocks.Summarizer).On("GetMsgSummary", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
					}).Return("", expErr)
			},
			expectedErrors: []error{context.DeadlineExceeded},
		},
		{
			name: "summary success",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, expErr error) {
				c.(*mocks.Summarizer).On("GetMsgSummary", mock.Anything, mock.Anything).
					Return("summary", expErr)
			},
			expected: []entity.SummaryMsg{
				*entity.NewSummaryMsg(
					entity.NewMsg("1", "i4u", "kek", true),
					"summary",
				),
			},
		},
		{
			name: "not internship request",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", false,
				),
			},
			pre: func(t *testing.T, c api.Summarizer, expErr error) {},
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errsCh, in, out := make(chan error), make(chan entity.Message), make(chan entity.SummaryMsg)
			defer close(errsCh)
			defer close(in)
			defer close(out)

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-time.After(6 * time.Second)
				cancel()
			}()

			var expErr error
			if len(tc.expectedErrors) > 0 {
				expErr = tc.expectedErrors[0]
			}

			client := mocks.NewSummarizer(t)
			tc.pre(t, client, expErr)
			go NewSummarizerJob(client, errsCh, in, out).Run(ctx)
			go func() {
				for _, msg := range tc.in {
					in <- msg
				}
			}()

			var outIndex, errIndex int
		verifyLoop:
			for {
				select {
				case err := <-errsCh:
					if !errors.Is(err, tc.expectedErrors[errIndex]) {
						require.Equal(t, tc.expectedErrors[errIndex], err)
					}
					errIndex++
				case msg := <-out:
					require.Equal(t, tc.expected[outIndex], msg)
					outIndex++
				case <-ctx.Done():
					break verifyLoop
				}
			}

			client.AssertExpectations(t)
		})
	}
}
