package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"strings"

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

func (r UpdateHandler) Handle(ctx context.Context, update types.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	r.log.Infow("msg", "user", msg.Chat.UserName, "text", msg.Text)

	chatID := msg.Chat.ID

	if !msg.IsCommand() {
		r.bot.MaybeSendText(ctx, chatID, "Невідома для мене дія")

		return
	}

	switch msg.Command() {
	case "tracking":
		list, err := r.notification.Tracking(ctx, chatID)
		if err != nil {
			r.log.Errorw("notification track", "err", err)
			r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

			return
		}

		if len(list) == 0 {
			r.bot.MaybeSendText(ctx, chatID, "ще нічого не трекаєш")

			return
		}

		areas := make([]string, 0, len(list))
		for _, notification := range list {
			areas = append(areas, notification.Area)
		}

		r.bot.MaybeSendText(ctx, chatID, strings.Join(areas, ", "))

	case "track":
		if err := r.notification.Track(ctx, chatID, msg.CommandArguments()); err != nil {
			r.log.Errorw("notification track", "err", err)
			r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

			return
		}

		r.bot.MaybeSendText(ctx, chatID, "буду пильнувати за "+msg.CommandArguments())

	default:
		r.bot.MaybeSendText(ctx, chatID, "Невідома для мене дія")
	}
}
