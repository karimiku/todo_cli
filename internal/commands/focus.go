package commands

import (
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
)

// FocusCommand moves a task to the top position.
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
	return "指定したタスクを一番上へ移動します"
}

func (c *FocusCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: todo focus <id>")
	}

	displayID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("IDは数値で指定してください")
	}

	task, err := ctx.Store.FocusTask(ctx, displayID)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "⇡ 最優先に移動: %s\n", task.Text)
	return nil
}
