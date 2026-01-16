package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/kamiriku/todo_cli/internal/render"
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
	cursor      int   // 選択中の項目のインデックス
	width       int   // ターミナルの幅
	height      int   // ターミナルの高さ
	err         error // エラー情報
	quitting    bool  // 終了フラグ

	// プレビュー表示
	viewport viewport.Model
	ready    bool // viewport の初期化完了フラグ
}

// Init は初期化処理を実行します
func (m model) Init() tea.Cmd {
	return nil
}

// Update はイベントを処理して状態を更新します
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// viewport を初期化
			previewHeight := m.height - 10 // ヘッダー、リスト、区切り線を考慮
			if previewHeight < 5 {
				previewHeight = 5
			}
			m.viewport = viewport.New(m.width, previewHeight)
			m.viewport.YPosition = 0
			m.ready = true
			m.updatePreview()
		} else {
			// viewport のサイズを更新
			previewHeight := m.height - 10
			if previewHeight < 5 {
				previewHeight = 5
			}
			m.viewport.Width = m.width
			m.viewport.Height = previewHeight
		}
		return m, nil
	}

	// viewport の更新
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleKeyPress はキー入力を処理します
func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 詳細表示モード時の特別なキーバインド
	if m.currentMode == modeDetail {
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q", "esc":
			// リスト選択モードに戻る
			m.currentMode = modeList
			m.updatePreview()
			return m, nil
		case "g":
			// 先頭へジャンプ
			m.viewport.GotoTop()
			return m, nil
		case "G":
			// 末尾へジャンプ
			m.viewport.GotoBottom()
			return m, nil
		case "j", "down":
			// 下にスクロール
			m.viewport.LineDown(1)
			return m, nil
		case "k", "up":
			// 上にスクロール
			m.viewport.LineUp(1)
			return m, nil
		case "d":
			// 半ページ下にスクロール
			m.viewport.HalfViewDown()
			return m, nil
		case "u":
			// 半ページ上にスクロール
			m.viewport.HalfViewUp()
			return m, nil
		}
		return m, nil
	}

	// リスト選択モード時のキーバインド
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.updatePreview() // プレビューを更新
		}
	case "down", "j":
		maxCursor := len(m.pending) - 1
		if m.cursor < maxCursor {
			m.cursor++
			m.updatePreview() // プレビューを更新
		}
	case "enter":
		// 詳細表示への切り替え
		m.currentMode = modeDetail
		m.updateDetailView()
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

// updatePreview はプレビュー表示を更新します
func (m *model) updatePreview() {
	if !m.ready || len(m.pending) == 0 || m.cursor >= len(m.pending) {
		m.viewport.SetContent("プレビューなし")
		return
	}

	task := m.pending[m.cursor]

	// Markdownをレンダリング
	rendered, err := render.RenderMarkdown(task.Text, m.width-4)
	if err != nil {
		// エラー時は生テキストを表示
		m.viewport.SetContent(task.Text)
		return
	}

	m.viewport.SetContent(rendered)
}

// updateDetailView は詳細表示モードのビューを更新します
func (m *model) updateDetailView() {
	if !m.ready || len(m.pending) == 0 || m.cursor >= len(m.pending) {
		m.viewport.SetContent("タスクが見つかりません")
		return
	}

	task := m.pending[m.cursor]

	// 詳細表示用のviewportサイズを調整（全画面に近い）
	detailHeight := m.height - 6 // ヘッダーとヘルプを考慮
	if detailHeight < 10 {
		detailHeight = 10
	}
	m.viewport.Width = m.width - 4
	m.viewport.Height = detailHeight

	// Markdownをレンダリング（全画面用）
	rendered, err := render.RenderMarkdown(task.Text, m.width-4)
	if err != nil {
		// エラー時は生テキストを表示
		m.viewport.SetContent(task.Text)
	} else {
		m.viewport.SetContent(rendered)
	}

	// 先頭にジャンプ
	m.viewport.GotoTop()
}

// renderList はリスト表示を描画します（プレビュー付き2ペイン表示）
func (m model) renderList() string {
	var s strings.Builder

	// ヘッダー
	s.WriteString(HeaderStyle.Render("📝 Todo List") + "\n\n")

	if len(m.pending) == 0 {
		s.WriteString("タスクはありません\n")
	} else {
		// ターミナル幅に応じて要約の長さを調整
		summaryWidth := m.width - 10 // 番号と余白を考慮
		if summaryWidth < 40 {
			summaryWidth = 40
		}
		if summaryWidth > 80 {
			summaryWidth = 80
		}

		// リスト表示（上部40%）
		listHeight := (m.height * 4) / 10
		if listHeight < 3 {
			listHeight = 3
		}

		displayCount := listHeight - 3 // ヘッダー等を除く
		if displayCount < 1 {
			displayCount = 1
		}

		// 表示範囲を計算（カーソル周辺を表示）
		startIdx := m.cursor - displayCount/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx := startIdx + displayCount
		if endIdx > len(m.pending) {
			endIdx = len(m.pending)
			startIdx = endIdx - displayCount
			if startIdx < 0 {
				startIdx = 0
			}
		}

		for i := startIdx; i < endIdx; i++ {
			task := m.pending[i]
			// 要約を抽出
			summary := render.ExtractSummary(task.Text, summaryWidth)

			// 番号とカーソル
			number := fmt.Sprintf("%d", i+1)
			cursor := " "

			if m.cursor == i {
				cursor = "▶"
				line := fmt.Sprintf("%s %s %s", cursor, number, summary)
				s.WriteString(SelectedStyle.Render(line) + "\n")
			} else {
				line := fmt.Sprintf("%s %s %s", cursor, number, summary)
				s.WriteString(NormalStyle.Render(line) + "\n")
			}
		}
	}

	// 区切り線
	dividerWidth := m.width - 2
	if dividerWidth < 10 {
		dividerWidth = 10
	}
	divider := strings.Repeat("─", dividerWidth)
	s.WriteString("\n" + DividerStyle.Render(divider) + "\n")

	// プレビューヘッダー
	s.WriteString(PreviewHeaderStyle.Render("📄 Preview") + "\n\n")

	// プレビュー表示（viewport）
	if m.ready {
		s.WriteString(m.viewport.View() + "\n")
	}

	// ヘルプ
	s.WriteString("\n" + HelpStyle.Render("j/k: 移動 | Enter: 詳細 | q: 終了") + "\n")

	return s.String()
}

// renderDetail は詳細表示を描画します（全画面表示）
func (m model) renderDetail() string {
	if m.cursor >= len(m.pending) {
		return "タスクが見つかりません\n"
	}

	var s strings.Builder

	// ヘッダー（タスク番号と情報）
	taskNum := fmt.Sprintf("Task #%d/%d", m.cursor+1, len(m.pending))
	s.WriteString(DetailHeaderStyle.Render(taskNum) + "\n\n")

	// viewport で全画面表示
	if m.ready {
		s.WriteString(m.viewport.View() + "\n")
	}

	// 区切り線
	dividerWidth := m.width - 2
	if dividerWidth < 10 {
		dividerWidth = 10
	}
	divider := strings.Repeat("─", dividerWidth)
	s.WriteString("\n" + DividerStyle.Render(divider) + "\n")

	// ヘルプ
	s.WriteString(HelpStyle.Render("j/k: スクロール | g/G: 先頭/末尾 | d/u: 半ページ | q/Esc: 戻る") + "\n")

	return s.String()
}
