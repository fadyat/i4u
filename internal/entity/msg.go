package entity

import (
	"encoding/base64"
	"fmt"
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
	for _, part := range msg.Payload.Parts {
		if part.MimeType == "text/plain" {
			data, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				return nil, err
			}

			return &Msg{
				Subject: "sub",
				Body:    string(data),
			}, nil
		}
	}

	return &Msg{}, nil
}
