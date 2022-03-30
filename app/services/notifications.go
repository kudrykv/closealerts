package services

import (
	"closealerts/app/repositories"
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

func NewNotification(log *zap.SugaredLogger, notification repositories.Notification) Notification {
	return Notification{log: log, notification: notification}
}
