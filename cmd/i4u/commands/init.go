package commands

import (
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
)

func Init(
	gmailConfig *config.Gmail,
	gptConfig *config.GPT,
	tgConfig *config.Telegram,
	appConfig *config.AppConfig,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "i4u",
		Short: "i4u is your personal assistant to manage your internship requests",
		Long: `
i4u helps you to manage your internship requests. It will read your
emails and try to understand is this message is an internship request
or not.

If current unread message is an internship request, that summary
will be sent to the Telegram channel.

When message is processed, it will get an label "i4u"
to avoid processing it again.
`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: appConfig.Version,
	}

	if appConfig.IsDev() {
		rootCmd.AddCommand(devOpenAI(gptConfig))
		rootCmd.AddCommand(devTg(tgConfig))
	}

	rootCmd.AddCommand(authorize(gmailConfig))
	rootCmd.AddCommand(run(gmailConfig, gptConfig, tgConfig, appConfig))
	rootCmd.AddCommand(setup(gmailConfig))
	return rootCmd
}
