package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown は Markdown テキストをターミナル用にレンダリングします。
func RenderMarkdown(content string, width int) (string, error) {
	// ターミナルレンダラーを作成
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		// レンダリング失敗時は生テキストを返す
		return content, nil
	}

	// Markdownをレンダリング
	rendered, err := r.Render(content)
	if err != nil {
		// レンダリング失敗時は生テキストを返す
		return content, nil
	}

	return rendered, nil
}

// RenderMarkdownPreview は Markdown テキストをプレビュー用にレンダリングします（高さ制限付き）。
func RenderMarkdownPreview(content string, width, maxLines int) (string, error) {
	// まず完全にレンダリング
	rendered, err := RenderMarkdown(content, width)
	if err != nil {
		return content, nil
	}

	// 行数制限を適用
	lines := strings.Split(strings.TrimSpace(rendered), "\n")
	if len(lines) <= maxLines {
		return rendered, nil
	}

	// maxLines を超える場合は切り詰めて「続きあり」を表示
	truncated := strings.Join(lines[:maxLines], "\n")
	truncated += fmt.Sprintf("\n\n↓ 続きあり (%d行中%d行表示)", len(lines), maxLines)

	return truncated, nil
}
