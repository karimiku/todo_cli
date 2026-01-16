package render

import (
	"strings"
)

// ExtractSummary はテキストから要約を抽出します。
// 最初の行を取得し、Markdown ヘッダー記号を除去し、maxLen で切り詰めます。
func ExtractSummary(text string, maxLen int) string {
	// 改行で分割して最初の行を取得
	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Markdown ヘッダー記号を除去
	firstLine = strings.TrimPrefix(firstLine, "# ")
	firstLine = strings.TrimPrefix(firstLine, "## ")
	firstLine = strings.TrimPrefix(firstLine, "### ")
	firstLine = strings.TrimPrefix(firstLine, "#### ")
	firstLine = strings.TrimPrefix(firstLine, "##### ")
	firstLine = strings.TrimPrefix(firstLine, "###### ")
	firstLine = strings.TrimSpace(firstLine)

	// 最大文字数で切り詰め（UTF-8 セーフ）
	return TruncateString(firstLine, maxLen)
}

// TruncateString は文字列を指定された文字数で切り詰めます（rune ベース）
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

// ExtractTitle は要約をさらに短くしたタイトルを抽出します
func ExtractTitle(text string, maxLen int) string {
	summary := ExtractSummary(text, maxLen)

	// 単語境界で切る（将来の最適化）
	// 現状はシンプルに切り詰めるだけ

	return summary
}
