package commands

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api/mail"
	setupConfig "github.com/fadyat/i4u/cmd/i4u/config"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"sync"
)

func setupLabels(gmailConfig *config.Gmail) *cobra.Command {
	var oauth2Config = token.GetOAuthConfig(gmailConfig)

	return &cobra.Command{
		Use:   "setup",
		Args:  cobra.NoArgs,
		Short: "Setup labels in Gmail",
		Long: fmt.Sprintf(`
This command will setup labels in your Gmail account. It will create the following
labels if they don't exist:
%s`, gmailConfig.LabelsLst),
		Run: func(cmd *cobra.Command, _ []string) {
			var staticToken, err = token.ReadTokenFromFile(gmailConfig.TokenFile)
			if err != nil {
				log.Fatal("unauthorized, run `i4u init` first")
			}

			var mu sync.Mutex
			withLock := func(f func()) {
				mu.Lock()
				defer mu.Unlock()
				f()
			}

			var (
				wg     sync.WaitGroup
				errsCh = make(chan error)
				labels = make(map[string]string)
			)

			gmailClient := mail.NewGmailClient(staticToken, oauth2Config, gmailConfig)
			for _, label := range gmailConfig.LabelsLst {
				wg.Add(1)

				go func(l string) {
					defer wg.Done()

					lbl, e := gmailClient.CreateLabel(context.Background(), l)
					if e != nil {
						errsCh <- e
						return
					}

					withLock(func() { labels[l] = lbl.ID })
					zap.S().Infof("created label: %s", l)
				}(label)
			}

			go func() {
				wg.Wait()
				close(errsCh)
			}()

			for e := range errsCh {
				log.Printf("failed to create label: %v", e)
			}

			if e := setupConfig.SaveLabelsToYaml(".i4u/config.yaml", labels); e != nil {
				log.Printf("failed to save labels to config file: %v", e)
			}
		},
	}
}
