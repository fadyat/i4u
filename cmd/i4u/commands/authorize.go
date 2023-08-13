package commands

import (
	"context"
	"errors"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func authorize(gmailConfig *config.Gmail) *cobra.Command {
	var oauth2Config = token.GetOAuthConfig(gmailConfig)

	return &cobra.Command{
		Use:   "init",
		Args:  cobra.NoArgs,
		Short: "Grant access to your Gmail account to i4u",
		Long: `
This command will guide you through the process of granting access to your Gmail
account to i4u. It will open a browser window and ask you to login to your
Google account and grant access to i4u. After that, it will save the token
on your local machine and you will be able to use i4u without having to
authenticate again.`,
		Run: func(cmd *cobra.Command, _ []string) {
			done := make(chan bool)
			defer close(done)

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
			defer close(signalChan)

			authURL := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
			go func() {
				if err := openBrowser(authURL); err != nil {
					zap.L().Fatal("failed to open browser", zap.Error(err))
				}
			}()

			localServer := &http.Server{Addr: ":80"}
			go func() {
				callback, err := url.Parse(oauth2Config.RedirectURL)
				if err != nil {
					zap.L().Fatal("failed to parse redirect URL", zap.Error(err))
				}

				http.HandleFunc(callback.Path, func(w http.ResponseWriter, r *http.Request) {
					tok, e := oauth2Config.Exchange(context.Background(), r.URL.Query().Get("code"))
					if e != nil {
						zap.L().Info("Unable to retrieve token from web", zap.Error(e))
						_, _ = w.Write([]byte("Unable to retrieve token from web"))
						return
					}

					if e = token.Save(gmailConfig.TokenFile, tok); e != nil {
						zap.L().Info("failed to save token", zap.Error(e))
						_, _ = w.Write([]byte("failed to save token"))
						return
					}

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("Authorization successful. You can close this window now."))
					done <- true
				})

				if e := localServer.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
					zap.L().Fatal("failed to start local server", zap.Error(e))
					signalChan <- syscall.SIGTERM
				}
			}()

			select {
			case <-signalChan:
				zap.L().Info("received signal, shutting down local server")
				if err := localServer.Shutdown(context.Background()); err != nil {
					zap.L().Fatal("failed to shutdown local server", zap.Error(err))
				}
			case <-done:
				zap.L().Info("configuration was successful, shutting down local server")
				if err := localServer.Shutdown(context.Background()); err != nil {
					zap.L().Fatal("failed to shutdown local server", zap.Error(err))
				}
			}
		},
	}
}
