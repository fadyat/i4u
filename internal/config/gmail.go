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

	// LabelsLst is a list of labels that will be created in your Gmail account,
	// used for marking processed messages to avoid processing them again.
	LabelsLst []string `env:"GMAIL_LABELS" env-default:"i4u,intern:true,intern:false"`

	// L is a labels parsed after setup from yaml config file.
	L *LabelsMapper `yaml:"labels"`

	// MessagesLimit is a batch size for fetching messages from Gmail.
	MessagesLimit int64 `env:"GMAIL_MESSAGES_LIMIT" env-description:"Batch size for fetching messages" env-default:"2"`
}

func NewGmail() (*Gmail, error) {
	var c Gmail
	if err := cleanenv.ReadEnv(&c); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadConfig(".i4u/config.yaml", &c); err != nil {
		return nil, err
	}

	return &c, nil
}
