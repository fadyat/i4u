package config

import "github.com/ilyakaznacheev/cleanenv"

var (
	FeatureFlags Flags
)

type Flags struct {
	IsLabelerJobEnabled    bool `env:"LABELER_JOB_ENABLED" env-default:"true"`
	IsAnalyzerJobEnabled   bool `env:"ANALYZER_JOB_ENABLED" env-default:"true"`
	IsSummarizerJobEnabled bool `env:"SUMMARIZER_JOB_ENABLED" env-default:"true"`
	IsSenderJobEnabled     bool `env:"SENDER_JOB_ENABLED" env-default:"true"`
}

func NewFeatureFlags() error {
	return cleanenv.ReadEnv(&FeatureFlags)
}
