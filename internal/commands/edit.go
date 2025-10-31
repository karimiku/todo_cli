package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// EditCommand は指定されたタスクの本文を更新するコマンドです。
type EditCommand struct{}

// NewEditCommand は edit コマンドのインスタンスを生成します。
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
	return "タスクの本文を更新します"
}

// Run は指定されたタスクの本文を更新します。
func (c *EditCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s edit <id> <text>", ctx.BinaryName)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		return fmt.Errorf("ID は 1 以上の整数で指定してください")
	}

	// 2番目以降の引数は空白で結合して1つのテキストとして扱う
	text := strings.Join(args[1:], " ")
	task, err := ctx.Store.EditTask(ctx, id, text)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEmptyText):
			return err
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
			return fmt.Errorf("タスクの編集に失敗しました: %w", err)
		}
	}

	fmt.Fprintf(ctx.Stdout, "✎ 更新: %s\n", task.Text)
	return nil
}
