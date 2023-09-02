package commands

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api/mail"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
	"log"
	"sync"
)

func setupLabels(cfg *config.Gmail) *cobra.Command {
	var (

		// todo: get from config?
		labels = []string{
			"i4u",
			"intern:true",
			"intern:false",
		}
		oauth2Config = token.GetOAuthConfig(cfg)
	)

	return &cobra.Command{
		Use:   "setup",
		Args:  cobra.NoArgs,
		Short: "Setup labels in Gmail",
		Long: fmt.Sprintf(`
This command will setup labels in your Gmail account. It will create the following
labels if they don't exist:
%s`, labels),
		Run: func(cmd *cobra.Command, _ []string) {
			var staticToken, err = token.ReadTokenFromFile(cfg.TokenFile)
			if err != nil {
				log.Fatal("unauthorized, run `i4u init` first")
			}

			var (
				wg     sync.WaitGroup
				errsCh = make(chan error)
			)
			gmailClient := mail.NewGmailClient(staticToken, oauth2Config)
			for _, label := range labels {
				wg.Add(1)

				go func(l string) {
					defer wg.Done()

					lbl, e := gmailClient.CreateLabel(context.Background(), l)
					if e != nil {
						errsCh <- e
						return
					}

					log.Printf("created label: %s: %s", lbl.Name, lbl.Id)
				}(label)
			}

			go func() {
				wg.Wait()
				close(errsCh)
			}()

			for e := range errsCh {
				log.Printf("failed to create label: %v", e)
			}
		},
	}
}
