package commands

import (
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
)

func Init(cfg *config.Gmail) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "i4u",
		Short: "i4u is a command line tool for reading your Gmail inbox",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: "v0.0.1",
	}

	rootCmd.AddCommand(authorize(cfg))
	rootCmd.AddCommand(run(cfg))
	rootCmd.AddCommand(status(cfg))
	return rootCmd
}
