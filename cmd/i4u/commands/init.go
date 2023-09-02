package commands

import (
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
)

func Init(
	gmailConfig *config.Gmail,
	gptConfig *config.GPT,
	tgConfig *config.Telegram,
) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "i4u",
		Short: "i4u is a command line tool for reading your Gmail inbox",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: "v0.0.1",
	}

	rootCmd.AddCommand(authorize(gmailConfig))
	rootCmd.AddCommand(run(gmailConfig, gptConfig, tgConfig))
	rootCmd.AddCommand(status(gmailConfig))
	rootCmd.AddCommand(analyze(gptConfig))
	rootCmd.AddCommand(setupLabels(gmailConfig))
	rootCmd.AddCommand(tgChannelDescription(tgConfig))
	return rootCmd
}
