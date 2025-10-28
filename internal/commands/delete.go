package commands

import (
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
)

// DeleteCommand physically deletes a task.
type DeleteCommand struct{}

// NewDeleteCommand constructs the delete command.
func NewDeleteCommand() cli.Command {
	return &DeleteCommand{}
}

func (c *DeleteCommand) Name() string {
	return "delete"
}

func (c *DeleteCommand) Aliases() []string {
	return nil
}

func (c *DeleteCommand) Description() string {
	return "指定したタスクを削除します"
}

func (c *DeleteCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: todo delete <id>")
	}

	displayID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("IDは数値で指定してください")
	}

	task, err := ctx.Store.DeleteTask(ctx, displayID)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "🗑 削除: %s\n", task.Text)
	return nil
}
