package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Task はSQLiteに永続化されるTodoアイテムを表します。
// タスクの順序管理には浮動小数点数の order_key を使用し、
// 挿入や移動の際に中間値を計算することで順序を柔軟に管理します。
type Task struct {
	ID        uint      `gorm:"primaryKey"`                                       // タスクの一意識別子
	Text      string    `gorm:"type:text;not null"`                               // タスクの本文
	IsDone    bool      `gorm:"not null;index:idx_tasks_undone_order,priority:1"` // 完了フラグ（false=未完了、true=完了）
	OrderKey  float64   `gorm:"not null;index:idx_tasks_undone_order,priority:2"` // 順序を決定するキー（小さいほど優先度が高い）
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`                          // 作成日時
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`                          // 更新日時
}

// Store はコマンド間で共有されるデータベース接続を管理します。
// すべてのタスク操作はこのStoreを通じて行われます。
type Store struct {
	db *gorm.DB // GORMのデータベース接続
}

// ErrEmptyText は空のテキストが提供された場合に返されるエラーです。
var ErrEmptyText = errors.New("空のテキストは不可")

// ErrNoPendingTasks は未完了タスクが必要な操作で未完了タスクが存在しない場合に返されるエラーです。
var ErrNoPendingTasks = errors.New("未完了はありません 🎉")

// ErrInvalidDisplayID は指定された表示番号が範囲外の場合に返されるエラーです。
var ErrInvalidDisplayID = errors.New("指定した番号のタスクは見つかりません")

// ErrInvalidPosition は指定された移動先の位置が不正な場合に返されるエラーです。
var ErrInvalidPosition = errors.New("移動先の番号が不正です")

// Open は指定されたパスのSQLiteデータベースを使用して新しいStoreを初期化します。
// データベースが存在しない場合は作成し、スキーマのマイグレーションも自動的に実行します。
func Open(path string) (*Store, error) {
	// SQLiteデータベースを開く（ログは出力しない）
	gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// SQLiteの設定（WALモード、タイムアウトなど）を適用
	if err := applyPragmas(gormDB); err != nil {
		return nil, err
	}

	// スキーマの自動マイグレーション（テーブルが存在しない場合は作成）
	if err := gormDB.AutoMigrate(&Task{}); err != nil {
		return nil, fmt.Errorf("auto migrate task schema: %w", err)
	}

	// データベースファイルのパーミッション設定（セキュリティ対策）
	if err := ensureDBPermissions(path); err != nil {
		return nil, err
	}

	return &Store{db: gormDB}, nil
}

// DB は内部の *gorm.DB を公開します。
// 主にトランザクション操作など、低レベルなデータベース操作が必要な場合に使用します。
func (s *Store) DB() *gorm.DB {
	return s.db
}

// Close はデータベース接続を閉じてリソースを解放します。
// アプリケーション終了時やテスト終了時に呼び出す必要があります。
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("access sql.DB: %w", err)
	}

	return sqlDB.Close()
}

// AddTask は新しいタスクを未完了リストの末尾に追加します。
func (s *Store) AddTask(ctx context.Context, text string) (*Task, error) {
	// 前後の空白を削除し、空文字チェック
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	var task Task
	db := s.db.WithContext(ctx)
	// トランザクション内で実行（競合を防ぐため）
	err := db.Transaction(func(tx *gorm.DB) error {
		// 未完了タスクの最大 order_key を取得（タスクが存在しない場合は0）
		var maxOrder sql.NullFloat64
		if err := tx.Model(&Task{}).
			Where("is_done = ?", false).
			Select("COALESCE(MAX(order_key), 0)").
			Scan(&maxOrder).Error; err != nil {
			return fmt.Errorf("select max order: %w", err)
		}

		// 新しいタスクを作成（order_key は最大値+1）
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

// FocusTask は指定されたタスクを未完了リストの最上位に移動します。
func (s *Store) FocusTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}

	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		// 排他ロックをかけて未完了タスクを取得
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

		// 既に最上位の場合は updated_at を更新するだけ
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

		// 現在の最上位タスクの order_key より小さい値を設定
		minOrder := tasks[0].OrderKey
		newOrder := minOrder - focusOrderStep

		// order_key が下限を下回るか、差が小さすぎる場合は再調整が必要
		shouldNormalize := newOrder < focusOrderFloor || math.Abs(minOrder-newOrder) < focusNormalizeEpsilon

		if shouldNormalize {
			// タスクを再配置して order_key を1, 2, 3...と再計算
			ordered := make([]Task, 0, len(tasks))
			ordered = append(ordered, target) // 対象タスクを先頭に
			for idx, task := range tasks {
				if idx == displayID-1 {
					continue // 対象タスクは既に追加済み
				}
				ordered = append(ordered, task)
			}

			if err := rebalanceOrderKeys(tx, ordered); err != nil {
				return err
			}

			updated = ordered[0]
			return nil
		}

		// 通常の場合は order_key を更新するだけ
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

// MoveTask は指定されたタスクを要求された位置に移動します。
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
		// 排他ロックをかけて未完了タスクを取得
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
		targetIndex := clamp(targetPos-1, 0, n-1) // 範囲外の場合は端にクランプ
		target := tasks[oldIdx]

		// 同じ位置の場合は updated_at を更新するだけ
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

		// 対象タスクを除いたリストを作成
		remaining := make([]Task, 0, n-1)
		for i, task := range tasks {
			if i == oldIdx {
				continue
			}
			remaining = append(remaining, task)
		}

		// 新しい位置に挿入
		insertIdx := targetIndex
		if insertIdx < 0 {
			insertIdx = 0
		}
		if insertIdx > len(remaining) {
			insertIdx = len(remaining)
		}

		// 再配置後の順序を構築
		ordered := make([]Task, 0, n)
		ordered = append(ordered, remaining[:insertIdx]...)
		ordered = append(ordered, target)
		ordered = append(ordered, remaining[insertIdx:]...)
		finalIdx := insertIdx

		// 前後のタスクを取得（order_key を計算するため）
		prev := (*Task)(nil)
		next := (*Task)(nil)
		if finalIdx > 0 {
			prev = &ordered[finalIdx-1]
		}
		if finalIdx+1 < len(ordered) {
			next = &ordered[finalIdx+1]
		}

		// 新しい order_key を計算
		var newOrder float64
		switch {
		case prev == nil && next == nil:
			// タスクが1件だけの場合
			newOrder = 1.0
		case prev == nil:
			// 先頭に移動する場合
			newOrder = next.OrderKey - 1.0
		case next == nil:
			// 末尾に移動する場合
			newOrder = prev.OrderKey + 1.0
		default:
			// 中間に移動する場合：前後の order_key の平均値を使用
			gap := math.Abs(prev.OrderKey - next.OrderKey)
			// 差が小さすぎる場合は再調整が必要
			if gap < normalizeThreshold {
				if err := rebalanceOrderKeys(tx, ordered); err != nil {
					return err
				}
				return tx.Where("id = ?", target.ID).First(&updated).Error
			}
			newOrder = (prev.OrderKey + next.OrderKey) / 2.0
		}

		// order_key を更新
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

// EditTask は指定されたタスクの本文を更新します。
func (s *Store) EditTask(ctx context.Context, displayID int, text string) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}
	// 前後の空白を削除し、空文字チェック
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, ErrEmptyText
	}

	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		// 排他ロックをかけて未完了タスクを取得
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

		// 対象タスクを取得して本文を更新
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

// DeleteTask は指定された未完了タスクを物理的に削除します。
func (s *Store) DeleteTask(ctx context.Context, displayID int) (*Task, error) {
	if displayID < 1 {
		return nil, ErrInvalidDisplayID
	}

	db := s.db.WithContext(ctx)
	var deleted Task
	err := db.Transaction(func(tx *gorm.DB) error {
		// 排他ロックをかけて未完了タスクを取得
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

		// 対象タスクを削除
		target := tasks[displayID-1]
		result := tx.Delete(&Task{}, target.ID)
		if result.Error != nil {
			return fmt.Errorf("delete task: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("delete task: no rows affected")
		}

		deleted = target
		return nil
	})
	if err != nil {
		return nil, mapStorageError(err)
	}

	return &deleted, nil
}

// HeadTask は未完了タスクの先頭（最小の order_key）を取得します。
func (s *Store) HeadTask(ctx context.Context) (*Task, error) {
	var task Task
	if err := s.db.WithContext(ctx).
		Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Limit(1).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoPendingTasks
		}
		return nil, fmt.Errorf("select head task: %w", err)
	}

	return &task, nil
}

// ListTasks は未完了タスクと完了済みタスクを取得します。
func (s *Store) ListTasks(ctx context.Context) (pending []Task, completed []Task, err error) {
	db := s.db.WithContext(ctx)

	// 未完了タスクを order_key の昇順で取得
	if err = db.Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Find(&pending).Error; err != nil {
		return nil, nil, fmt.Errorf("list pending tasks: %w", err)
	}

	// 完了済みタスクを更新日時の降順で最大50件取得
	if err = db.Where("is_done = ?", true).
		Order("updated_at desc, id desc").
		Limit(50).
		Find(&completed).Error; err != nil {
		return nil, nil, fmt.Errorf("list completed tasks: %w", err)
	}

	return pending, completed, nil
}

// ListAllCompletedTasks は全ての完了済みタスクを取得します。
func (s *Store) ListAllCompletedTasks(ctx context.Context) ([]Task, error) {
	var tasks []Task
	if err := s.db.WithContext(ctx).
		Where("is_done = ?", true).
		Order("updated_at desc, id desc").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list all completed tasks: %w", err)
	}
	return tasks, nil
}

// DeleteCompletedTasksByDate は指定された日付の完了済みタスクを削除します。
// dateは YYYY/MM/DD 形式の文字列で、時刻部分は無視されます。
func (s *Store) DeleteCompletedTasksByDate(ctx context.Context, date string) (int, error) {
	// 日付文字列をパース（YYYY/MM/DD形式）
	dateLayout := "2006/01/02"
	targetDate, err := time.Parse(dateLayout, date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w (expected YYYY/MM/DD)", err)
	}

	// 日付の開始時刻と終了時刻を計算（UTC）
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	var deletedCount int64
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 指定された日付の完了済みタスクを削除
		result := tx.Where("is_done = ? AND updated_at >= ? AND updated_at < ?", true, startOfDay, endOfDay).
			Delete(&Task{})
		if result.Error != nil {
			return fmt.Errorf("delete completed tasks by date: %w", result.Error)
		}
		deletedCount = result.RowsAffected
		return nil
	})

	if err != nil {
		return 0, err
	}

	return int(deletedCount), nil
}

// CompleteNext は未完了タスクの先頭（最小の order_key）を完了状態にします。
func (s *Store) CompleteNext(ctx context.Context) (*Task, error) {
	db := s.db.WithContext(ctx)
	var updated Task
	err := db.Transaction(func(tx *gorm.DB) error {
		// 未完了タスクの先頭を取得（order_key が最小のもの）
		var task Task
		if err := tx.Where("is_done = ?", false).
			Order("order_key asc, id asc").
			First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNoPendingTasks
			}
			return fmt.Errorf("select next task: %w", err)
		}

		// タスクを完了状態に更新
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

// applyPragmas はSQLiteの設定を適用します。
func applyPragmas(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("access sql.DB: %w", err)
	}

	// SQLiteの設定を適用
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",   // Write-Ahead Logging（パフォーマンス向上）
		"PRAGMA synchronous = NORMAL;", // 同期モード（バランスの取れた設定）
		"PRAGMA foreign_keys = ON;",    // 外部キー制約を有効化
		"PRAGMA busy_timeout = 4000;",  // ロック待機タイムアウト（4秒）
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return fmt.Errorf("apply %s: %w", pragma, err)
		}
	}

	// 接続プールの設定（SQLiteは単一接続推奨）
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(0)

	return nil
}

// ensureDBPermissions はデータベースファイルのパーミッションを設定します。
func ensureDBPermissions(path string) error {
	// Windowsではパーミッション設定が不要
	if runtime.GOOS == "windows" {
		return nil
	}

	// ファイルが存在しない場合は作成
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("create database file: %w", err)
	}
	if closeErr := file.Close(); closeErr != nil {
		return fmt.Errorf("close database file: %w", closeErr)
	}

	// パーミッションを600（所有者のみ読み書き可能）に設定
	if err := os.Chmod(path, 0o600); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			if errno, ok := pathErr.Err.(syscall.Errno); ok {
				// パーミッション設定ができない環境（EPERM, ENOTSUP）の場合は644にフォールバック
				switch errno {
				case syscall.EPERM, syscall.ENOTSUP:
					if chmodErr := os.Chmod(path, 0o644); chmodErr != nil {
						return fmt.Errorf("set database permissions: %w", chmodErr)
					}
					return nil
				}
			}
		}
		return fmt.Errorf("set database permissions: %w", err)
	}

	return nil
}

const (
	// normalizeThreshold は order_key の再調整が必要と判断する閾値です。
	// 前後の order_key の差がこの値より小さい場合、再調整が行われます。
	normalizeThreshold = 1e-3

	// focusOrderStep は focus コマンドで order_key を減らす際のステップ値です。
	focusOrderStep = 1.0

	// focusOrderFloor は order_key の下限値です。
	// この値を下回る場合は再調整が必要と判断されます。
	focusOrderFloor = -1e9

	// focusNormalizeEpsilon は focus コマンドで再調整が必要か判断する際の閾値です。
	focusNormalizeEpsilon = 1e-6
)

// loadPendingTasksForUpdate は排他ロックをかけて未完了タスクを取得します。
func loadPendingTasksForUpdate(tx *gorm.DB) ([]Task, error) {
	var tasks []Task
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}). // SELECT FOR UPDATE（排他ロック）
		Where("is_done = ?", false).
		Order("order_key asc, id asc").
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("list pending tasks: %w", err)
	}

	return tasks, nil
}

// mapStorageError はストレージ層のエラーを適切なエラー型にマッピングします。
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

// clamp は値を指定された範囲内に制限します。
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// rebalanceOrderKeys はタスクの order_key を1, 2, 3...と連番で再計算します。
func rebalanceOrderKeys(tx *gorm.DB, ordered []Task) error {
	now := time.Now().UTC()
	for idx, task := range ordered {
		// 1から始まる連番を order_key として設定
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
		// スライス内のタスクも更新（参照を保持するため）
		if task.ID == ordered[idx].ID {
			ordered[idx].OrderKey = order
			ordered[idx].UpdatedAt = now
		}
	}

	return nil
}
