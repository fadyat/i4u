package entity

import "google.golang.org/api/gmail/v1"

type Label struct {
	ID   string
	Name string
}

func NewLabelFromGmailLabel(l *gmail.Label) *Label {
	return &Label{
		ID:   l.Id,
		Name: l.Name,
	}
}
