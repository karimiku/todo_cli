package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kamiriku/todo_cli/internal/storage"
)

func newStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "todo.db")
	st, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Fatalf("store.Close: %v", err)
		}
	})
	return st
}

func TestOpenInitialisesDatabase(t *testing.T) {
	t.Parallel()

	st := newStore(t)

	if !st.DB().Migrator().HasTable(&storage.Task{}) {
		t.Fatalf("expected tasks table to exist")
	}

	var mode string
	if err := st.DB().Raw("PRAGMA journal_mode").Scan(&mode).Error; err != nil {
		t.Fatalf("failed to read journal_mode pragma: %v", err)
	}

	if strings.ToLower(mode) != "wal" {
		t.Fatalf("journal_mode = %q, want %q", mode, "wal")
	}
}

func TestOpenSetsDBPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions are not enforced on Windows")
	}

	temp := t.TempDir()
	dbPath := filepath.Join(temp, "todo.db")
	st, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Fatalf("store.Close: %v", err)
		}
	})

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("os.Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 && perm != 0o644 {
		t.Fatalf("unexpected permissions: %#o", perm)
	}
}

func TestAddAndListTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(context.Background(), "メール返信"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.AddTask(context.Background(), "図書館に本返却"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	pending, completed, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(pending) != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", len(pending))
	}

	if pending[0].Text != "メール返信" || pending[1].Text != "図書館に本返却" {
		t.Fatalf("unexpected task order: %+v", pending)
	}

	if len(completed) != 0 {
		t.Fatalf("expected 0 completed tasks, got %d", len(completed))
	}
}

func TestCompleteNext(t *testing.T) {
	st := newStore(t)

	first, err := st.AddTask(context.Background(), "メール返信")
	if err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}

	if _, err := st.AddTask(context.Background(), "図書館に本返却"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}

	task, err := st.CompleteNext(context.Background())
	if err != nil {
		t.Fatalf("CompleteNext: %v", err)
	}

	if task.ID != first.ID {
		t.Fatalf("completed unexpected task: got %d want %d", task.ID, first.ID)
	}

	if !task.IsDone {
		t.Fatalf("completed task must be marked done")
	}

	pending, completed, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(pending) != 1 {
		t.Fatalf("expected 1 pending task, got %d", len(pending))
	}

	if len(completed) != 1 {
		t.Fatalf("expected 1 completed task, got %d", len(completed))
	}
}

func TestAddTaskEmptyText(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(context.Background(), " "); err == nil {
		t.Fatalf("expected error for empty text")
	} else if err != storage.ErrEmptyText {
		t.Fatalf("expected ErrEmptyText, got %v", err)
	}
}

func TestCompleteNextNoTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.CompleteNext(context.Background()); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}
}

func TestFocusTaskMovesToTop(t *testing.T) {
	st := newStore(t)

	first, err := st.AddTask(context.Background(), "朝会の準備")
	if err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := st.AddTask(context.Background(), "メール返信"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}
	third, err := st.AddTask(context.Background(), "資料確認")
	if err != nil {
		t.Fatalf("AddTask #3: %v", err)
	}

	focused, err := st.FocusTask(context.Background(), 3)
	if err != nil {
		t.Fatalf("FocusTask: %v", err)
	}
	if focused.ID != third.ID {
		t.Fatalf("focused task mismatch: got %d want %d", focused.ID, third.ID)
	}

	pending, _, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if pending[0].ID != third.ID {
		t.Fatalf("expected task %d to be first, got %d", third.ID, pending[0].ID)
	}
	if pending[1].ID != first.ID {
		t.Fatalf("expected original first task to shift, got %d", pending[1].ID)
	}
	if pending[0].OrderKey >= pending[1].OrderKey {
		t.Fatalf("expected focused order_key < next, got %f vs %f", pending[0].OrderKey, pending[1].OrderKey)
	}
}

func TestFocusTaskTriggersNormalizeWhenOrderFloorExceeded(t *testing.T) {
	st := newStore(t)

	first, err := st.AddTask(context.Background(), "朝会の準備")
	if err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	second, err := st.AddTask(context.Background(), "レビュー")
	if err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}

	if err := st.DB().Model(&storage.Task{}).
		Where("id = ?", first.ID).
		Update("order_key", -1e9-0.5).Error; err != nil {
		t.Fatalf("prepare extreme order: %v", err)
	}

	focused, err := st.FocusTask(context.Background(), 2)
	if err != nil {
		t.Fatalf("FocusTask: %v", err)
	}
	if focused.ID != second.ID {
		t.Fatalf("expected second task to be focused, got %d", focused.ID)
	}

	pending, _, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if pending[0].ID != second.ID {
		t.Fatalf("expected second task to be first after normalize, got %d", pending[0].ID)
	}
	if pending[0].OrderKey != 1.0 || pending[1].OrderKey != 2.0 {
		t.Fatalf("expected normalized order keys 1.0/2.0, got %f/%f", pending[0].OrderKey, pending[1].OrderKey)
	}
}

func TestFocusTaskInvalidID(t *testing.T) {
	st := newStore(t)
	if _, err := st.FocusTask(context.Background(), 1); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}

	if _, err := st.AddTask(context.Background(), "朝会の準備"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if _, err := st.FocusTask(context.Background(), 2); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID, got %v", err)
	}
}

func TestFocusTaskKeepsOrderWhenAlreadyTop(t *testing.T) {
	st := newStore(t)
	task, err := st.AddTask(context.Background(), "朝会")
	if err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	origOrder := task.OrderKey

	focused, err := st.FocusTask(context.Background(), 1)
	if err != nil {
		t.Fatalf("FocusTask: %v", err)
	}
	if focused.OrderKey != origOrder {
		t.Fatalf("expected order_key to stay %f, got %f", origOrder, focused.OrderKey)
	}
}

func TestMoveTaskReorders(t *testing.T) {
	st := newStore(t)
	first, err := st.AddTask(context.Background(), "朝会準備")
	if err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	second, err := st.AddTask(context.Background(), "メール")
	if err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}
	third, err := st.AddTask(context.Background(), "資料確認")
	if err != nil {
		t.Fatalf("AddTask #3: %v", err)
	}

	moved, err := st.MoveTask(context.Background(), 3, 1)
	if err != nil {
		t.Fatalf("MoveTask: %v", err)
	}
	if moved.ID != third.ID {
		t.Fatalf("moved unexpected task: got %d want %d", moved.ID, third.ID)
	}

	pending, _, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	expected := []uint{third.ID, first.ID, second.ID}
	for i, task := range pending {
		if task.ID != expected[i] {
			t.Fatalf("unexpected order at %d: got %d want %d", i, task.ID, expected[i])
		}
	}

	if _, err := st.MoveTask(context.Background(), 1, 3); err != nil {
		t.Fatalf("MoveTask to tail: %v", err)
	}
	pending, _, err = st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	expected = []uint{first.ID, second.ID, third.ID}
	for i, task := range pending {
		if task.ID != expected[i] {
			t.Fatalf("unexpected order after tail move at %d: got %d want %d", i, task.ID, expected[i])
		}
	}
}

func TestMoveTaskValidation(t *testing.T) {
	st := newStore(t)
	if _, err := st.MoveTask(context.Background(), 1, 1); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}

	if _, err := st.AddTask(context.Background(), "朝会"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.MoveTask(context.Background(), 0, 1); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID for id=0, got %v", err)
	}

	if _, err := st.MoveTask(context.Background(), 1, 0); err != storage.ErrInvalidPosition {
		t.Fatalf("expected ErrInvalidPosition for pos=0, got %v", err)
	}

	if _, err := st.MoveTask(context.Background(), 2, 1); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID for id out of range, got %v", err)
	}
}

func TestEditTaskUpdatesText(t *testing.T) {
	st := newStore(t)
	if _, err := st.AddTask(context.Background(), "旧テキスト"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	before := time.Now()
	edited, err := st.EditTask(context.Background(), 1, "新しいテキスト")
	if err != nil {
		t.Fatalf("EditTask: %v", err)
	}
	if edited.Text != "新しいテキスト" {
		t.Fatalf("text not updated: %q", edited.Text)
	}
	if !edited.UpdatedAt.After(before) {
		t.Fatalf("expected updated_at to be refreshed")
	}

	pending, _, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if pending[0].Text != "新しいテキスト" {
		t.Fatalf("list should reflect new text, got %q", pending[0].Text)
	}
}

func TestEditTaskValidation(t *testing.T) {
	st := newStore(t)
	if _, err := st.EditTask(context.Background(), 1, "更新"); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}

	if _, err := st.AddTask(context.Background(), "テスト"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.EditTask(context.Background(), 2, "更新"); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID, got %v", err)
	}

	if _, err := st.EditTask(context.Background(), 1, " "); err != storage.ErrEmptyText {
		t.Fatalf("expected ErrEmptyText, got %v", err)
	}
}

func TestDeleteTaskRemovesPending(t *testing.T) {
	st := newStore(t)
	if _, err := st.AddTask(context.Background(), "A"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if _, err := st.AddTask(context.Background(), "B"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if _, err := st.AddTask(context.Background(), "C"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	deleted, err := st.DeleteTask(context.Background(), 2)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if deleted.Text != "B" {
		t.Fatalf("expected to delete second task, got %s", deleted.Text)
	}

	pending, _, err := st.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", len(pending))
	}
	if pending[0].Text != "A" || pending[1].Text != "C" {
		t.Fatalf("unexpected tasks remaining: %+v", pending)
	}
}

func TestDeleteTaskValidation(t *testing.T) {
	st := newStore(t)
	if _, err := st.DeleteTask(context.Background(), 1); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}

	if _, err := st.AddTask(context.Background(), "タスク"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.DeleteTask(context.Background(), 0); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID for id=0, got %v", err)
	}
	if _, err := st.DeleteTask(context.Background(), 2); err != storage.ErrInvalidDisplayID {
		t.Fatalf("expected ErrInvalidDisplayID for id out of range, got %v", err)
	}
}

func TestHeadTaskReturnsFirstPending(t *testing.T) {
	st := newStore(t)
	if _, err := st.AddTask(context.Background(), "A"); err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := st.AddTask(context.Background(), "B"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}

	first, err := st.HeadTask(context.Background())
	if err != nil {
		t.Fatalf("HeadTask: %v", err)
	}
	if first.Text != "A" {
		t.Fatalf("expected first task to be A, got %s", first.Text)
	}

	if _, err := st.CompleteNext(context.Background()); err != nil {
		t.Fatalf("CompleteNext: %v", err)
	}

	second, err := st.HeadTask(context.Background())
	if err != nil {
		t.Fatalf("HeadTask after next: %v", err)
	}
	if second.Text != "B" {
		t.Fatalf("expected second task to be B, got %s", second.Text)
	}

	if _, err := st.CompleteNext(context.Background()); err != nil {
		t.Fatalf("CompleteNext #2: %v", err)
	}

	if _, err := st.HeadTask(context.Background()); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}
}
