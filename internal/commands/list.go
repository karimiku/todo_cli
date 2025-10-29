package commands

import (
	"fmt"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// ListCommand prints pending and completed tasks.
type ListCommand struct{}

// NewListCommand constructs the list command.
func NewListCommand() cli.Command {
	return &ListCommand{}
}

func (c *ListCommand) Name() string {
	return "list"
}

func (c *ListCommand) Aliases() []string {
	return nil
}

func (c *ListCommand) Description() string {
	return "未完了タスクを順番に表示します"
}

func (c *ListCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("追加の引数は不要です")
	}

	pending, completed, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
	}

	if len(pending) == 0 && len(completed) == 0 {
		fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
		return nil
	}

	if len(pending) == 0 {
		fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
	} else {
		for idx, task := range pending {
			fmt.Fprintf(ctx.Stdout, "%d %s\n", idx+1, task.Text)
		}
	}

	if len(completed) > 0 {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, "---")
		fmt.Fprintln(ctx.Stdout)
		for _, task := range completed {
			fmt.Fprintf(ctx.Stdout, "[✅] %s\n", task.Text)
		}
	}

	return nil
}
