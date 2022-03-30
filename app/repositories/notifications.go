package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Notification struct {
	db  clients.DB
	log *zap.SugaredLogger
}

func (n Notification) Track(ctx context.Context, chatID int64, area string) error {
	err := n.db.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var notif types2.Notification

		err := tx.WithContext(ctx).Where("chat_id = ? and area = ?", chatID, area).Select(&notif).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("record already exists")
		}

		if err := tx.WithContext(ctx).Create(types2.Notification{ChatID: chatID, Area: area}).Error; err != nil {
			return fmt.Errorf("track %d %s: %w", chatID, area, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}

	return nil
}

func (n Notification) Tracking(ctx context.Context, id int64) ([]types2.Notification, error) {
	var out []types2.Notification

	if err := n.db.DB().WithContext(ctx).Where("chat_id = ?", id).Find(&out).Error; err != nil {
		return nil, fmt.Errorf("tracking %d: %w", id, err)
	}

	return out, nil
}

func NewNotification(log *zap.SugaredLogger, db clients.DB) Notification {
	return Notification{log: log, db: db}
}
