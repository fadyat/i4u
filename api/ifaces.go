package api

import (
	"context"
	"github.com/fadyat/i4u/internal/entity"
)

type Mail interface {
	GetUnreadMsgs(ctx context.Context) <-chan entity.MessageWithError
	LabelMsg(context.Context, entity.MessageForLabeler) error
	CreateLabel(context.Context, string) (*entity.Label, error)
}

type Analyzer interface {
	IsInternshipRequest(context.Context, entity.Message) (bool, error)
}

type Summarizer interface {
	GetMsgSummary(context.Context, entity.Message) (string, error)
}

type Sender interface {
	Send(context.Context, entity.SummaryMessage) error
}
