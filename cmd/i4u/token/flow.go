package token

import (
	"encoding/json"
	"github.com/fadyat/i4u/internal/config"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"os"
)

func GetOAuthConfig(cfg *config.Gmail) *oauth2.Config {
	clientCreds, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		zap.L().Fatal("failed to read credentials file", zap.Error(err))
	}

	oauth2Config, err := google.ConfigFromJSON(clientCreds, gmail.GmailReadonlyScope)
	if err != nil {
		zap.L().Fatal("failed to parse credentials", zap.Error(err))
	}

	return oauth2Config
}

func ReadTokenFromFile(tokenFile string) (*oauth2.Token, error) {
	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := f.Close(); e != nil {
			zap.L().Warn("failed to close token file", zap.Error(e))
		}
	}()

	var token oauth2.Token
	if e := json.NewDecoder(f).Decode(&token); e != nil {
		return nil, e
	}

	return &token, nil
}

func Save(tokenFile string, token *oauth2.Token) error {
	f, err := os.Create(tokenFile)
	if err != nil {
		return err
	}

	defer func() {
		if e := f.Close(); e != nil {
			zap.L().Warn("failed to close token file", zap.Error(e))
		}
	}()

	return json.NewEncoder(f).Encode(token)
}
