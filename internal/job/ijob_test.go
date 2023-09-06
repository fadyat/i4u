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
	lg := zap.NewNop()
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

func newLabelsMapper() *config.LabelsMapper {
	return &config.LabelsMapper{
		I4U:       "i4u",
		IsIntern:  "is_intern",
		NotIntern: "not_intern",
	}
}

func TestMain(m *testing.M) {
	setup()

	m.Run()
}
