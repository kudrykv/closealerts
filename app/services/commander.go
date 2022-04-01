package services

import (
	types2 "closealerts/app/repositories/types"
	"closealerts/app/types"
	"context"
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Commander struct {
	notification Notification
	chat         Chats
	alert        Alerts
}

func NewCommander(
	chat Chats,
	notification Notification,
	alert Alerts,
) Commander {
	return Commander{
		chat:         chat,
		notification: notification,
		alert:        alert,
	}
}

func (r Commander) Track(ctx context.Context, chat types2.Chat, args string) (tgbotapi.MessageConfig, error) {
	if len(args) > 0 {
		if err := r.notification.Track(ctx, chat.ID, args); err != nil {
			if errors.Is(err, types.ErrLinkExists) {
				return tgbotapi.NewMessage(chat.ID, "вже пильную за "+args), nil
			}

			return tgbotapi.MessageConfig{}, fmt.Errorf("track: %w", err)
		}

		return tgbotapi.NewMessage(chat.ID, "буду пильнувати за "+args), nil
	}

	if err := r.chat.SetCommand(ctx, chat.ID, "track"); err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("set command: %w", err)
	}

	return tgbotapi.NewMessage(chat.ID, "вкажи територію, за якою пильнувати"), nil
}

func (r Commander) Tracking(ctx context.Context, chat types2.Chat, _ string) (tgbotapi.MessageConfig, error) {
	list, err := r.notification.Tracking(ctx, chat.ID)
	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("tracking: %w", err)
	}

	if len(list) == 0 {
		return tgbotapi.NewMessage(chat.ID, "ще нічого не трекаєш"), nil
	}

	return tgbotapi.NewMessage(chat.ID, strings.Join(list.Areas(), ", ")), nil
}

func (r Commander) Stop(ctx context.Context, chat types2.Chat, args string) (tgbotapi.MessageConfig, error) {
	if len(args) > 0 {
		if err := r.notification.Stop(ctx, chat.ID, args); err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("stop: %w", err)
		}

		return tgbotapi.NewMessage(chat.ID, "відписуюсь від "+args), nil
	}

	if err := r.chat.SetCommand(ctx, chat.ID, "stop"); err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("set command: %w", err)
	}

	return tgbotapi.NewMessage(chat.ID, "вкажи територію від якої відписатись"), nil
}

func (r Commander) Alerts(ctx context.Context, chat types2.Chat, _ string) (tgbotapi.MessageConfig, error) {
	alerts, err := r.alert.GetActive(ctx)
	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("get active: %w", err)
	}

	if len(alerts) == 0 {
		return tgbotapi.NewMessage(chat.ID, "все тихо"), nil
	}

	return tgbotapi.NewMessage(chat.ID, strings.Join(alerts.Areas(), ", ")), nil
}

func (r Commander) Start(_ context.Context, chat types2.Chat, _ string) (tgbotapi.MessageConfig, error) {
	return tgbotapi.NewMessage(
		chat.ID,
		`Пильнуй сповіщення в сусідніх областях.

Приклад, як створити сповіщення:
/track Тернопільська

Назва має повністю збігатись з тією, що на карті https://war.ukrzen.in.ua/alerts/`,
	), nil
}

func (r Commander) Areas(ctx context.Context, chat types2.Chat, _ string) (tgbotapi.Chattable, error) {
	areas := map[string]struct{}{
		"Волинська":         {},
		"Вінницька":         {},
		"Дніпропетровська":  {},
		"Донецька":          {},
		"Житомирська":       {},
		"Закарпатська":      {},
		"Запорізька":        {},
		"Івано-Франківська": {},
		"Київська":          {},
		"Кіровоградська":    {},
		"Луганська":         {},
		"Львівська":         {},
		"Миколаївська":      {},
		"Одеська":           {},
		"Полтавська":        {},
		"Рівненська":        {},
		"Сумська":           {},
		"Тернопільська":     {},
		"Харківська":        {},
		"Херсонська":        {},
		"Хмельницька":       {},
		"Черкаська":         {},
		"Чернівецька":       {},
		"Чернігівська":      {},
	}

	tracking, err := r.notification.Tracking(ctx, chat.ID)
	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("tracking: %w", err)
	}

	var areasTracking types.Stringies

	for _, notification := range tracking {
		if _, ok := areas[notification.Area]; ok {
			areasTracking = append(areasTracking, notification.Area)
		}
	}

	text := "можеш обрати на які області підписатись"
	if len(areasTracking) > 0 {
		text += "\n\nПідписки: " + areasTracking.Sort().Join(", ")
	}

	msg := tgbotapi.NewMessage(chat.ID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Волинська", "toggle_area:Волинська"),
			tgbotapi.NewInlineKeyboardButtonData("Вінницька", "toggle_area:Вінницька"),
			tgbotapi.NewInlineKeyboardButtonData("Дніпропетровська", "toggle_area:Дніпропетровська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Донецька", "toggle_area:Донецька"),
			tgbotapi.NewInlineKeyboardButtonData("Житомирська", "toggle_area:Житомирська"),
			tgbotapi.NewInlineKeyboardButtonData("Закарпатська", "toggle_area:Закарпатська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Запорізька", "toggle_area:Запорізька"),
			tgbotapi.NewInlineKeyboardButtonData("Івано-Франківська", "toggle_area:Івано-Франківська"),
			tgbotapi.NewInlineKeyboardButtonData("Київська", "toggle_area:Київська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Кіровоградська", "toggle_area:Кіровоградська"),
			tgbotapi.NewInlineKeyboardButtonData("Луганська", "toggle_area:Луганська"),
			tgbotapi.NewInlineKeyboardButtonData("Львівська", "toggle_area:Львівська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Миколаївська", "toggle_area:Миколаївська"),
			tgbotapi.NewInlineKeyboardButtonData("Одеська", "toggle_area:Одеська"),
			tgbotapi.NewInlineKeyboardButtonData("Полтавська", "toggle_area:Полтавська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Рівненська", "toggle_area:Рівненська"),
			tgbotapi.NewInlineKeyboardButtonData("Сумська", "toggle_area:Сумська"),
			tgbotapi.NewInlineKeyboardButtonData("Тернопільська", "toggle_area:Тернопільська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Харківська", "toggle_area:Харківська"),
			tgbotapi.NewInlineKeyboardButtonData("Херсонська", "toggle_area:Херсонська"),
			tgbotapi.NewInlineKeyboardButtonData("Хмельницька", "toggle_area:Хмельницька"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Черкаська", "toggle_area:Черкаська"),
			tgbotapi.NewInlineKeyboardButtonData("Чернівецька", "toggle_area:Чернівецька"),
			tgbotapi.NewInlineKeyboardButtonData("Чернігівська", "toggle_area:Чернігівська"),
		),
	)

	return msg, nil
}

func (r Commander) ToggleArea(
	ctx context.Context, cq *tgbotapi.CallbackQuery, payload string,
) (tgbotapi.EditMessageTextConfig, error) {
	tracking, err := r.notification.Tracking(ctx, cq.Message.Chat.ID)
	if err != nil {
		return tgbotapi.EditMessageTextConfig{}, fmt.Errorf("tracking: %w", err)
	}

	trackingAreas := tracking.Areas()

	if tracking.Tracking(payload) {
		if err = r.notification.Stop(ctx, cq.Message.Chat.ID, payload); err != nil {
			return tgbotapi.EditMessageTextConfig{}, fmt.Errorf("stop: %w", err)
		}

		trackingAreas = trackingAreas.Delete(payload)
	} else {
		if err = r.notification.Track(ctx, cq.Message.Chat.ID, payload); err != nil {
			return tgbotapi.EditMessageTextConfig{}, fmt.Errorf("track: %w", err)
		}

		trackingAreas = append(trackingAreas, payload)
	}

	text := "Підписки: " + trackingAreas.Sort().Join(", ")
	if len(trackingAreas) == 0 {
		text = "Нема підписок"
	}

	return tgbotapi.
			NewEditMessageTextAndMarkup(cq.Message.Chat.ID, cq.Message.MessageID, text, *cq.Message.ReplyMarkup),
		nil
}
