package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// FocusCommand promotes a task to the top of the pending list.
type FocusCommand struct{}

// NewFocusCommand constructs the focus command.
func NewFocusCommand() cli.Command {
	return &FocusCommand{}
}

func (c *FocusCommand) Name() string {
	return "focus"
}

func (c *FocusCommand) Aliases() []string {
	return nil
}

func (c *FocusCommand) Description() string {
	return "指定番号のタスクを最優先にします"
}

func (c *FocusCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s focus <id>", ctx.BinaryName)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		return fmt.Errorf("ID は 1 以上の整数で指定してください")
	}

	task, err := ctx.Store.FocusTask(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNoPendingTasks), errors.Is(err, storage.ErrInvalidDisplayID):
			if notifyErr := notifyPendingCount(ctx); notifyErr != nil {
				return notifyErr
			}
			return nil
		default:
			return fmt.Errorf("タスクの移動に失敗しました: %w", err)
		}
	}

	fmt.Fprintf(ctx.Stdout, "⇡ 最優先に移動: %s\n", task.Text)
	return nil
}
