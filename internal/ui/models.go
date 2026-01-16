package ui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/kamiriku/todo_cli/internal/storage"
)

// mode はTUIの表示モードを表す型
type mode int

const (
	modeList   mode = iota // リスト表示モード
	modeDetail             // 詳細表示モード
)

// model はTUIの状態を管理する構造体
type model struct {
	// タスクデータ
	pending   []storage.Task
	completed []storage.Task

	// 表示状態
	currentMode mode
	cursor      int    // 選択中の項目のインデックス
	width       int    // ターミナルの幅
	height      int    // ターミナルの高さ
	err         error  // エラー情報
	quitting    bool   // 終了フラグ
}

// Init は初期化処理を実行します
func (m model) Init() tea.Cmd {
	return nil
}

// Update はイベントを処理して状態を更新します
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

// handleKeyPress はキー入力を処理します
func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		maxCursor := len(m.pending) - 1
		if m.cursor < maxCursor {
			m.cursor++
		}
	case "enter":
		// 詳細表示への切り替え（後で実装）
		if m.currentMode == modeList {
			m.currentMode = modeDetail
		} else {
			m.currentMode = modeList
		}
	case "esc":
		// リスト表示に戻る
		if m.currentMode == modeDetail {
			m.currentMode = modeList
		}
	}
	return m, nil
}

// View は画面を描画します
func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return "エラーが発生しました: " + m.err.Error() + "\n"
	}

	switch m.currentMode {
	case modeDetail:
		return m.renderDetail()
	default:
		return m.renderList()
	}
}

// renderList はリスト表示を描画します（骨組み）
func (m model) renderList() string {
	s := HeaderStyle.Render("📝 Todo List") + "\n\n"

	if len(m.pending) == 0 {
		s += "タスクはありません\n"
	} else {
		for i, task := range m.pending {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += SelectedStyle.Render(cursor + " " + task.Text[:min(len(task.Text), 60)]) + "\n"
			} else {
				s += NormalStyle.Render(cursor + " " + task.Text[:min(len(task.Text), 60)]) + "\n"
			}
		}
	}

	s += "\n" + DividerStyle.Render("────────────────────────────────────────") + "\n"
	s += HelpStyle.Render("j/k: 移動 | Enter: 詳細 | q: 終了") + "\n"

	return s
}

// renderDetail は詳細表示を描画します（骨組み）
func (m model) renderDetail() string {
	if m.cursor >= len(m.pending) {
		return "タスクが見つかりません\n"
	}

	task := m.pending[m.cursor]

	s := HeaderStyle.Render("📄 Task Detail") + "\n\n"
	s += DetailStyle.Render(task.Text) + "\n\n"
	s += HelpStyle.Render("Esc/Enter: 戻る | q: 終了") + "\n"

	return s
}

// min は2つの整数のうち小さい方を返します
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
