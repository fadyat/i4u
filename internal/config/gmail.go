package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Gmail struct {

	// CredentialsFile is a path to your credentials file for performing OAuth2
	// authentication. You can get it from Google Cloud Console.
	//
	// https://console.cloud.google.com/apis/credentials
	//
	CredentialsFile string `env:"GMAIL_CREDENTIALS_FILE" env-description:"Path to your credentials file" env-default:"credentials.json"`

	// TokenFile is a path to your token file for performing OAuth2
	// authentication. Will be used to store token after authentication and
	// refreshing access token automatically, when it expires.
	TokenFile string `env:"GMAIL_TOKEN_FILE" env-description:"Path to your token file" env-default:"token.json"`
}

func NewGmail() (*Gmail, error) {
	var c Gmail
	if err := cleanenv.ReadEnv(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
