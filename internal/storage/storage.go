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

// ErrInvalidPosition is returned when a move position is less than 1.
var ErrInvalidPosition = errors.New("順位は 1 以上を指定してください")

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

// FocusTask moves the task at displayID to the top position (displayID=1).
func (s *Store) FocusTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, fmt.Errorf("表示IDは1以上である必要があります")
	}

	db := s.db.WithContext(ctx)
	var focused Task
	err := db.Transaction(func(tx *gorm.DB) error {
		var pending []Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			Find(&pending).Error; err != nil {
			return fmt.Errorf("select pending tasks: %w", err)
		}

		if len(pending) == 0 {
			return ErrNoPendingTasks
		}

		if displayID > len(pending) {
			return fmt.Errorf("未完了の件数は %d 件です", len(pending))
		}

		idx := displayID - 1
		target := pending[idx]

		// If already at position 1, just update timestamp
		if idx == 0 {
			now := time.Now().UTC()
			if err := tx.Model(&Task{}).
				Where("id = ?", target.ID).
				Update("updated_at", now).Error; err != nil {
				return fmt.Errorf("update timestamp: %w", err)
			}
			target.UpdatedAt = now
			focused = target
			return nil
		}

		// Move to top: set order_key to minOrder - 1.0
		minOrder := pending[0].OrderKey
		newOrder := minOrder - 1.0

		now := time.Now().UTC()
		if err := tx.Model(&Task{}).
			Where("id = ?", target.ID).
			Updates(map[string]any{
				"order_key":  newOrder,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("update order_key: %w", err)
		}

		target.OrderKey = newOrder
		target.UpdatedAt = now
		focused = target

		// Normalize if needed
		if newOrder < -1e9 || minOrder-newOrder < 1e-6 {
			if err := normalize(tx); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &focused, nil
}

// MoveTask moves the task at fromDisplayID to toPosition.
func (s *Store) MoveTask(ctx context.Context, fromDisplayID, toPosition int) (*Task, error) {
	if fromDisplayID < 1 {
		return nil, fmt.Errorf("表示IDは1以上である必要があります")
	}
	if toPosition < 1 {
		return nil, ErrInvalidPosition
	}

	db := s.db.WithContext(ctx)
	var moved Task
	err := db.Transaction(func(tx *gorm.DB) error {
		var pending []Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			Find(&pending).Error; err != nil {
			return fmt.Errorf("select pending tasks: %w", err)
		}

		if len(pending) == 0 {
			return ErrNoPendingTasks
		}

		if fromDisplayID > len(pending) {
			return fmt.Errorf("未完了の件数は %d 件です", len(pending))
		}

		// Clamp toPosition to valid range
		if toPosition > len(pending) {
			toPosition = len(pending)
		}

		fromIdx := fromDisplayID - 1
		toIdx := toPosition - 1
		target := pending[fromIdx]

		// No movement needed
		if fromIdx == toIdx {
			now := time.Now().UTC()
			if err := tx.Model(&Task{}).
				Where("id = ?", target.ID).
				Update("updated_at", now).Error; err != nil {
				return fmt.Errorf("update timestamp: %w", err)
			}
			target.UpdatedAt = now
			moved = target
			return nil
		}

		var newOrder float64
		if toIdx == 0 {
			// Move to top
			newOrder = pending[0].OrderKey - 1.0
		} else if toIdx == len(pending)-1 {
			// Move to bottom
			newOrder = pending[len(pending)-1].OrderKey + 1.0
		} else {
			// Move to middle
			var prev, next float64
			if fromIdx < toIdx {
				// Moving down: insert after toIdx
				prev = pending[toIdx].OrderKey
				if toIdx+1 < len(pending) {
					next = pending[toIdx+1].OrderKey
				} else {
					next = prev + 2.0
				}
			} else {
				// Moving up: insert before toIdx
				next = pending[toIdx].OrderKey
				if toIdx > 0 {
					prev = pending[toIdx-1].OrderKey
				} else {
					prev = next - 2.0
				}
			}
			newOrder = (prev + next) / 2.0
		}

		now := time.Now().UTC()
		if err := tx.Model(&Task{}).
			Where("id = ?", target.ID).
			Updates(map[string]any{
				"order_key":  newOrder,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("update order_key: %w", err)
		}

		target.OrderKey = newOrder
		target.UpdatedAt = now
		moved = target

		// Check if normalization is needed
		needsNormalize := false
		if toIdx > 0 && toIdx < len(pending)-1 {
			var prev, next float64
			if fromIdx < toIdx {
				prev = pending[toIdx].OrderKey
				if toIdx+1 < len(pending) {
					next = pending[toIdx+1].OrderKey
				}
			} else {
				next = pending[toIdx].OrderKey
				if toIdx > 0 {
					prev = pending[toIdx-1].OrderKey
				}
			}
			if prev != 0 && next != 0 {
				gap := next - prev
				if gap < 1e-3 && gap > 0 {
					needsNormalize = true
				}
			}
		}

		if needsNormalize {
			if err := normalize(tx); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &moved, nil
}

// EditTask updates the text of the task at displayID.
func (s *Store) EditTask(ctx context.Context, displayID int, newText string) (*Task, error) {
	if displayID < 1 {
		return nil, fmt.Errorf("表示IDは1以上である必要があります")
	}

	trimmed := strings.TrimSpace(newText)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	db := s.db.WithContext(ctx)
	var edited Task
	err := db.Transaction(func(tx *gorm.DB) error {
		var pending []Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			Find(&pending).Error; err != nil {
			return fmt.Errorf("select pending tasks: %w", err)
		}

		if len(pending) == 0 {
			return ErrNoPendingTasks
		}

		if displayID > len(pending) {
			return fmt.Errorf("未完了の件数は %d 件です", len(pending))
		}

		idx := displayID - 1
		target := pending[idx]

		now := time.Now().UTC()
		if err := tx.Model(&Task{}).
			Where("id = ?", target.ID).
			Updates(map[string]any{
				"text":       trimmed,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("update task text: %w", err)
		}

		target.Text = trimmed
		target.UpdatedAt = now
		edited = target
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &edited, nil
}

// DeleteTask physically deletes the task at displayID.
func (s *Store) DeleteTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, fmt.Errorf("表示IDは1以上である必要があります")
	}

	db := s.db.WithContext(ctx)
	var deleted Task
	err := db.Transaction(func(tx *gorm.DB) error {
		var pending []Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			Find(&pending).Error; err != nil {
			return fmt.Errorf("select pending tasks: %w", err)
		}

		if len(pending) == 0 {
			return ErrNoPendingTasks
		}

		if displayID > len(pending) {
			return fmt.Errorf("未完了の件数は %d 件です", len(pending))
		}

		idx := displayID - 1
		target := pending[idx]
		deleted = target

		result := tx.Where("id = ? AND is_done = ?", target.ID, false).
			Delete(&Task{})
		if result.Error != nil {
			return fmt.Errorf("delete task: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("タスクの削除に失敗しました")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &deleted, nil
}

// normalize re-assigns order_key values to 1.0, 2.0, ... for all pending tasks.
func normalize(tx *gorm.DB) error {
	var pending []Task
	if err := tx.Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Find(&pending).Error; err != nil {
		return fmt.Errorf("normalize: select pending tasks: %w", err)
	}

	for i := range pending {
		newOrder := float64(i + 1)
		if err := tx.Model(&Task{}).
			Where("id = ?", pending[i].ID).
			Update("order_key", newOrder).Error; err != nil {
			return fmt.Errorf("normalize: update order_key: %w", err)
		}
	}

	return nil
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
