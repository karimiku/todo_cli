package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// EditCommand updates the text of a task.
type EditCommand struct{}

// NewEditCommand constructs the edit command.
func NewEditCommand() cli.Command {
	return &EditCommand{}
}

func (c *EditCommand) Name() string {
	return "edit"
}

func (c *EditCommand) Aliases() []string {
	return nil
}

func (c *EditCommand) Description() string {
	return "指定したタスクの文言を編集します"
}

func (c *EditCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: todo edit <id> <text>")
	}

	displayID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("IDは数値で指定してください")
	}

	text := strings.Join(args[1:], " ")
	task, err := ctx.Store.EditTask(ctx, displayID, text)
	if err != nil {
		if err == storage.ErrEmptyText {
			return err
		}
		return err
	}

	fmt.Fprintf(ctx.Stdout, "✎ 更新: %s\n", task.Text)
	return nil
}
