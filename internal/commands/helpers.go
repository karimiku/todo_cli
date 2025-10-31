package commands

import (
	"fmt"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// notifyPendingCount は現在の未完了タスク数をユーザーに通知します。
func notifyPendingCount(ctx *cli.CommandContext) error {
	count, err := pendingCount(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
		return nil
	}

	fmt.Fprintf(ctx.Stdout, "未完了の件数は %d 件です\n", count)
	return nil
}

// pendingCount は現在の未完了タスク数を取得します。
func pendingCount(ctx *cli.CommandContext) (int, error) {
	pending, _, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return 0, fmt.Errorf("タスク件数の取得に失敗しました: %w", err)
	}

	return len(pending), nil
}

// currentPosition は指定されたタスクIDの現在の表示位置（1から始まる）を取得します。
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
