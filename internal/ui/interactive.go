package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kamiriku/todo_cli/internal/cli"
)

// ShowInteractive はインタラクティブモードを起動します
func ShowInteractive(ctx *cli.CommandContext) error {
	// タスクを取得
	pending, completed, err := ctx.Store.ListTasks(ctx)
	if err != nil {
		return fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
	}

	// モデルを初期化
	m := model{
		pending:     pending,
		completed:   completed,
		currentMode: modeList,
		cursor:      0,
	}

	// TUIを起動
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUIの起動に失敗しました: %w", err)
	}

	return nil
}
