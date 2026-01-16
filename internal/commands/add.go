package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// AddCommand は新しいタスクを未完了リストの末尾に追加するコマンドです。
type AddCommand struct{}

// AddFlags は add コマンドのフラグを管理します。
type AddFlags struct {
	Quiet  bool
	Urgent bool
}

// NewAddCommand は add コマンドのインスタンスを生成します。
func NewAddCommand() cli.Command {
	return &AddCommand{}
}

func (c *AddCommand) Name() string {
	return "add"
}

func (c *AddCommand) Aliases() []string {
	return nil
}

func (c *AddCommand) Description() string {
	return "未完了タスクを末尾へ追加します（-q: 静か, --urgent: 先頭に追加）"
}

// parseAddFlags は add コマンドのフラグをパースします。
func parseAddFlags(args []string) (flags AddFlags, remaining []string) {
	for _, arg := range args {
		if arg == "--quiet" || arg == "-q" {
			flags.Quiet = true
		} else if arg == "--urgent" {
			flags.Urgent = true
		} else {
			remaining = append(remaining, arg)
		}
	}
	return flags, remaining
}

// readFromStdin は標準入力からテキストを読み取ります。
func readFromStdin() (string, error) {
	// 標準入力がパイプかどうかを確認
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("標準入力の状態確認に失敗しました: %w", err)
	}

	// パイプでない場合は空文字列を返す
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}

	// 標準入力から複数行を読み取る
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("標準入力の読み取りに失敗しました: %w", err)
	}

	// 改行を保持して結合
	return strings.Join(lines, "\n"), nil
}

// Run は新しいタスクを未完了リストの末尾に追加します。
func (c *AddCommand) Run(ctx *cli.CommandContext, args []string) error {
	// フラグをパース
	flags, remaining := parseAddFlags(args)

	var text string
	if len(remaining) > 0 {
		// 引数がある場合: 複数の引数は空白で結合して1つのテキストとして扱う
		text = strings.Join(remaining, " ")
	} else {
		// 引数がない場合: 標準入力から読み取る
		stdinText, err := readFromStdin()
		if err != nil {
			return err
		}
		if stdinText == "" {
			return fmt.Errorf("usage: %s add <text>", ctx.BinaryName)
		}
		text = stdinText
	}

	task, err := ctx.Store.AddTask(ctx, text)
	if err != nil {
		if err == storage.ErrEmptyText {
			return err
		}
		return fmt.Errorf("タスク追加に失敗しました: %w", err)
	}

	// 緊急モードの場合は先頭に移動
	if flags.Urgent {
		// タスクIDを取得して先頭に移動
		pending, _, err := ctx.Store.ListTasks(ctx)
		if err != nil {
			return fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
		}

		// 追加したタスクを見つける
		taskIndex := -1
		for i, t := range pending {
			if t.ID == task.ID {
				taskIndex = i + 1 // 1-based index
				break
			}
		}

		if taskIndex > 0 {
			// 先頭に移動（move コマンドと同じロジック）
			if _, err := ctx.Store.MoveTask(ctx, taskIndex, 1); err != nil {
				return fmt.Errorf("緊急タスクの移動に失敗しました: %w", err)
			}
		}
	}

	// quietモードでは出力を抑制
	if !flags.Quiet {
		prefix := "追加"
		if flags.Urgent {
			prefix = "緊急追加"
		}
		fmt.Fprintf(ctx.Stdout, "%s: %s\n", prefix, task.Text)
	}
	return nil
}
