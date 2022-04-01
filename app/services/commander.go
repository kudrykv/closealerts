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

	var areasTracking []string

	for _, notification := range tracking {
		if _, ok := areas[notification.Area]; ok {
			areasTracking = append(areasTracking, notification.Area)
		}
	}

	text := "можеш обрати на які області підписатись"
	if len(areasTracking) > 0 {
		text += "\n\nПідписки: " + strings.Join(areasTracking, ", ")
	}

	msg := tgbotapi.NewMessage(chat.ID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Волинська", "Волинська"),
			tgbotapi.NewInlineKeyboardButtonData("Вінницька", "Вінницька"),
			tgbotapi.NewInlineKeyboardButtonData("Дніпропетровська", "Дніпропетровська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Донецька", "Донецька"),
			tgbotapi.NewInlineKeyboardButtonData("Житомирська", "Житомирська"),
			tgbotapi.NewInlineKeyboardButtonData("Закарпатська", "Закарпатська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Запорізька", "Запорізька"),
			tgbotapi.NewInlineKeyboardButtonData("Івано-Франківська", "Івано-Франківська"),
			tgbotapi.NewInlineKeyboardButtonData("Київська", "Київська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Кіровоградська", "Кіровоградська"),
			tgbotapi.NewInlineKeyboardButtonData("Луганська", "Луганська"),
			tgbotapi.NewInlineKeyboardButtonData("Львівська", "Львівська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Миколаївська", "Миколаївська"),
			tgbotapi.NewInlineKeyboardButtonData("Одеська", "Одеська"),
			tgbotapi.NewInlineKeyboardButtonData("Полтавська", "Полтавська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Рівненська", "Рівненська"),
			tgbotapi.NewInlineKeyboardButtonData("Сумська", "Сумська"),
			tgbotapi.NewInlineKeyboardButtonData("Тернопільська", "Тернопільська"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Харківська", "Харківська"),
			tgbotapi.NewInlineKeyboardButtonData("Херсонська", "Херсонська"),
			tgbotapi.NewInlineKeyboardButtonData("Хмельницька", "Хмельницька"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Черкаська", "Черкаська"),
			tgbotapi.NewInlineKeyboardButtonData("Чернівецька", "Чернівецька"),
			tgbotapi.NewInlineKeyboardButtonData("Чернігівська", "Чернігівська"),
		),
	)

	return msg, nil
}
