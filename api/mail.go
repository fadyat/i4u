package api

import (
	"context"
)

type Mail interface {
	GetUnreadMsgs(context.Context)
}
