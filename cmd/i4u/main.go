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

	var lg, _ = zap.NewProduction()
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

	cmd := commands.Init(gmailConfig, gptConfig)
	if e := cmd.Execute(); e != nil {
		zap.L().Fatal("failed to execute command", zap.Error(e))
	}
}
