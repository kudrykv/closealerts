package clients

import (
	"closealerts/app/types"
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	db *gorm.DB
}

func NewDBFromSQLite(config types.Config) (DB, error) {
	db, err := gorm.Open(sqlite.Open(config.SQLite3DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return DB{}, fmt.Errorf("gorm open sqlite open: %w", err)
	}

	return DB{db: db}, nil
}

func (r DB) AutoMigrate(dst ...interface{}) error {
	if err := r.db.AutoMigrate(dst...); err != nil {
		return fmt.Errorf("db auto migrate: %w", err)
	}

	return nil
}

func (r DB) Insert(val interface{}) error {
	if err := r.db.Create(val).Error; err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (r DB) DB() *gorm.DB {
	return r.db
}
