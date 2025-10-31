package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// FocusCommand は指定されたタスクを未完了リストの最上位に移動するコマンドです。
type FocusCommand struct{}

// NewFocusCommand は focus コマンドのインスタンスを生成します。
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

// Run は指定されたタスクを未完了リストの最上位に移動します。
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
		case errors.Is(err, storage.ErrNoPendingTasks):
			fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
			return nil
		case errors.Is(err, storage.ErrInvalidDisplayID):
			// 無効なIDの場合は現在の未完了タスク数を表示
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
