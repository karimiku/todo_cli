package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// ErrInvalidDisplayID indicates the provided display number is out of range.
var ErrInvalidDisplayID = errors.New("指定した番号のタスクは見つかりません")

// ErrInvalidPosition indicates the provided target position is invalid.
var ErrInvalidPosition = errors.New("移動先の番号が不正です")

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

// FocusTask moves the specified pending task to the very top of the queue.
func (s *Store) FocusTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}

	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		tasks, err := loadPendingTasksForUpdate(tx)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return ErrNoPendingTasks
		}
		if displayID > len(tasks) {
			return ErrInvalidDisplayID
		}

		target := tasks[displayID-1]
		now := time.Now().UTC()
		if displayID == 1 {
			if err := tx.Model(&Task{}).
				Where("id = ?", target.ID).
				Update("updated_at", now).Error; err != nil {
				return fmt.Errorf("touch task: %w", err)
			}

			target.UpdatedAt = now
			updated = target
			return nil
		}

		minOrder := tasks[0].OrderKey
		newOrder := minOrder - 1.0

		if err := tx.Model(&Task{}).
			Where("id = ?", target.ID).
			Updates(map[string]any{
				"order_key":  newOrder,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("update task order: %w", err)
		}

		target.OrderKey = newOrder
		target.UpdatedAt = now
		updated = target
		return nil
	})
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &updated, nil
}

// MoveTask reorders the pending task to the requested position (1-indexed).
func (s *Store) MoveTask(ctx context.Context, displayID, targetPos int) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}
	if targetPos < 1 {
		return nil, ErrInvalidPosition
	}

	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		tasks, err := loadPendingTasksForUpdate(tx)
		if err != nil {
			return err
		}
		n := len(tasks)
		if n == 0 {
			return ErrNoPendingTasks
		}
		if displayID > n {
			return ErrInvalidDisplayID
		}

		oldIdx := displayID - 1
		targetIndex := clamp(targetPos-1, 0, n-1)
		target := tasks[oldIdx]
		if oldIdx == targetIndex {
			target.UpdatedAt = time.Now().UTC()
			if err := tx.Model(&Task{}).
				Where("id = ?", target.ID).
				Update("updated_at", target.UpdatedAt).Error; err != nil {
				return fmt.Errorf("touch task: %w", err)
			}
			updated = target
			return nil
		}

		remaining := make([]Task, 0, n-1)
		for i, task := range tasks {
			if i == oldIdx {
				continue
			}
			remaining = append(remaining, task)
		}

		insertIdx := targetIndex
		if insertIdx < 0 {
			insertIdx = 0
		}
		if insertIdx > len(remaining) {
			insertIdx = len(remaining)
		}

		ordered := make([]Task, 0, n)
		ordered = append(ordered, remaining[:insertIdx]...)
		ordered = append(ordered, target)
		ordered = append(ordered, remaining[insertIdx:]...)
		finalIdx := insertIdx

		prev := (*Task)(nil)
		next := (*Task)(nil)
		if finalIdx > 0 {
			prev = &ordered[finalIdx-1]
		}
		if finalIdx+1 < len(ordered) {
			next = &ordered[finalIdx+1]
		}

		var newOrder float64
		switch {
		case prev == nil && next == nil:
			newOrder = 1.0
		case prev == nil:
			newOrder = next.OrderKey - 1.0
		case next == nil:
			newOrder = prev.OrderKey + 1.0
		default:
			gap := math.Abs(prev.OrderKey - next.OrderKey)
			if gap < normalizeThreshold {
				if err := rebalanceOrderKeys(tx, ordered); err != nil {
					return err
				}
				return tx.Where("id = ?", target.ID).First(&updated).Error
			}
			newOrder = (prev.OrderKey + next.OrderKey) / 2.0
		}

		now := time.Now().UTC()
		if err := tx.Model(&Task{}).
			Where("id = ?", target.ID).
			Updates(map[string]any{
				"order_key":  newOrder,
				"updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("update task order: %w", err)
		}

		target.OrderKey = newOrder
		target.UpdatedAt = now
		updated = target
		return nil
	})
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &updated, nil
}

// EditTask updates the text of the specified pending task.
func (s *Store) EditTask(ctx context.Context, displayID int, text string) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		tasks, err := loadPendingTasksForUpdate(tx)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return ErrNoPendingTasks
		}
		if displayID > len(tasks) {
			return ErrInvalidDisplayID
		}

		target := tasks[displayID-1]
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
		updated = target
		return nil
	})
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &updated, nil
}

// DeleteTask removes the specified pending task.
func (s *Store) DeleteTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}

	db := s.db.WithContext(ctx)
	var deleted Task
	err := db.Transaction(func(tx *gorm.DB) error {
		tasks, err := loadPendingTasksForUpdate(tx)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return ErrNoPendingTasks
		}
		if displayID > len(tasks) {
			return ErrInvalidDisplayID
		}

		target := tasks[displayID-1]
		if err := tx.Delete(&Task{}, target.ID).Error; err != nil {
			return fmt.Errorf("delete task: %w", err)
		}

		deleted = target
		return nil
	})
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &deleted, nil
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

const normalizeThreshold = 1e-3

func loadPendingTasksForUpdate(tx *gorm.DB) ([]Task, error) {
	var tasks []Task
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list pending tasks: %w", err)
	}

	return tasks, nil
}

func mapStorageError(err error) error {
	switch {
	case errors.Is(err, ErrNoPendingTasks):
		return ErrNoPendingTasks
	case errors.Is(err, ErrInvalidDisplayID):
		return ErrInvalidDisplayID
	case errors.Is(err, ErrInvalidPosition):
		return ErrInvalidPosition
	default:
		return err
	}
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func rebalanceOrderKeys(tx *gorm.DB, ordered []Task) error {
	now := time.Now().UTC()
	for idx, task := range ordered {
		order := float64(idx + 1)
		updates := map[string]any{
			"order_key":  order,
			"updated_at": now,
		}
		if err := tx.Model(&Task{}).
			Where("id = ?", task.ID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("rebalance task order: %w", err)
		}
		if task.ID == ordered[idx].ID {
			ordered[idx].OrderKey = order
			ordered[idx].UpdatedAt = now
		}
	}

	return nil
}
