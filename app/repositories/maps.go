package repositories

import (
	"closealerts/app/clients"
	types2 "closealerts/app/repositories/types"
	"context"
	"fmt"
)

type Maps struct {
	db clients.DB
}

func NewMaps(db clients.DB) Maps {
	return Maps{db: db}
}

func (r Maps) Get(ctx context.Context, alertsKey string) (types2.Map, error) {
	var mapz types2.Map
	if err := r.db.DB().WithContext(ctx).Where("alerts_key = ?", alertsKey).First(&mapz).Error; err != nil {
		return types2.Map{}, fmt.Errorf("select: %w", err)
	}

	return mapz, nil
}

func (r Maps) Save(ctx context.Context, key string, fileID string) (types2.Map, error) {
	mapz := types2.Map{AlertsKey: key, FileID: fileID}
	if err := r.db.DB().WithContext(ctx).Create(&mapz).Error; err != nil {
		return mapz, fmt.Errorf("create: %w", err)
	}

	return mapz, nil
}
