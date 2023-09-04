package main

import (
	"github.com/fadyat/i4u/cmd/i4u/commands"
	"github.com/fadyat/i4u/internal/config"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
)

func init() {
	log.SetFlags(0)

	var lg, _ = zap.NewDevelopment()
	zap.ReplaceGlobals(lg)

	_ = godotenv.Load(".env")
}

func main() {
	defer func() { _ = zap.L().Sync() }()

	gmailConfig, err := config.NewGmail()
	if err != nil {
		zap.L().Fatal("failed to initialize gmail config", zap.Error(err))
	}

	gptConfig, err := config.NewGPT()
	if err != nil {
		zap.L().Fatal("failed to initialize gpt config", zap.Error(err))
	}

	tgConfig, err := config.NewTelegram()
	if err != nil {
		zap.L().Fatal("failed to initialize telegram config", zap.Error(err))
	}

	if err = config.NewFeatureFlags(); err != nil {
		zap.L().Fatal("failed to initialize feature flags", zap.Error(err))
	}

	appConfig, err := config.NewAppConfig()
	if err != nil {
		zap.L().Fatal("failed to initialize app config", zap.Error(err))
	}

	cmd := commands.Init(gmailConfig, gptConfig, tgConfig, appConfig)
	if e := cmd.Execute(); e != nil {
		zap.L().Fatal("failed to execute command", zap.Error(e))
	}
}
