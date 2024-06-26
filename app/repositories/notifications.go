package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"closealerts/app/types"
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

func (r Notification) Track(ctx context.Context, chatID int64, area string) error {
	err := r.db.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var notif types2.Notification

		err := tx.WithContext(ctx).Where("chat_id = ? and area = ?", chatID, area).Take(&notif).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("select: %w", err)
		}

		if notif.ChatID > 0 {
			return fmt.Errorf("%d-%s: %w", chatID, area, types.ErrLinkExists)
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

func (r Notification) Tracking(ctx context.Context, id int64) ([]types2.Notification, error) {
	var out []types2.Notification

	if err := r.db.DB().WithContext(ctx).Where("chat_id = ?", id).Find(&out).Error; err != nil {
		return nil, fmt.Errorf("tracking %d: %w", id, err)
	}

	return out, nil
}

func (r Notification) Stop(ctx context.Context, id int64, area string) error {
	err := r.db.DB().WithContext(ctx).Where("chat_id = ? and area = ?", id, area).Delete(&types2.Notification{}).Error
	if err != nil {
		return fmt.Errorf("delete %d %s: %w", id, area, err)
	}

	return nil
}

func (r Notification) Eligible(ctx context.Context, alerts []types2.Alert) (types2.Notifications, error) {
	if len(alerts) == 0 {
		return nil, nil
	}

	areas := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		areas = append(areas, alert.ID)
	}

	var notif types2.Notifications

	err := r.db.DB().WithContext(ctx).
		Where("area in (?) and notified = false", areas).
		Order("chat_id").
		Find(&notif).
		Error
	if err != nil {
		return nil, fmt.Errorf("eligible: %w", err)
	}

	r.log.Debugw("eligible", "group", notif.GroupByChatID())

	return notif, nil
}

func (r Notification) Notified(ctx context.Context, eligible types2.Notification) error {
	err := r.db.DB().
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

func (r Notification) Unmark(ctx context.Context, alerts []types2.Alert) error {
	areas := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		areas = append(areas, alert.ID)
	}

	tx := r.db.DB().WithContext(ctx).Model(&types2.Notification{})

	if len(areas) > 0 {
		tx = tx.Where("area not in (?)", areas)
	} else {
		tx = tx.Where("1 = 1")
	}

	if err := tx.UpdateColumn("notified", false).Error; err != nil {
		return fmt.Errorf("unmark: %w", err)
	}

	r.log.Debugw("unmark areas where alert has ended")

	return nil
}

func (r Notification) AlertEnded(ctx context.Context, alerts []types2.Alert) (types2.Notifications, error) {
	areas := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		areas = append(areas, alert.ID)
	}

	var (
		endedFor types2.Notifications
		err      error
	)

	if len(alerts) == 0 {
		err = r.db.DB().WithContext(ctx).Where("notified = true").Find(&endedFor).Error
	} else {
		err = r.db.DB().WithContext(ctx).Where("area not in (?) and notified = true", areas).Find(&endedFor).Error
	}

	if err != nil {
		return nil, fmt.Errorf("alert ended: %w", err)
	}

	r.log.Debug("mark which alerts has ended")

	return endedFor, nil
}

func NewNotification(log *zap.SugaredLogger, db clients.DB) Notification {
	return Notification{log: log, db: db}
}
