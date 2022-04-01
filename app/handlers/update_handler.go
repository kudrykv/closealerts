package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/services"
	"closealerts/app/types"
	"context"

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
	if msg == nil {
		return
	}

	r.log.Infow("msg", "user", msg.Chat.UserName, "text", msg.Text)

	var (
		text string
		err  error
	)

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

	switch command {
	case "track":
		text, err = r.commander.Track(ctx, chat, args)

	case "tracking":
		text, err = r.commander.Tracking(ctx, chat, args)

	case "stop":
		text, err = r.commander.Stop(ctx, chat, args)

	case "alerts":
		text, err = r.commander.Alerts(ctx, chat, args)

	default:
		text = "я такої команди не знаю"
	}

	if err != nil {
		r.log.Errorw("track", "err", err)
		r.bot.MaybeSendText(ctx, chat.ID, "в мене щось пішло не так, спробуй ще раз")
	}

	r.bot.MaybeSendText(ctx, chat.ID, text)

	if clearCmd {
		if err := r.chat.ClearCommand(ctx, chat.ID); err != nil {
			r.log.Errorw("clear command", "err", err)

			return
		}
	}
}
