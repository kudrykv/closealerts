package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Notification struct {
	db  clients.DB
	log *zap.SugaredLogger
}

func (n Notification) Track(ctx context.Context, chatID int64, area string) error {
	if err := n.db.DB().WithContext(ctx).Create(types2.Notification{ChatID: chatID, Area: area}).Error; err != nil {
		return fmt.Errorf("track %d %s: %w", chatID, area, err)
	}

	return nil
}

func NewNotification(log *zap.SugaredLogger, db clients.DB) Notification {
	return Notification{log: log, db: db}
}
