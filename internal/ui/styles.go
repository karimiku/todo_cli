package ui

import "github.com/charmbracelet/lipgloss"

var (
	// HeaderStyle はヘッダー用のスタイル
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	// DividerStyle は区切り線用のスタイル
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// SelectedStyle は選択中の項目用のスタイル
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	// NormalStyle は通常の項目用のスタイル
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	// HelpStyle はヘルプテキスト用のスタイル
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// DetailStyle は詳細表示用のスタイル
	DetailStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	// PreviewHeaderStyle はプレビューヘッダー用のスタイル
	PreviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("99")).
				Padding(0, 1)
)
