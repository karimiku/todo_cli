package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenInitialisesDatabase(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "todo.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})

	if !store.DB().Migrator().HasTable(&Task{}) {
		t.Fatalf("expected tasks table to exist")
	}

	var mode string
	if err := store.DB().Raw("PRAGMA journal_mode").Scan(&mode).Error; err != nil {
		t.Fatalf("failed to read journal_mode pragma: %v", err)
	}

	if strings.ToLower(mode) != "wal" {
		t.Fatalf("journal_mode = %q, want %q", mode, "wal")
	}
}
