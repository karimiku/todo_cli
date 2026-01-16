package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// CleanCommand は完了済みタスクを日付ごとに表示・削除するコマンドです。
type CleanCommand struct{}

// NewCleanCommand は clean コマンドのインスタンスを生成します。
func NewCleanCommand() cli.Command {
	return &CleanCommand{}
}

func (c *CleanCommand) Name() string {
	return "clean"
}

func (c *CleanCommand) Aliases() []string {
	return nil
}

func (c *CleanCommand) Description() string {
	return "完了済みタスクを日付ごとに表示・削除します（clean all で全削除）"
}

// Run は完了済みタスクを日付ごとに表示し、日付が指定された場合は削除します。
func (c *CleanCommand) Run(ctx *cli.CommandContext, args []string) error {
	// 全ての完了済みタスクを取得
	tasks, err := ctx.Store.ListAllCompletedTasks(ctx)
	if err != nil {
		return fmt.Errorf("完了済みタスクの取得に失敗しました: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(ctx.Stdout, "完了済みタスクはありません")
		return nil
	}

	// 日付ごとにグルーピング
	grouped := groupTasksByDate(tasks)

	// 日付を降順でソート（新しい日付が先頭）
	dates := make([]string, 0, len(grouped))
	for date := range grouped {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// 引数が指定されている場合は削除を実行
	if len(args) > 0 {
		if len(args) != 1 {
			return fmt.Errorf("usage: %s clean [YYYY/MM/DD|all]", ctx.BinaryName)
		}

		arg := args[0]

		// "all"が指定された場合はすべての完了タスクを削除
		if arg == "all" {
			// 削除を実行
			count, err := ctx.Store.DeleteAllCompletedTasks(ctx)
			if err != nil {
				return fmt.Errorf("タスクの削除に失敗しました: %w", err)
			}

			if count == 0 {
				fmt.Fprintln(ctx.Stdout, "完了済みタスクはありません")
				return nil
			}

			fmt.Fprintf(ctx.Stdout, "🗑 すべての完了済みタスク %d 件を削除しました\n", count)
			return nil
		}

		// 日付フォーマットの検証
		date := arg
		if _, err := time.Parse("2006/01/02", date); err != nil {
			return fmt.Errorf("日付は YYYY/MM/DD 形式で指定してください（例: 2025/01/15）、または 'all' を指定してください")
		}

		// 指定された日付のタスクが存在するか確認
		if _, exists := grouped[date]; !exists {
			fmt.Fprintf(ctx.Stdout, "%s の完了済みタスクはありません\n", date)
			return nil
		}

		// 削除を実行
		count, err := ctx.Store.DeleteCompletedTasksByDate(ctx, date)
		if err != nil {
			return fmt.Errorf("タスクの削除に失敗しました: %w", err)
		}

		fmt.Fprintf(ctx.Stdout, "🗑 %s の完了済みタスク %d 件を削除しました\n", date, count)
		return nil
	}

	// 日付ごとに表示
	for _, date := range dates {
		tasksForDate := grouped[date]
		fmt.Fprintf(ctx.Stdout, "\n%s (%d件)\n", date, len(tasksForDate))
		fmt.Fprintln(ctx.Stdout, strings.Repeat("-", len(date)+10))
		for _, task := range tasksForDate {
			fmt.Fprintf(ctx.Stdout, "  [✅] %s\n", task.Text)
		}
	}
	fmt.Fprintln(ctx.Stdout, "\n削除する場合は: clean YYYY/MM/DD または clean all")

	return nil
}

// groupTasksByDate はタスクを日付（YYYY/MM/DD形式）でグルーピングします。
func groupTasksByDate(tasks []storage.Task) map[string][]storage.Task {
	grouped := make(map[string][]storage.Task)

	for _, task := range tasks {
		// updated_at を日付文字列に変換（YYYY/MM/DD形式）
		dateStr := task.UpdatedAt.Format("2006/01/02")
		grouped[dateStr] = append(grouped[dateStr], task)
	}

	// 各日付のタスクを updated_at の降順でソート
	for date := range grouped {
		tasks := grouped[date]
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		})
		grouped[date] = tasks
	}

	return grouped
}
