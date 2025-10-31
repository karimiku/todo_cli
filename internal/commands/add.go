package commands

import (
	"fmt"
	"strings"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// AddCommand は新しいタスクを未完了リストの末尾に追加するコマンドです。
type AddCommand struct{}

// NewAddCommand は add コマンドのインスタンスを生成します。
func NewAddCommand() cli.Command {
	return &AddCommand{}
}

func (c *AddCommand) Name() string {
	return "add"
}

func (c *AddCommand) Aliases() []string {
	return nil
}

func (c *AddCommand) Description() string {
	return "未完了タスクを末尾へ追加します"
}

// Run は新しいタスクを未完了リストの末尾に追加します。
func (c *AddCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s add <text>", ctx.BinaryName)
	}

	// 複数の引数は空白で結合して1つのテキストとして扱う
	text := strings.Join(args, " ")
	task, err := ctx.Store.AddTask(ctx, text)
	if err != nil {
		if err == storage.ErrEmptyText {
			return err
		}
		return fmt.Errorf("タスク追加に失敗しました: %w", err)
	}

	fmt.Fprintf(ctx.Stdout, "追加: %s\n", task.Text)
	return nil
}
