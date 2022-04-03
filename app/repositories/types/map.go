package types

import "time"

type Map struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	FileID    string    `gorm:"column:file_id"`
	AlertsKey string    `gorm:"column:alerts_key;unique"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}
