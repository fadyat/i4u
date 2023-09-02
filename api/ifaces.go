package api

import (
	"context"
	"github.com/fadyat/i4u/internal/entity"
	"google.golang.org/api/gmail/v1"
)

type Mail interface {
	GetUnreadMsgs(context.Context) ([]*gmail.Message, error) // todo: make an interface for gmail message
	LabelMsg(context.Context, entity.MessageForLabeler) error
	CreateLabel(context.Context, string) (*gmail.Label, error) // todo: make an interface for gmail label
}

type Analyzer interface {
	IsInternshipRequest(context.Context, entity.Message) (bool, error)
}

type Summarizer interface {
	GetMsgSummary(context.Context, entity.Message) (string, error)
}

type Sender interface {
	Send(context.Context, *entity.SummaryMsg) error
}
