package commands

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/api/analyzer"
	"github.com/fadyat/i4u/api/mail"
	"github.com/fadyat/i4u/api/sender"
	"github.com/fadyat/i4u/api/summary"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/internal/job"
	"github.com/fadyat/i4u/pkg/syncs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func run(
	gmailConfig *config.Gmail,
	gptConfig *config.GPT,
	tgConfig *config.Telegram,
	appConfig *config.AppConfig,
) *cobra.Command {
	var oauth2Config = token.GetOAuthConfig(gmailConfig)

	return &cobra.Command{
		Use:   "run",
		Args:  cobra.NoArgs,
		Short: "Entrypoint for the application",
		Long: `
Starting all jobs of application. Each job launches in a separate goroutine.
Context is used to stop all jobs when signal is received.

All messages started for processing will go through all stages of the pipeline.
`,
		Run: func(cmd *cobra.Command, _ []string) {
			var staticToken, err = token.ReadTokenFromFile(gmailConfig.TokenFile)
			if err != nil {
				log.Fatal("unauthorized, run `i4u init` first")
			}

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			defer close(signalChan)

			var alertsNotifier api.Sender = sender.NewTg(
				func() *tgbotapi.BotAPI {
					bot, e := tgbotapi.NewBotAPI(tgConfig.Token)
					if e != nil {
						log.Fatal(e)
					}

					return bot
				}(),
				tgConfig.AlertsChatID,
			)

			producer := job.NewProducer(
				mail.NewGmailClient(staticToken, oauth2Config, gmailConfig),
				analyzer.NewKWAnalyzer(appConfig.Keywords),
				summary.NewOpenAI(openai.NewClient(gptConfig.OpenAIKey), gptConfig),
				sender.NewTg(
					func() *tgbotapi.BotAPI {
						bot, e := tgbotapi.NewBotAPI(tgConfig.Token)
						if e != nil {
							log.Fatal(e)
						}

						return bot
					}(),
					tgConfig.ChatID,
				),
				gmailConfig.L,
			)

			ctx, cancel := context.WithCancel(context.Background())

			var wg syncs.WaitGroup
			wg.Go(func() {
				for e := range producer.Produce(ctx) {
					zap.L().Error("got error during processing", zap.Error(e))

					if er := alertsNotifier.Send(ctx, entity.NewAlertMsg(e)); er != nil {
						zap.L().Error("failed to send alert", zap.Error(er))
					}
				}
			})

			<-signalChan
			zap.L().Info("received signal, exiting")
			cancel()

			wg.Wait()
			zap.L().Info("exiting")
		},
	}
}
