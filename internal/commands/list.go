package commands

import (
	"fmt"
	"sort"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// ListCommand は未完了タスクと完了済みタスクを表示するコマンドです。
type ListCommand struct{}

// NewListCommand は list コマンドのインスタンスを生成します。
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

// Run は未完了タスクと完了済みタスクを表示します。
func (c *ListCommand) Run(ctx *cli.CommandContext, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("追加の引数は不要です")
	}

	pending, completed, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
	}

	// タスクが1件もない場合
	if len(pending) == 0 && len(completed) == 0 {
		fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
		return nil
	}

	// 未完了タスクを番号付きで表示（1から始まる）
	if len(pending) == 0 {
		fmt.Fprintln(ctx.Stdout, storage.ErrNoPendingTasks.Error())
	} else {
		for idx, task := range pending {
			fmt.Fprintf(ctx.Stdout, "%d %s\n", idx+1, task.Text)
		}
	}

	// 完了済みタスクがあれば区切り線の後に日付ごとにグルーピングして表示
	if len(completed) > 0 {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, "---")
		fmt.Fprintln(ctx.Stdout)

		// 日付ごとにグルーピング
		grouped := groupTasksByDate(completed)

		// 日付を降順でソート（新しい日付が先頭）
		dates := make([]string, 0, len(grouped))
		for date := range grouped {
			dates = append(dates, date)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(dates)))

		// 日付ごとに表示
		for _, date := range dates {
			tasksForDate := grouped[date]
			fmt.Fprintf(ctx.Stdout, "%s (%d件)\n", date, len(tasksForDate))
			for _, task := range tasksForDate {
				fmt.Fprintf(ctx.Stdout, "  [✅] %s\n", task.Text)
			}
			fmt.Fprintln(ctx.Stdout)
		}
	}

	return nil
}
