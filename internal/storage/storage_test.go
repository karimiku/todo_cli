package storage_test

import (
	"path/filepath"
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
