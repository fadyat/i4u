package sender

import (
	"context"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tg struct {
	c   *tgbotapi.BotAPI
	cfg *config.Telegram
}

func NewTg(c *tgbotapi.BotAPI, cfg *config.Telegram) *Tg {
	return &Tg{c: c, cfg: cfg}
}

func (t *Tg) Send(_ context.Context, msg *entity.SummaryMsg) error {
	_, err := t.c.Send(tgbotapi.NewMessage(t.cfg.ChatID, msg.Summary))
	return err
}
