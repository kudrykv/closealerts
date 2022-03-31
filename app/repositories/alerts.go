package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"

	"gorm.io/gorm/clause"
)

type Alerts struct {
	db clients.DB
}

func NewAlerts(db clients.DB) Alerts {
	return Alerts{db: db}
}

func (r Alerts) ReplaceAlerts(ctx context.Context, alerts []types2.Alert) error {
	if len(alerts) == 0 {
		if err := r.db.DB().WithContext(ctx).Where("1 = 1").Delete(&types2.Alert{}).Error; err != nil {
			return fmt.Errorf("clear alerts: %w", err)
		}

		return nil
	}

	cond := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}

	if err := r.db.DB().WithContext(ctx).Clauses(cond).Create(alerts).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	ids := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		ids = append(ids, alert.ID)
	}

	if err := r.db.DB().WithContext(ctx).Where("id not in (?)", ids).Delete(&types2.Alert{}).Error; err != nil {
		return fmt.Errorf("delete rest: %w", err)
	}

	return nil
}

func (r Alerts) GetActive(ctx context.Context) ([]types2.Alert, error) {
	var list []types2.Alert
	if err := r.db.DB().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("get alerts from db: %w", err)
	}

	return list, nil
}
