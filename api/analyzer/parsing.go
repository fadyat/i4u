package analyzer

import (
	"context"
	"github.com/fadyat/i4u/internal/entity"
	"strings"
)

type KeywordsAnalyzer struct {
	keywords []string
}

func NewKWAnalyzer(keywords []string) *KeywordsAnalyzer {
	return &KeywordsAnalyzer{
		keywords: keywords,
	}
}

func (a *KeywordsAnalyzer) IsInternshipRequest(_ context.Context, msg entity.Message) (bool, error) {
	body := strings.ToLower(msg.Body())
	for _, keyword := range a.keywords {
		if strings.Contains(body, keyword) {
			return true, nil
		}
	}

	return false, nil
}
