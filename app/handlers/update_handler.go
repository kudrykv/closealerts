package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type UpdateHandler struct {
	bot          clients.Telegram
	alerts       services.Alerts
	log          *zap.SugaredLogger
	notification services.Notification
	chat         services.Chats
	commander    services.Commander
}

func NewUpdate(
	log *zap.SugaredLogger,
	bot clients.Telegram,
	alerts services.Alerts,
	notification services.Notification,
	chat services.Chats,
	commander services.Commander,
) UpdateHandler {
	return UpdateHandler{
		log:          log,
		bot:          bot,
		alerts:       alerts,
		chat:         chat,
		notification: notification,
		commander:    commander,
	}
}

func (r UpdateHandler) Handle(ctx context.Context, update types.Update) {
	msg := update.Message
	cq := update.CallbackQuery

	switch {
	case msg != nil:
		r.handleMessage(ctx, msg)

	case cq != nil:
		r.handleCallbackQuery(ctx, cq)
	}
}

func (r UpdateHandler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	r.log.Infow("msg", "user", msg.Chat.UserName, "text", msg.Text)

	chat, err := r.chat.FirstOrCreate(ctx, msg.Chat)
	if err != nil {
		r.log.Errorw("load or create chat", "err", err)

		return
	}

	command := chat.Command
	args := msg.Text
	clearCmd := true

	if msg.IsCommand() {
		command = msg.Command()
		args = msg.CommandArguments()
		clearCmd = false
	}

	var chattable tgbotapi.Chattable

	switch command {
	case "start":
		chattable, err = r.commander.Start(ctx, msg, args)

	case "areas":
		chattable, err = r.commander.Areas(ctx, msg, args)

	case "alerts":
		chattable, err = r.commander.Alerts(ctx, msg, args)

	case "track":
		chattable, err = r.commander.Track(ctx, msg, args)

	case "tracking":
		chattable, err = r.commander.Tracking(ctx, msg, args)

	case "stop":
		chattable, err = r.commander.Stop(ctx, msg, args)

	case "auth":
		chattable, err = r.commander.Auth(ctx, msg, args)

	case "admin_fake_alert_in":
		if !chat.PrivSendFakeEvent {
			chattable = tgbotapi.NewMessage(chat.ID, "Please auth first")

			break
		}

		chattable, err = r.commander.AdminFakeAlertIn(ctx, msg, args)

	default:
		chattable = tgbotapi.NewMessage(chat.ID, "я такої команди не знаю")
	}

	if err != nil {
		r.log.Errorw(command, "err", err)
		r.bot.MaybeSendText(ctx, chat.ID, "в мене щось пішло не так, спробуй ще раз")
	} else {
		r.bot.MaybeSend(ctx, chattable)
	}

	if clearCmd {
		if err := r.chat.ClearCommand(ctx, chat.ID); err != nil {
			r.log.Errorw("clear command", "err", err)

			return
		}
	}
}

func (r UpdateHandler) handleCallbackQuery(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	msg := cq.Message
	chat := msg.Chat
	split := strings.SplitN(cq.Data, ":", 2)

	if len(split) != 2 {
		r.log.Errorw(
			"bad data",
			"data", cq.Data,
			"msg_id", msg.MessageID,
			"chat_id", chat.ID,
		)

		return
	}

	action, payload := split[0], split[1]
	var (
		chattable tgbotapi.Chattable
		err       error
	)

	switch action {
	case "toggle_area":
		chattable, err = r.commander.ToggleArea(ctx, cq, payload)
	default:
		err = fmt.Errorf("%s: %w", action, types.ErrUnknownCBAction)
	}

	if err != nil {
		r.log.Errorw(action, "err", err)
	} else {
		r.bot.MaybeSend(ctx, chattable)
	}

	return
}
