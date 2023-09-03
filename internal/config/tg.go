package config

import "github.com/ilyakaznacheev/cleanenv"

type Telegram struct {
	ChatID       int64  `env:"TG_CHAT_ID" env-description:"Telegram chat ID" env-required:"true"`
	AlertsChatID int64  `env:"TG_ALERTS_CHAT_ID" env-description:"Telegram alerts chat ID" env-required:"true"`
	Token        string `env:"TG_TOKEN" env-description:"Telegram bot token" env-required:"true"`
}

func NewTelegram() (*Telegram, error) {
	var c Telegram
	if err := cleanenv.ReadEnv(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
