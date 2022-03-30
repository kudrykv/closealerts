package handlers

import (
	"closealerts/app/clients"
	"closealerts/app/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandler struct {
	bot clients.Telegram
}

func NewUpdate(bot clients.Telegram) UpdateHandler {
	return UpdateHandler{bot: bot}
}

func (h UpdateHandler) Handle(_ context.Context, update types.Update) {
	bts, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		log.Println(fmt.Errorf("marshal indent: %w", err))

		return
	}

	log.Println(string(bts))

	if msg := update.Message; msg != nil {
		if strings.ToLower(msg.Text) == "test" {
			outgoing := tgbotapi.NewMessage(msg.Chat.ID, "some text in the message")
			outgoing.ReplyToMessageID = msg.MessageID
			outgoing.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.NewInlineKeyboardButtonData("label 1", "data1"),
						tgbotapi.NewInlineKeyboardButtonData("label 2", "data2"),
					},
				},
			}

			if _, err := h.bot.Client.Send(outgoing); err != nil {
				fmt.Println(fmt.Errorf("bot client send: %w", err))

				return
			}
		}
	}

	if cb := update.CallbackQuery; cb != nil {
		var outgoingEditMsg tgbotapi.EditMessageTextConfig

		if cb.Data == "data1" {
			outgoingEditMsg = tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, cb.From.UserName)
		} else {
			outgoingEditMsg = tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID, cb.From.LanguageCode)
		}

		outgoingEditMsg.ReplyMarkup = cb.Message.ReplyMarkup

		if _, err := h.bot.Client.Send(outgoingEditMsg); err != nil {
			fmt.Println(fmt.Errorf("bot client send: %w", err))

			return
		}
	}
}
