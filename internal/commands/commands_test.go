package commands

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

type commandHarness struct {
	ctx    *cli.CommandContext
	store  *storage.Store
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func newHarness(t *testing.T) *commandHarness {
	t.Helper()

	// `t.TempDir()` でテスト用の一時ディレクトリを作成し、その中に "todo.db" というファイル名のパスを作成している。
	// これにより、テストごとに衝突しない一時データベースファイルを利用できるようにしている。
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

	return &commandHarness{
		ctx: &cli.CommandContext{
			Context:    context.Background(),
			Store:      store,
			Stdout:     stdout,
			Stderr:     stderr,
			BinaryName: "todo",
		},
		store:  store,
		stdout: stdout,
		stderr: stderr,
	}
}

func TestFocusCommandNoPendingShowsCelebration(t *testing.T) {
	h := newHarness(t)

	cmd := NewFocusCommand()
	if err := cmd.Run(h.ctx, []string{"1"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := h.stdout.String()
	want := "未完了はありません 🎉\n"
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestFocusCommandInvalidIDShowsPendingCount(t *testing.T) {
	h := newHarness(t)

	if _, err := h.store.AddTask(h.ctx, "A"); err != nil {
		t.Fatalf("AddTask #1: %v", err)
	}
	if _, err := h.store.AddTask(h.ctx, "B"); err != nil {
		t.Fatalf("AddTask #2: %v", err)
	}

	h.stdout.Reset()

	cmd := NewFocusCommand()
	if err := cmd.Run(h.ctx, []string{"3"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := h.stdout.String()
	want := "未完了の件数は 2 件です\n"
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestMoveCommandNoPendingShowsCelebration(t *testing.T) {
	h := newHarness(t)

	cmd := NewMoveCommand()
	if err := cmd.Run(h.ctx, []string{"1", "1"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := h.stdout.String()
	want := "未完了はありません 🎉\n"
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestDeleteCommandInvalidIDShowsPendingCount(t *testing.T) {
	h := newHarness(t)

	if _, err := h.store.AddTask(h.ctx, "朝会"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	h.stdout.Reset()

	cmd := NewDeleteCommand()
	if err := cmd.Run(h.ctx, []string{"2"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got := h.stdout.String()
	want := "未完了の件数は 1 件です\n"
	if got != want {
		t.Fatalf("unexpected output: got %q want %q", got, want)
	}
}

func TestListCommandShowsCompletedWithCheckMark(t *testing.T) {
	h := newHarness(t)

	if _, err := h.store.AddTask(h.ctx, "朝会"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}
	if _, err := h.store.AddTask(h.ctx, "資料作成"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	if _, err := h.store.CompleteNext(h.ctx); err != nil {
		t.Fatalf("CompleteNext: %v", err)
	}

	h.stdout.Reset()

	cmd := NewListCommand()
	if err := cmd.Run(h.ctx, nil); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !bytes.Contains(h.stdout.Bytes(), []byte("[✅] 朝会")) {
		t.Fatalf("expected completed task marker [✅], got %q", h.stdout.String())
	}
}

func TestBarCommandHeadOnlyOutputsTask(t *testing.T) {
	h := newHarness(t)
	if _, err := h.store.AddTask(h.ctx, "全角テストあいうえお"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	cmd := NewBarCommand()
	if err := cmd.Run(h.ctx, []string{"--head-only", "--maxlen", "4", "--icon", "★"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := h.stdout.String()
	if !strings.Contains(out, "bash=todo param1=next") {
		t.Fatalf("expected bash invocation, got %q", out)
	}
	if !strings.Contains(out, "★ 全角テス…") {
		t.Fatalf("expected truncated title with icon, got %q", out)
	}
}

func TestBarCommandNoPendingShowsEmptyMessage(t *testing.T) {
	h := newHarness(t)
	cmd := NewBarCommand()
	if err := cmd.Run(h.ctx, []string{"--head-only"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got := h.stdout.String(); got != "🎉 空っぽ！ | tooltip=おつかれさま\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestBarCommandNoTooltipOption(t *testing.T) {
	h := newHarness(t)
	if _, err := h.store.AddTask(h.ctx, "メール"); err != nil {
		t.Fatalf("AddTask: %v", err)
	}

	cmd := NewBarCommand()
	if err := cmd.Run(h.ctx, []string{"--head-only", "--no-tooltip"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := h.stdout.String()
	if strings.Contains(out, "tooltip=") {
		t.Fatalf("tooltip should be omitted, got %q", out)
	}
}

func TestBarCommandReportsDatabaseError(t *testing.T) {
	h := newHarness(t)
	if err := h.store.Close(); err != nil {
		t.Fatalf("store.Close: %v", err)
	}

	cmd := NewBarCommand()
	if err := cmd.Run(h.ctx, []string{"--head-only"}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	out := h.stdout.String()
	if !strings.Contains(out, "⚠︎ Error") {
		t.Fatalf("expected error marker, got %q", out)
	}
}
