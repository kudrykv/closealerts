package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/types"
	"context"
)

type UpdateHandler struct {
	bot clients.Telegram
}

func NewUpdate(bot clients.Telegram) UpdateHandler {
	return UpdateHandler{bot: bot}
}

func (h UpdateHandler) Handle(_ context.Context, update types.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	if msg.Text == "шопачьом" {

	}
}
