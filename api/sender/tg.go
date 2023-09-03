package sender

import (
	"context"
	"github.com/fadyat/i4u/internal/entity"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tg struct {
	c      *tgbotapi.BotAPI
	chatID int64
}

func NewTg(c *tgbotapi.BotAPI, chatID int64) *Tg {
	return &Tg{c: c, chatID: chatID}
}

func (t *Tg) Send(_ context.Context, msg entity.SummaryMessage) error {
	_, err := t.c.Send(tgbotapi.NewMessage(t.chatID, msg.Summary()))
	return err
}
