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
	chat         services.Chats
}

func NewUpdate(
	log *zap.SugaredLogger,
	bot clients.Telegram,
	alerts services.Alerts,
	notification services.Notification,
	chat services.Chats,
) UpdateHandler {
	return UpdateHandler{
		log:          log,
		bot:          bot,
		alerts:       alerts,
		chat:         chat,
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
	text := strings.TrimSpace(msg.Text)

	// load "user" from db
	// if not a command:
	//   switch on the command saved to user
	//   apply text to the command
	// else
	//   store command on the user

	chat, err := r.chat.FirstOrCreate(ctx, chatID)
	if err != nil {
		r.log.Errorw("load or create chat", "err", err)

		return
	}

	if msg.IsCommand() {
		switch msg.Command() {
		case "track":
			if err := r.chat.SetCommand(ctx, chatID, "track"); err != nil {
				r.log.Errorw("set command", "err", err)
				r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

				return
			}

			r.bot.MaybeSendText(ctx, chatID, "вкажи територію, за якою пильнувати")

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

		case "stop":
			if err := r.chat.SetCommand(ctx, chatID, "stop"); err != nil {
				r.log.Errorw("set command", "err", err)
				r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

				return
			}

			r.bot.MaybeSendText(ctx, chatID, "вкажи територію від якої відписатись")

		case "alerts":
			alerts, err := r.alerts.GetActive(ctx)
			if err != nil {
				r.log.Errorw("notification track", "err", err)
				r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

				return
			}

			if len(alerts) == 0 {
				r.bot.MaybeSendText(ctx, chatID, "все тихо")

				return
			}

			areas := make([]string, 0, len(alerts))
			for _, alert := range alerts {
				areas = append(areas, alert.ID)
			}

			r.bot.MaybeSendText(ctx, chatID, strings.Join(areas, ", "))

		default:
			r.bot.MaybeSendText(ctx, chatID, "я такої команди не знаю")
		}

		return
	} else {
		switch chat.Command {
		case "":
			r.bot.MaybeSendText(ctx, chatID, "а, шо?")

		case "track":
			if err := r.notification.Track(ctx, chatID, text); err != nil {
				r.log.Errorw("notification track", "err", err)
				r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

				return
			}

			r.bot.MaybeSendText(ctx, chatID, "буду пильнувати за "+text)

			if err := r.chat.ClearCommand(ctx, chatID); err != nil {
				r.log.Errorw("clear command", "err", err)

				return
			}

		case "stop":
			if err := r.notification.Stop(ctx, chatID, text); err != nil {
				r.log.Errorw("notification track", "err", err)
				r.bot.MaybeSendText(ctx, chatID, "в мене щось пішло не так, спробуй ще раз")

				return
			}

			r.bot.MaybeSendText(ctx, chatID, "відписуюсь від "+text)
		}
	}
}
