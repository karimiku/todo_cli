package commands

import (
	"fmt"
	"strings"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// AddCommand appends a new task to the queue.
type AddCommand struct{}

// NewAddCommand constructs the add command.
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

func (c *AddCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s add <text>", ctx.BinaryName)
	}

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
