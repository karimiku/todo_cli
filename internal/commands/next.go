package commands

import (
	"fmt"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// NextCommand marks the top pending task as complete.
type NextCommand struct{}

// NewNextCommand constructs the next command.
func NewNextCommand() cli.Command {
	return &NextCommand{}
}

func (c *NextCommand) Name() string {
	return "next"
}

func (c *NextCommand) Aliases() []string {
	return nil
}

func (c *NextCommand) Description() string {
	return "先頭のタスクを完了します"
}

func (c *NextCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("next: 追加の引数は不要です")
	}

	task, err := ctx.Store.CompleteNext(ctx)
	if err != nil {
		if err == storage.ErrNoPendingTasks {
			fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
			return nil
		}
		return fmt.Errorf("完了処理に失敗しました: %w", err)
	}

	fmt.Fprintf(ctx.Stdout, "✓ 完了: %s\n", task.Text)
	return nil
}
