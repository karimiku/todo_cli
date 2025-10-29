package commands

import (
	"fmt"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

func notifyPendingCount(ctx *cli.CommandContext) error {
	count, err := pendingCount(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "未完了の件数は %d 件です\n", count)
	return nil
}

func pendingCount(ctx *cli.CommandContext) (int, error) {
	pending, _, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("タスク件数の取得に失敗しました: %w", err)
	}

	return len(pending), nil
}

func currentPosition(ctx *cli.CommandContext, id uint) (int, error) {
	pending, _, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
	}

	for idx, task := range pending {
		if task.ID == id {
			return idx + 1, nil
		}
	}

	return 0, storage.ErrInvalidDisplayID
}
