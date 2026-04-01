package db

import (
	"fmt"
	"os"
	"path/filepath"

	"tempmail/backend/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dbPath string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.User{},
		&models.Domain{},
		&models.Mailbox{},
		&models.Message{},
	); err != nil {
		return nil, fmt.Errorf("migrate schema: %w", err)
	}

	return db, nil
}
