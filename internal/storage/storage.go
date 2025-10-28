package storage

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Task represents a Todo item persisted in SQLite.
type Task struct {
	ID        uint      `gorm:"primaryKey"`
	Text      string    `gorm:"type:text;not null"`
	IsDone    bool      `gorm:"not null;index:idx_tasks_undone_order,priority:1"`
	OrderKey  float64   `gorm:"not null;index:idx_tasks_undone_order,priority:2"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}

// Store owns the database connection used across commands.
type Store struct {
	db *gorm.DB
}

// Open initialises a new Store backed by SQLite located at the provided path.
func Open(path string) (*Store, error) {
	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := applyPragmas(gormDB); err != nil {
		return nil, err
	}

	if err := gormDB.AutoMigrate(&Task{}); err != nil {
		return nil, fmt.Errorf("auto migrate task schema: %w", err)
	}

	return &Store{db: gormDB}, nil
}

// DB exposes the underlying *gorm.DB, primarily for transactional operations.
func (s *Store) DB() *gorm.DB {
	return s.db
}

// Close releases the underlying database resources.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("access sql.DB: %w", err)
	}

	return sqlDB.Close()
}

func applyPragmas(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("access sql.DB: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 4000;",
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return fmt.Errorf("apply %s: %w", pragma, err)
		}
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(0)

	return nil
}
