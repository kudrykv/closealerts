package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"

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

	switch {
	case msg != nil:
		r.handleMessage(ctx, msg)
	}
}

func (r UpdateHandler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	r.log.Infow("msg", "user", msg.Chat.UserName, "text", msg.Text)

	// load "user" from db
	// if not a command:
	//   switch on the command saved to user
	//   apply text to the command
	// else
	//   store command on the user

	chat, err := r.chat.FirstOrCreate(ctx, msg.Chat.ID)
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
		chattable, err = r.commander.Start(ctx, chat, args)

	case "track":
		chattable, err = r.commander.Track(ctx, chat, args)

	case "tracking":
		chattable, err = r.commander.Tracking(ctx, chat, args)

	case "stop":
		chattable, err = r.commander.Stop(ctx, chat, args)

	case "alerts":
		chattable, err = r.commander.Alerts(ctx, chat, args)

	case "areas":
		chattable, err = r.commander.Areas(ctx, chat, args)

	default:
		chattable = tgbotapi.NewMessage(chat.ID, "я такої команди не знаю")
	}

	if err != nil {
		r.log.Errorw("track", "err", err)
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
