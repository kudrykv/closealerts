package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
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
		//var notif types2.Notification
		//
		//err := tx.WithContext(ctx).Where("chat_id = ? and area = ?", chatID, area).Select(&notif).Error
		//if !errors.Is(err, gorm.ErrRecordNotFound) {
		//	return errors.New("record already exists")
		//}

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

func (n Notification) Stop(ctx context.Context, id int64, area string) error {
	err := n.db.DB().WithContext(ctx).Where("chat_id = ? and area = ?", id, area).Delete(&types2.Notification{}).Error
	if err != nil {
		return fmt.Errorf("delete %d %s: %w", id, area, err)
	}

	return nil
}

func (n Notification) Eligible(ctx context.Context, alerts []types2.Alert) ([]types2.Notification, error) {
	if len(alerts) == 0 {
		return nil, nil
	}

	areas := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		areas = append(areas, alert.ID)
	}

	var notif []types2.Notification

	err := n.db.DB().WithContext(ctx).Where("area in (?) and notified = false", areas).Find(&notif).Error
	if err != nil {
		return nil, fmt.Errorf("eligible: %w", err)
	}

	return notif, nil
}

func (n Notification) Notified(ctx context.Context, eligible types2.Notification) error {
	err := n.db.DB().
		WithContext(ctx).
		Model(&types2.Notification{}).
		Where("chat_id = ? and area = ?", eligible.ChatID, eligible.Area).
		UpdateColumn("notified", true).
		Error

	if err != nil {
		return fmt.Errorf("mark as notified: %w", err)
	}

	return nil
}

func (n Notification) Unmark(ctx context.Context, alerts []types2.Alert) ([]types2.Notification, error) {
	areas := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		areas = append(areas, alert.ID)
	}

	var unmark []types2.Notification

	err := n.db.DB().Transaction(func(tx *gorm.DB) error {
		if err := n.db.DB().WithContext(ctx).Where("area not in (?)", areas).Find(&unmark).Error; err != nil {
			return fmt.Errorf("get ones to unmark: %w", err)
		}

		err := n.db.DB().
			WithContext(ctx).
			Model(&types2.Notification{}).
			Where("area not in (?)", areas).
			UpdateColumn("notified", false).
			Error
		if err != nil {
			return fmt.Errorf("unmark: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}

	return unmark, nil
}

func NewNotification(log *zap.SugaredLogger, db clients.DB) Notification {
	return Notification{log: log, db: db}
}
