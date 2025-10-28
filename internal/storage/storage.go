package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

// ErrEmptyText indicates the provided text was blank.
var ErrEmptyText = errors.New("空のテキストは不可")

// ErrNoPendingTasks is returned when an operation requires pending tasks but none exist.
var ErrNoPendingTasks = errors.New("未完了はありません 🎉")

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

// AddTask appends a new task to the end of the pending list.
func (s *Store) AddTask(ctx context.Context, text string) (*Task, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	var task Task
	db := s.db.WithContext(ctx)
	err := db.Transaction(func(tx *gorm.DB) error {
		var maxOrder sql.NullFloat64
		if err := tx.Model(&Task{}).
			Where("is_done = ?", false).
			Select("COALESCE(MAX(order_key), 0)").
			Scan(&maxOrder).Error; err != nil {
			return fmt.Errorf("select max order: %w", err)
		}

		now := time.Now().UTC()
		task = Task{
			Text:      trimmed,
			IsDone:    false,
			OrderKey:  maxOrder.Float64 + 1.0,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := tx.Create(&task).Error; err != nil {
			return fmt.Errorf("insert task: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// ListTasks retrieves pending and completed tasks.
func (s *Store) ListTasks(ctx context.Context) (pending []Task, completed []Task, err error) {
	db := s.db.WithContext(ctx)

	if err = db.Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Find(&pending).Error; err != nil {
		return nil, nil, fmt.Errorf("list pending tasks: %w", err)
	}

	if err = db.Where("is_done = ?", true).
		Order("updated_at desc, id desc").
		Limit(50).
		Find(&completed).Error; err != nil {
		return nil, nil, fmt.Errorf("list completed tasks: %w", err)
	}

	return pending, completed, nil
}

// CompleteNext marks the top-most pending task as done.
func (s *Store) CompleteNext(ctx context.Context) (*Task, error) {
	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoPendingTasks
			}
			return fmt.Errorf("select next task: %w", err)
		}

		now := time.Now().UTC()
		if err := tx.Model(&Task{}).
			Where("id = ?", task.ID).
			Updates(map[string]any{
				"is_done":    true,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("mark task complete: %w", err)
		}

		task.IsDone = true
		task.UpdatedAt = now
		updated = task
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrNoPendingTasks) {
			return nil, ErrNoPendingTasks
		}
		return nil, err
	}

	return &updated, nil
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
