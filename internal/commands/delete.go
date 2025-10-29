package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// DeleteCommand removes a pending task entirely.
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
	return "未完了タスクを削除します"
}

func (c *DeleteCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s delete <id>", ctx.BinaryName)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		return fmt.Errorf("ID は 1 以上の整数で指定してください")
	}

	task, err := ctx.Store.DeleteTask(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNoPendingTasks), errors.Is(err, storage.ErrInvalidDisplayID):
			if notifyErr := notifyPendingCount(ctx); notifyErr != nil {
				return notifyErr
			}
			return nil
		default:
			return fmt.Errorf("タスクの削除に失敗しました: %w", err)
		}
	}

	fmt.Fprintf(ctx.Stdout, "🗑 削除: %s\n", task.Text)
	return nil
}
