package commands

import (
	"context"
	"errors"
	"github.com/fadyat/i4u/api/mail"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/job"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func run(config *config.Gmail) *cobra.Command {
	var oauth2Config = token.GetOAuthConfig(config)
	var staticToken, err = token.ReadTokenFromFile(config.TokenFile)
	if err != nil {
		log.Fatal("unauthorized, run `i4u init` first")
	}

	return &cobra.Command{
		Use:   "run",
		Args:  cobra.NoArgs,
		Short: "Starting web server to read your Gmail inbox",
		Long: `
Start point of the application. It will start a web server
on port 8080 by default. It will read your Gmail inbox and
trying to understand is this message is an internship request
or not.

If current unread message is an internship request, it will
be added to the tracker queue for further processing.

When message is processed, it will get an label "i4u-processed"
to avoid processing it again.
`,
		Run: func(cmd *cobra.Command, _ []string) {
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			defer close(signalChan)

			var wg sync.WaitGroup
			wg.Add(1)
			mailContext, cancel := context.WithCancel(context.Background())
			go func() {
				defer wg.Done()

				mailJob := job.NewMailJob(mail.NewGmailClient(staticToken, oauth2Config))
				if e := mailJob.Run(mailContext); e != nil && !errors.Is(e, context.Canceled) {
					zap.L().Fatal("failed to run mail job", zap.Error(e))
				}
			}()

			select {
			case <-signalChan:
				zap.L().Info("received signal, exiting")
				cancel()
			}

			wg.Wait()
		},
	}
}
