package commands

import (
	"github.com/fadyat/i4u/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
)

func tgChannelDescription(tgConfig *config.Telegram) *cobra.Command {
	return &cobra.Command{
		Use:   "tg",
		Short: "Telegram description for i4u",
		Long: `
This command will print the message to the telegram chat,
that tells the user what this bot is about.
`,
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, _ []string) {
			bot, err := tgbotapi.NewBotAPI(tgConfig.Token)
			if err != nil {
				log.Fatal(err)
			}

			_, err = bot.Send(tgbotapi.NewMessage(tgConfig.ChatID, `
Hello, I'm i4u bot. 😎 I'm here to help you find an internship. 🤝
I will read your emails 📧, and if I find something interesting, 🕵️
I will send you a message. 🚀`))
			if err != nil {
				zap.L().Fatal("failed to send message", zap.Error(err))
			}

			zap.L().Info("message sent")
		},
	}
}
