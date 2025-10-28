package commands

import (
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
)

// MoveCommand moves a task to a specific position.
type MoveCommand struct{}

// NewMoveCommand constructs the move command.
func NewMoveCommand() cli.Command {
	return &MoveCommand{}
}

func (c *MoveCommand) Name() string {
	return "move"
}

func (c *MoveCommand) Aliases() []string {
	return nil
}

func (c *MoveCommand) Description() string {
	return "指定したタスクを任意の位置へ移動します"
}

func (c *MoveCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: todo move <id> <pos>")
	}

	displayID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("IDは数値で指定してください")
	}

	position, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("順位は数値で指定してください")
	}

	task, err := ctx.Store.MoveTask(ctx, displayID, position)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "⇄ 順番変更: %s → %d\n", task.Text, position)
	return nil
}
