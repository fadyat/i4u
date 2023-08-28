package entity

import (
	"fmt"
	"github.com/fadyat/i4u/pkg/parser"
	"google.golang.org/api/gmail/v1"
)

type Msg struct {
	Subject string
	Body    string
}

func (m *Msg) String() string {
	return fmt.Sprintf("Subject:\n%s\n\nBody: %s", m.Subject, m.Body)
}

func NewMsgFromGmailMessage(msg *gmail.Message) (*Msg, error) {
	content, err := parser.CleanMsg(msg, parser.PlainText)
	if err != nil {
		return nil, err
	}

	return &Msg{
		Subject: "example",
		Body:    content,
	}, nil
}
