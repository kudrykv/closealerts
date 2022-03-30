package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	bot          clients.Telegram
	alerts       services.Alerts
	log          *zap.SugaredLogger
	notification services.Notification
}

func NewUpdate(
	log *zap.SugaredLogger,
	bot clients.Telegram,
	alerts services.Alerts,
	notification services.Notification,
) UpdateHandler {
	return UpdateHandler{
		log:          log,
		bot:          bot,
		alerts:       alerts,
		notification: notification,
	}
}

func (h UpdateHandler) Handle(ctx context.Context, update types.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	h.log.Infow("msg", "user", msg.Chat.UserName, "text", msg.Text)

	if !msg.IsCommand() {
		return
	}

	switch msg.Command() {
	case "look":
		if err := h.notification.Track(ctx, msg.Chat.ID, msg.CommandArguments()); err != nil {
			h.log.Errorw("notification track", "err", err)

			if _, err := h.bot.Client.Send(tgbotapi.NewMessage(msg.Chat.ID, "в мене щось пішло не так, спробуй ще раз")); err != nil {
				h.log.Errorw("send new message", "err", err)

				return
			}
		}

		if _, err := h.bot.Client.Send(tgbotapi.NewMessage(msg.Chat.ID, "буду пильнувати за "+msg.CommandArguments())); err != nil {
			h.log.Errorw("send new message", "err", err)

			return
		}

		return
	}

	if msg.Text == "шопачьом" {
		alerts, err := h.alerts.GetActiveFromRemote(ctx)
		if err != nil {
			h.log.Errorw("get active alerts from remote", "err", err)

			return
		}

		if len(alerts) == 0 {
			if _, err := h.bot.Client.Send(tgbotapi.NewMessage(msg.Chat.ID, "всьо норм")); err != nil {
				h.log.Errorw("send new message", "err", err)

				return
			}

			return
		}

		areas := make([]string, 0, len(alerts))
		for _, alert := range alerts {
			areas = append(areas, alert.ID)
		}

		if _, err := h.bot.Client.Send(tgbotapi.NewMessage(msg.Chat.ID, strings.Join(areas, ", "))); err != nil {
			h.log.Errorw("send new message", "err", err)

			return
		}
	}
}
