package storage_test

import (
	"path/filepath"
	"strings"
	"testing"

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

func TestAddAndListTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(t.Context(), "メール返信"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.AddTask(t.Context(), "図書館に本返却"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	pending, completed, err := st.ListTasks(t.Context())
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

	first, err := st.AddTask(t.Context(), "メール返信")
	if err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}

	if _, err := st.AddTask(t.Context(), "図書館に本返却"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}

	task, err := st.CompleteNext(t.Context())
	if err != nil {
		t.Fatalf("CompleteNext: %v", err)
	}

	if task.ID != first.ID {
		t.Fatalf("completed unexpected task: got %d want %d", task.ID, first.ID)
	}

	if !task.IsDone {
		t.Fatalf("completed task must be marked done")
	}

	pending, completed, err := st.ListTasks(t.Context())
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

	if _, err := st.AddTask(t.Context(), " "); err == nil {
		t.Fatalf("expected error for empty text")
	} else if err != storage.ErrEmptyText {
		t.Fatalf("expected ErrEmptyText, got %v", err)
	}
}

func TestCompleteNextNoTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.CompleteNext(t.Context()); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}
}

func TestFocusTask(t *testing.T) {
	st := newStore(t)

	// Add three tasks
	if _, err := st.AddTask(t.Context(), "Task 1"); err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 2"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 3"); err != nil {
		t.Fatalf("AddTask #3: %v", err)
	}

	// Focus task 2 (should move it to position 1)
	task, err := st.FocusTask(t.Context(), 2)
	if err != nil {
		t.Fatalf("FocusTask: %v", err)
	}

	if task.Text != "Task 2" {
		t.Fatalf("focused wrong task: got %s, want Task 2", task.Text)
	}

	// Verify order
	pending, _, err := st.ListTasks(t.Context())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(pending) != 3 {
		t.Fatalf("expected 3 pending tasks, got %d", len(pending))
	}

	if pending[0].Text != "Task 2" || pending[1].Text != "Task 1" || pending[2].Text != "Task 3" {
		t.Fatalf("unexpected task order after focus: %v, %v, %v", pending[0].Text, pending[1].Text, pending[2].Text)
	}
}

func TestFocusTaskOutOfRange(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(t.Context(), "Task 1"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.FocusTask(t.Context(), 5); err == nil {
		t.Fatalf("expected error for out of range ID")
	}
}

func TestFocusTaskNoTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.FocusTask(t.Context(), 1); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}
}

func TestMoveTask(t *testing.T) {
	st := newStore(t)

	// Add three tasks
	if _, err := st.AddTask(t.Context(), "Task 1"); err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 2"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 3"); err != nil {
		t.Fatalf("AddTask #3: %v", err)
	}

	// Move task 3 to position 1
	task, err := st.MoveTask(t.Context(), 3, 1)
	if err != nil {
		t.Fatalf("MoveTask: %v", err)
	}

	if task.Text != "Task 3" {
		t.Fatalf("moved wrong task: got %s, want Task 3", task.Text)
	}

	// Verify order
	pending, _, err := st.ListTasks(t.Context())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(pending) != 3 {
		t.Fatalf("expected 3 pending tasks, got %d", len(pending))
	}

	if pending[0].Text != "Task 3" || pending[1].Text != "Task 1" || pending[2].Text != "Task 2" {
		t.Fatalf("unexpected task order after move: %v, %v, %v", pending[0].Text, pending[1].Text, pending[2].Text)
	}
}

func TestMoveTaskInvalidPosition(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(t.Context(), "Task 1"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.MoveTask(t.Context(), 1, 0); err != storage.ErrInvalidPosition {
		t.Fatalf("expected ErrInvalidPosition, got %v", err)
	}
}

func TestEditTask(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(t.Context(), "Original text"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	task, err := st.EditTask(t.Context(), 1, "Updated text")
	if err != nil {
		t.Fatalf("EditTask: %v", err)
	}

	if task.Text != "Updated text" {
		t.Fatalf("task text not updated: got %s, want Updated text", task.Text)
	}

	// Verify via list
	pending, _, err := st.ListTasks(t.Context())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if pending[0].Text != "Updated text" {
		t.Fatalf("task text not persisted: got %s", pending[0].Text)
	}
}

func TestEditTaskEmptyText(t *testing.T) {
	st := newStore(t)

	if _, err := st.AddTask(t.Context(), "Original"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := st.EditTask(t.Context(), 1, "  "); err != storage.ErrEmptyText {
		t.Fatalf("expected ErrEmptyText, got %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	st := newStore(t)

	// Add three tasks
	if _, err := st.AddTask(t.Context(), "Task 1"); err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 2"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}
	if _, err := st.AddTask(t.Context(), "Task 3"); err != nil {
		t.Fatalf("AddTask #3: %v", err)
	}

	// Delete task 2
	task, err := st.DeleteTask(t.Context(), 2)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	if task.Text != "Task 2" {
		t.Fatalf("deleted wrong task: got %s, want Task 2", task.Text)
	}

	// Verify only 2 tasks remain
	pending, _, err := st.ListTasks(t.Context())
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(pending) != 2 {
		t.Fatalf("expected 2 pending tasks after delete, got %d", len(pending))
	}

	if pending[0].Text != "Task 1" || pending[1].Text != "Task 3" {
		t.Fatalf("unexpected tasks after delete: %v, %v", pending[0].Text, pending[1].Text)
	}
}

func TestDeleteTaskNoTasks(t *testing.T) {
	st := newStore(t)

	if _, err := st.DeleteTask(t.Context(), 1); err != storage.ErrNoPendingTasks {
		t.Fatalf("expected ErrNoPendingTasks, got %v", err)
	}
}
