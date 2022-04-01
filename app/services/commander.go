package services

import (
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

func (r Commander) Track(ctx context.Context, msg *tgbotapi.Message, args string) (tgbotapi.MessageConfig, error) {
	if len(args) > 0 {
		if err := r.notification.Track(ctx, msg.Chat.ID, args); err != nil {
			if errors.Is(err, types.ErrLinkExists) {
				return tgbotapi.NewMessage(msg.Chat.ID, "вже пильную за "+args), nil
			}

			return tgbotapi.MessageConfig{}, fmt.Errorf("track: %w", err)
		}

		return tgbotapi.NewMessage(msg.Chat.ID, "буду пильнувати за "+args), nil
	}

	if err := r.chat.SetCommand(ctx, msg.Chat.ID, "track"); err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("set command: %w", err)
	}

	return tgbotapi.NewMessage(msg.Chat.ID, "вкажи територію, за якою пильнувати"), nil
}

func (r Commander) Tracking(ctx context.Context, msg *tgbotapi.Message, _ string) (tgbotapi.MessageConfig, error) {
	list, err := r.notification.Tracking(ctx, msg.Chat.ID)
	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("tracking: %w", err)
	}

	if len(list) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "ще нічого не трекаєш"), nil
	}

	return tgbotapi.NewMessage(msg.Chat.ID, strings.Join(list.Areas(), ", ")), nil
}

func (r Commander) Stop(ctx context.Context, msg *tgbotapi.Message, args string) (tgbotapi.MessageConfig, error) {
	if len(args) > 0 {
		if err := r.notification.Stop(ctx, msg.Chat.ID, args); err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("stop: %w", err)
		}

		return tgbotapi.NewMessage(msg.Chat.ID, "відписуюсь від "+args), nil
	}

	if err := r.chat.SetCommand(ctx, msg.Chat.ID, "stop"); err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("set command: %w", err)
	}

	return tgbotapi.NewMessage(msg.Chat.ID, "вкажи територію від якої відписатись"), nil
}

func (r Commander) Alerts(ctx context.Context, msg *tgbotapi.Message, _ string) (tgbotapi.MessageConfig, error) {
	alerts, err := r.alert.GetActive(ctx)
	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("get active: %w", err)
	}

	if len(alerts) == 0 {
		return tgbotapi.NewMessage(msg.Chat.ID, "все тихо"), nil
	}

	return tgbotapi.NewMessage(msg.Chat.ID, strings.Join(alerts.Areas(), ", ")), nil
}

func (r Commander) Start(_ context.Context, msg *tgbotapi.Message, _ string) (tgbotapi.MessageConfig, error) {
	return tgbotapi.NewMessage(
		msg.Chat.ID,
		`Пильнуй сповіщення в сусідніх областях.

Приклад, як створити сповіщення:
/areas
І далі наклацати області, які цікавлять.

Можна власноруч ще встановлювати регіони:
/track Чортківський
Назва має повністю збігатись з тією, що на карті https://war.ukrzen.in.ua/alerts/`,
	), nil
}

func (r Commander) Areas(ctx context.Context, msg *tgbotapi.Message, _ string) (tgbotapi.Chattable, error) {
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

	tracking, err := r.notification.Tracking(ctx, msg.Chat.ID)
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

	outMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	outMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
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

	return outMsg, nil
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

func (r Commander) Auth(ctx context.Context, msg *tgbotapi.Message, args string) (tgbotapi.Chattable, error) {
	split := strings.SplitN(args, ":", 2)
	if len(split) != 2 {
		return tgbotapi.NewMessage(msg.Chat.ID, "nopes"), nil
	}

	priv, passwd := split[0], split[1]
	if passwd != "passwd" {
		return tgbotapi.NewMessage(msg.Chat.ID, "nopes"), nil
	}

	switch priv {
	case "send_fake_event":
		if err := r.chat.Grant(ctx, msg.Chat.ID, "send_fake_event"); err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("grant: %w", err)
		}
	}

	return tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID), nil
}
