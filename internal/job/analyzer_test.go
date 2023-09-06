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

type analyzerJobTestcase struct {
	name          string
	in            []entity.Message
	pre           func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase)
	expectedOut   []entity.Message
	expectedError error
	outputChSz    int
}

func TestMessageAnalyzerJob_Run(t *testing.T) {
	testCases := []analyzerJobTestcase{
		{
			name: "empty body",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "", true,
				),
			},
			pre:           func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {},
			expectedError: fmt.Errorf("got empty body for message: %s", "0"),
		},
		{
			name: "context deadline",
			in: []entity.Message{
				entity.NewMsg(
					"1", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
					}).Return(false, tc.expectedError)
			},
			expectedError: context.DeadlineExceeded,
		},
		{
			name: "failed to get analysis",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, mock.Anything).
					Return(false, tc.expectedError)
			},
			expectedError: fmt.Errorf("failed to analyze message: %s", "0"),
		},
		{
			name: "unsupported message type",
			in: []entity.Message{
				struct {
					entity.Message
				}{
					entity.NewMsg(
						"0", "i4u", "kek", true,
					),
				},
			},
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, mock.Anything).
					Return(false, nil)
			},
			expectedError: fmt.Errorf("unknown message type: %T", struct{ entity.Message }{}),
		},
		{
			name: "not intern message",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", false,
				),
			},
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, mock.Anything).
					Return(false, nil)
			},
			outputChSz: 1,
			expectedOut: []entity.Message{
				entity.NewMsg(
					"0", "not_intern", "kek", false,
				),
			},
		},
		{
			name: "intern message to multiple channels",
			in: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", true,
				),
			},
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, mock.Anything).
					Return(true, nil)
			},
			outputChSz: 2,
			expectedOut: []entity.Message{
				entity.NewMsg(
					"0", "is_intern", "kek", true,
				),
			},
		},
		{
			name: "multiple messages to multiple channels",
			in: func() []entity.Message {
				size := 500
				var msgs = make([]entity.Message, 0, size)
				for i := 0; i < size; i++ {
					msgs = append(msgs, entity.NewMsg(
						fmt.Sprintf("%d", i), "i4u", "kek", i%2 == 0,
					))
				}

				return msgs
			}(),
			pre: func(t *testing.T, c api.Analyzer, tc analyzerJobTestcase) {
				for _, msg := range tc.in {
					c.(*mocks.Analyzer).On("IsInternshipRequest", mock.Anything, msg).
						Return(msg.IsInternshipRequest(), nil)
				}
			},
			outputChSz: 5,
			expectedOut: func() []entity.Message {
				size := 500
				var msgs = make([]entity.Message, 0, size)
				for i := 0; i < size; i++ {
					msgs = append(msgs, entity.NewMsg(
						fmt.Sprintf("%d", i), func() string {
							if i%2 == 0 {
								return "is_intern"
							}

							return "not_intern"
						}(), "kek", i%2 == 0,
					))
				}

				return msgs
			}(),
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lm := newLabelsMapper()
			analyzer := mocks.NewAnalyzer(t)

			tc.pre(t, analyzer, tc)

			var (
				errsCh = make(chan error)
				in     = make(chan entity.Message)
				out    = make([]chan entity.Message, tc.outputChSz)
				outW   = make([]chan<- entity.Message, tc.outputChSz)

				jobWg                 syncs.WaitGroup
				jobContext, jobCancel = context.WithCancel(context.Background())

				inputWg                     syncs.WaitGroup
				outputWg                    syncs.WaitGroup
				verifyContext, cancelVerify = context.WithCancel(context.Background())
			)
			defer close(in)

			for i := 0; i < tc.outputChSz; i++ {
				out[i] = make(chan entity.Message)
				outW[i] = out[i]
			}

			jobWg.Go(func() {
				defer func() {
					for _, ch := range out {
						close(ch)
					}
				}()
				defer close(errsCh)

				NewAnalyzerJob(analyzer, lm, errsCh, in, outW).Run(jobContext)
			})

			for _, msg := range tc.in {
				m := msg
				inputWg.Go(func() { in <- m })
			}

			for _, c := range out {
				ch := c

				outputWg.Go(func() {
					isAppeared := make(map[string]bool)
					for _, msg := range tc.expectedOut {
						isAppeared[msg.ID()] = false
					}

					defer func() {
						for _, msg := range tc.expectedOut {
							if !isAppeared[msg.ID()] {
								assert.Fail(t, "message did not appear")
							}
						}
					}()

					for {
						select {
						case msg, ok := <-ch:
							if !ok {
								continue
							}

							if isAppeared[msg.ID()] {
								assert.Fail(t, "message appeared twice")
							}

							assert.Equal(t, tc.expectedOut[parseInt(t, msg.ID())], msg)
							isAppeared[msg.ID()] = true
						case <-verifyContext.Done():
							return
						}
					}
				})
			}

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
					case <-verifyContext.Done():
						return
					}
				}
			}()

			inputWg.Wait()
			jobCancel()
			jobWg.Wait()
			cancelVerify()
			outputWg.Wait()
		})
	}
}
