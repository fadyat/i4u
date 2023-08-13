package api

import (
	"context"
)

type Mail interface {
	GetUnreadEmails(context.Context)
}
