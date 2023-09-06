package config

import "github.com/ilyakaznacheev/cleanenv"

type AppConfig struct {
	Keywords []string `env:"APP_ANALYZER_KEYWORDS" env-default:"internship,opportunity,training,intern"`
	Version  string   `env:"APP_VERSION" env-default:"development"`
}

func (a *AppConfig) IsDev() bool {
	return a.Version == "development"
}

func NewAppConfig() (*AppConfig, error) {
	var appConfig AppConfig
	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		return nil, err
	}

	return &appConfig, nil
}
