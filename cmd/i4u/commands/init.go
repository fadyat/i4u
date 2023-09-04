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
		Short: "i4u is a command line tool for reading your Gmail inbox",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: appConfig.Version,
	}

	rootCmd.AddCommand(authorize(gmailConfig))
	rootCmd.AddCommand(run(gmailConfig, gptConfig, tgConfig, appConfig))
	rootCmd.AddCommand(analyze(gptConfig))
	rootCmd.AddCommand(setupLabels(gmailConfig))
	rootCmd.AddCommand(tgChannelDescription(tgConfig))
	return rootCmd
}
