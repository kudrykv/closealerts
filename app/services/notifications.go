package services

import (
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Notification struct {
	notification repositories.Notification
	log          *zap.SugaredLogger
}

func (n Notification) Track(ctx context.Context, chatID int64, area string) error {
	if err := n.notification.Track(ctx, chatID, area); err != nil {
		return fmt.Errorf("track: %w", err)
	}

	return nil
}

func (n Notification) Tracking(ctx context.Context, id int64) ([]types2.Notification, error) {
	list, err := n.notification.Tracking(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tracking: %w", err)
	}

	return list, nil
}

func (n Notification) Stop(ctx context.Context, id int64, area string) error {
	if err := n.notification.Stop(ctx, id, area); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	return nil
}

func NewNotification(log *zap.SugaredLogger, notification repositories.Notification) Notification {
	return Notification{log: log, notification: notification}
}
