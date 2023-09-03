package entity

import (
	"github.com/fadyat/i4u/pkg/parser"
	"google.golang.org/api/gmail/v1"
)

type MessageForLabeler interface {
	ID() string
	Label() string
}

type Message interface {
	Body() string
	IsInternshipRequest() bool

	MessageForLabeler
}

type Msg struct {

	// id is a unique identifier of the message, in case of gmail
	// it is a gmail message id.
	id string

	// body is a plain text of the message, without any html tags.
	// Usable for getting short description of the message.
	body string

	// isInternshipRequest is a flag that indicates
	// whether the message is related to internship request.
	//
	// By default, it is false, when an analyzer job done his work
	// it will set this flag to true if message is related to internship request.
	isInternshipRequest bool

	// label is a tag name for performing labeling in gmail.
	//
	// Be default, it is empty, and will be replaced with `i4u` tag.
	// For another cases, like, second labeling, you can provide custom value.
	label string
}

func (m *Msg) Body() string {
	return m.body
}

func (m *Msg) IsInternshipRequest() bool {
	return m.isInternshipRequest
}

func (m *Msg) Copy() *Msg {
	return &Msg{
		id:                  m.id,
		body:                m.body,
		isInternshipRequest: m.isInternshipRequest,
		label:               m.label,
	}
}

func (m *Msg) WithLabel(v string) *Msg {
	m.label = v
	return m
}

func (m *Msg) WithIsIntern(v bool) *Msg {
	m.isInternshipRequest = v
	return m
}

func NewMsgFromGmailMessage(msg *gmail.Message) (*Msg, error) {
	content, err := parser.CleanMsg(msg, parser.PlainText)
	if err != nil {
		return nil, err
	}

	return &Msg{
		id:                  msg.Id,
		body:                content,
		isInternshipRequest: false,
	}, nil
}

func NewMsg(
	id, label, body string,
	isInternshipRequest bool,
) *Msg {
	return &Msg{
		id:                  id,
		body:                body,
		isInternshipRequest: isInternshipRequest,
		label:               label,
	}
}

func (m *Msg) ID() string {
	return m.id
}

func (m *Msg) Label() string {
	if m.label == "" {
		// todo: get label name from config
		return "Label_12"
	}

	return m.label
}

type SummaryMsg struct {
	Message
	Summary string
}

func NewSummaryMsg(msg Message, summary string) *SummaryMsg {
	return &SummaryMsg{
		Message: msg,
		Summary: summary,
	}
}
