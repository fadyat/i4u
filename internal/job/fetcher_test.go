package job

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/mocks"
	"github.com/fadyat/i4u/pkg/syncs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type fetcherJobTestcase struct {
	name           string
	pre            func(t *testing.T, c api.Mail, tc fetcherJobTestcase)
	expectedOut    []entity.Message
	expectedErrors map[string]bool
	outputChSz     int
}

func TestMessageFetcherJob_Run(t *testing.T) {
	testCases := []fetcherJobTestcase{
		{
			name: "context deadline",
			pre: func(t *testing.T, c api.Mail, tc fetcherJobTestcase) {
				outCh := make(chan entity.MessageWithError, 5)

				c.(*mocks.Mail).On("GetUnreadMsgs", mock.Anything).
					Run(func(args mock.Arguments) {
						defer close(outCh)

						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
						outCh <- entity.MessageWithError{Err: context.DeadlineExceeded}
					}).Return((<-chan entity.MessageWithError)(outCh))
			},
			expectedErrors: map[string]bool{
				fmt.Errorf("failed to fetch message: %w", context.DeadlineExceeded).Error(): false,
			},
		},
		{
			name: "failed single message",
			pre: func(t *testing.T, c api.Mail, tc fetcherJobTestcase) {
				outCh := make(chan entity.MessageWithError, 5)

				c.(*mocks.Mail).On("GetUnreadMsgs", mock.Anything).
					Run(func(args mock.Arguments) {
						defer close(outCh)

						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
						outCh <- entity.MessageWithError{Err: fmt.Errorf("kek")}
					}).Return((<-chan entity.MessageWithError)(outCh))
			},
			expectedErrors: map[string]bool{
				fmt.Errorf("failed to fetch message: %w", fmt.Errorf("kek")).Error(): false,
			},
		},
		{
			name: "success single message",
			pre: func(t *testing.T, c api.Mail, tc fetcherJobTestcase) {
				outCh := make(chan entity.MessageWithError, 5)

				c.(*mocks.Mail).On("GetUnreadMsgs", mock.Anything).
					Run(func(args mock.Arguments) {
						defer close(outCh)

						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
						outCh <- entity.MessageWithError{Msg: tc.expectedOut[0]}
					}).Return((<-chan entity.MessageWithError)(outCh))
			},
			expectedOut: []entity.Message{
				entity.NewMsg(
					"0", "i4u", "kek", false,
				),
			},
		},
		{
			name: "some messages failed",
			pre: func(t *testing.T, c api.Mail, tc fetcherJobTestcase) {
				outCh := make(chan entity.MessageWithError, 200)

				c.(*mocks.Mail).On("GetUnreadMsgs", mock.Anything).
					Run(func(args mock.Arguments) {
						defer close(outCh)

						ctx := args.Get(0).(context.Context)
						<-ctx.Done()
						for i, msg := range tc.expectedOut {
							if i%2 == 0 {
								outCh <- entity.MessageWithError{Msg: msg}
							} else {
								err := fmt.Errorf("something went wrong %s", msg.ID())
								outCh <- entity.MessageWithError{Err: err}
							}
						}
					}).Return((<-chan entity.MessageWithError)(outCh))
			},
			expectedOut: func() []entity.Message {
				size := 100
				out := make([]entity.Message, size)
				for i := 0; i < size; i++ {
					out[i] = entity.NewMsg(
						fmt.Sprintf("%d", i), "i4u", fmt.Sprintf("kek%d", i), false,
					)
				}

				return out
			}(),
			expectedErrors: func() map[string]bool {
				size := 100
				out := make(map[string]bool)
				for i := 0; i < size; i++ {
					if i%2 == 0 {
						continue
					}

					out[fmt.Errorf("failed to fetch message: %w",
						fmt.Errorf("something went wrong %d", i),
					).Error()] = false
				}

				return out
			}(),
		},
	}

	for _, tt := range testCases {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fetcher := mocks.NewMail(t)
			tc.pre(t, fetcher, tc)

			var (
				errsCh = make(chan error)
				out    = make([]chan entity.Message, tc.outputChSz)
				outW   = make([]chan<- entity.Message, tc.outputChSz)

				jobWg                 syncs.WaitGroup
				jobContext, jobCancel = context.WithCancel(context.Background())

				outputWg                    syncs.WaitGroup
				verifyContext, cancelVerify = context.WithCancel(context.Background())
			)

			for i := 0; i < tc.outputChSz; i++ {
				out[i] = make(chan entity.Message, tc.outputChSz)
				outW[i] = out[i]
			}

			jobWg.Go(func() {
				defer func() {
					for _, ch := range out {
						close(ch)
					}
				}()
				defer close(errsCh)

				NewFetcherJob(fetcher, 300*time.Millisecond, errsCh, outW).Run(jobContext)
			})

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

						if _, have := tc.expectedErrors[err.Error()]; !have {
							assert.Fail(t, "unexpected error", err)
						}

						if tc.expectedErrors[err.Error()] {
							assert.Error(t, err, "error can only be appeared once")
						}
					case <-verifyContext.Done():
						return
					}
				}
			}()

			time.Sleep(500 * time.Millisecond)
			jobCancel()
			jobWg.Wait()
			cancelVerify()
			outputWg.Wait()
		})
	}
}
