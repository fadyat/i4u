package commands

import (
	"fmt"
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

type statusMsg struct {
	authorized bool
	tokenPath  string
}

func (s *statusMsg) String() string {
	emotion := "😃"
	if !s.authorized {
		emotion = "😢"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Status: %-6s\n", emotion))
	sb.WriteString(fmt.Sprintf("Token Path: %-20s\n", s.tokenPath))
	return sb.String()
}

func status(cfg *config.Gmail) *cobra.Command {

	return &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Short: "Show status of i4u",
		Run: func(cmd *cobra.Command, _ []string) {
			if _, err := os.Stat(cfg.TokenFile); os.IsNotExist(err) {
				log.Print(&statusMsg{
					authorized: false,
					tokenPath:  cfg.TokenFile,
				})
			}

			log.Print(&statusMsg{
				authorized: true,
				tokenPath:  cfg.TokenFile,
			})
		},
	}
}
