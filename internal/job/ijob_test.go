package job

import (
	"fmt"
	"github.com/fadyat/i4u/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strconv"
	"testing"
)

func setup() {
	lg, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(lg)

	config.FeatureFlags = config.Flags{
		IsAnalyzerJobEnabled:   true,
		IsSenderJobEnabled:     true,
		IsLabelerJobEnabled:    true,
		IsSummarizerJobEnabled: true,
	}
}

func parseInt(t *testing.T, s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("failed to parse int: %s", err))
	}

	return i
}

func TestMain(m *testing.M) {
	setup()

	m.Run()
}
