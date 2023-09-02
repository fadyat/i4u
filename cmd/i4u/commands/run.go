package commands

import (
	"context"
	"github.com/fadyat/i4u/api/analyzer"
	"github.com/fadyat/i4u/api/mail"
	"github.com/fadyat/i4u/api/sender"
	"github.com/fadyat/i4u/api/summary"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/job"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func run(
	gmailConfig *config.Gmail,
	gptConfig *config.GPT,
	tgConfig *config.Telegram,
) *cobra.Command {
	var oauth2Config = token.GetOAuthConfig(gmailConfig)

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
			var staticToken, err = token.ReadTokenFromFile(gmailConfig.TokenFile)
			if err != nil {
				log.Fatal("unauthorized, run `i4u init` first")
			}

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			defer close(signalChan)

			producer := job.NewProducer(
				mail.NewGmailClient(staticToken, oauth2Config),
				// todo: get from config
				analyzer.NewKWAnalyzer([]string{
					"internship",
					"opportunity",
					"training",
					"intern",
				}),
				summary.NewOpenAI(
					openai.NewClient(gptConfig.OpenAIKey),
					gptConfig,
				),
				sender.NewTg(
					func() *tgbotapi.BotAPI {
						bot, e := tgbotapi.NewBotAPI(tgConfig.Token)
						if e != nil {
							log.Fatal(e)
						}

						return bot
					}(),
					tgConfig,
				),
			)

			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()

				errsCh := producer.Produce(ctx)
				for e := range errsCh {
					zap.L().Error("got error during processing", zap.Error(e))
				}
			}()

			<-signalChan
			zap.L().Info("received signal, exiting")
			cancel()

			wg.Wait()
		},
	}
}
