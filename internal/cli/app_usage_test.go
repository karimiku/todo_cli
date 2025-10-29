package cli_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	cliPkg "github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/commands"
	"github.com/kamiriku/todo_cli/internal/storage"
)

func TestUsageLabelWithCustomAliases(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "todo.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("store.Close: %v", err)
		}
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	app := cliPkg.NewApp(store, stdout, stderr, "todo", []string{"todo", "tb", "td", "Todo"})
	if err := app.Register(commands.NewListCommand()); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if code := app.Run(context.Background(), []string{"help"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	out := stdout.String()
	if !bytes.Contains([]byte(out), []byte("usage: todo (tb/td)")) {
		t.Fatalf("usage should prefer canonical alias order, got %q", out)
	}
}
