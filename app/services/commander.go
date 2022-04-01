package services

import (
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
	"strings"
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

func (r Commander) Track(ctx context.Context, chat types2.Chat, args string) (string, error) {
	if len(args) > 0 {
		if err := r.notification.Track(ctx, chat.ID, args); err != nil {
			return "", fmt.Errorf("track: %w", err)
		}

		return "буду пильнувати за " + args, nil
	}

	if err := r.chat.SetCommand(ctx, chat.ID, "track"); err != nil {
		return "", fmt.Errorf("set command: %w", err)
	}

	return "вкажи територію, за якою пильнувати", nil
}

func (r Commander) Tracking(ctx context.Context, chat types2.Chat, _ string) (string, error) {
	list, err := r.notification.Tracking(ctx, chat.ID)
	if err != nil {
		return "", fmt.Errorf("tracking: %w", err)
	}

	if len(list) == 0 {
		return "ще нічого не трекаєш", nil
	}

	return strings.Join(list.Areas(), ", "), nil
}

func (r Commander) Stop(ctx context.Context, chat types2.Chat, args string) (string, error) {
	if len(args) > 0 {
		if err := r.notification.Stop(ctx, chat.ID, args); err != nil {
			return "", fmt.Errorf("stop: %w", err)
		}

		return "відписуюсь від " + args, nil
	}

	if err := r.chat.SetCommand(ctx, chat.ID, "stop"); err != nil {
		return "", fmt.Errorf("set command: %w", err)
	}

	return "вкажи територію від якої відписатись", nil
}

func (r Commander) Alerts(ctx context.Context, _ types2.Chat, _ string) (string, error) {
	alerts, err := r.alert.GetActive(ctx)
	if err != nil {
		return "", fmt.Errorf("get active: %w", err)
	}

	if len(alerts) == 0 {
		return "все тихо", nil
	}

	return strings.Join(alerts.Areas(), ", "), nil
}