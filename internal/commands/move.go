package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// MoveCommand は指定されたタスクを任意の位置に移動するコマンドです。
type MoveCommand struct{}

// NewMoveCommand は move コマンドのインスタンスを生成します。
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
	return "タスクの順番を入れ替えます"
}

// Run は指定されたタスクを要求された位置に移動します。
func (c *MoveCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: %s move <id> <pos>", ctx.BinaryName)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		return fmt.Errorf("ID は 1 以上の整数で指定してください")
	}

	pos, err := strconv.Atoi(args[1])
	if err != nil || pos < 1 {
		return fmt.Errorf("順位は 1 以上を指定してください")
	}

	task, err := ctx.Store.MoveTask(ctx, id, pos)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidPosition):
			return fmt.Errorf("順位は 1 以上を指定してください")
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

	// 移動後の実際の位置を取得して表示
	position, err := currentPosition(ctx, task.ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.Stdout, "⇄ 順番変更: %s → %d\n", task.Text, position)
	return nil
}
