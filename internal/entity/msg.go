package entity

import (
	"github.com/fadyat/i4u/pkg/parser"
	"google.golang.org/api/gmail/v1"
)

type MessageForLabeler interface {
	ID() string
	Label() string
}

type MessageWithError struct {
	Msg Message
	Err error
}

type Message interface {
	Body() string
	IsInternshipRequest() bool
	Link() string

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

	// link is a fast-access link to the message.
	//
	// For gmail, it's done via `https://mail.google.com/mail/u/0/#inbox/` + id.
	link string
}

func (m *Msg) Body() string {
	return m.body
}

func (m *Msg) IsInternshipRequest() bool {
	return m.isInternshipRequest
}

func (m *Msg) Link() string {
	return m.link
}

func (m *Msg) Copy() *Msg {
	return &Msg{
		id:                  m.id,
		body:                m.body,
		isInternshipRequest: m.isInternshipRequest,
		label:               m.label,
		link:                m.link,
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
		link:                "https://mail.google.com/mail/u/0/#inbox/" + msg.Id,
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
	return m.label
}
