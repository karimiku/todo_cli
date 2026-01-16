package render

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		contains []string // レンダリング結果に含まれるべき文字列
	}{
		{
			name:  "シンプルなテキスト",
			input: "これはテストです",
			width: 80,
			contains: []string{
				"これはテストです",
			},
		},
		{
			name:  "見出し",
			input: "# 見出し1\n\n## 見出し2",
			width: 80,
			contains: []string{
				"見出し1",
				"見出し2",
			},
		},
		{
			name:  "リスト",
			input: "- 項目1\n- 項目2\n- 項目3",
			width: 80,
			contains: []string{
				"項目1",
				"項目2",
				"項目3",
			},
		},
		{
			name:  "コードブロック",
			input: "```go\nfunc main() {\n}\n```",
			width: 80,
			contains: []string{
				"func",
				"main",
			},
		},
		{
			name:  "複合的なMarkdown",
			input: "# データベース最適化\n\n## 背景\n検索が遅い\n\n## 実施内容\n1. インデックス追加\n2. テスト実施",
			width: 80,
			contains: []string{
				"データベース最適化",
				"背景",
				"検索が遅い",
				"実施内容",
				"インデックス追加",
				"テスト実施",
			},
		},
		{
			name:  "空文字列",
			input: "",
			width: 80,
			contains: []string{
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderMarkdown(tt.input, tt.width)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("result does not contain %q\ngot:\n%s", expected, result)
				}
			}
		})
	}
}

func TestRenderMarkdownPreview(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		width       int
		maxLines    int
		shouldTrunc bool
	}{
		{
			name:        "短いテキスト（切り詰めなし）",
			input:       "短いテキスト",
			width:       80,
			maxLines:    10,
			shouldTrunc: false,
		},
		{
			name:        "長いテキスト（切り詰めあり）",
			input:       strings.Repeat("- 項目\n", 20),
			width:       80,
			maxLines:    5,
			shouldTrunc: true,
		},
		{
			name: "複数行Markdown（切り詰めあり）",
			input: `# タイトル

## セクション1
内容1

## セクション2
内容2

## セクション3
内容3

## セクション4
内容4`,
			width:       80,
			maxLines:    10,
			shouldTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderMarkdownPreview(tt.input, tt.width, tt.maxLines)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(result), "\n")
			hasContinuation := strings.Contains(result, "↓ 続きあり")

			if tt.shouldTrunc {
				if !hasContinuation {
					t.Errorf("expected truncation marker, but not found\ngot:\n%s", result)
				}
			} else {
				if hasContinuation {
					t.Errorf("unexpected truncation marker\ngot:\n%s", result)
				}
				if len(lines) > tt.maxLines+5 { // +5 は余裕を持たせる（レンダリングで行が増える可能性）
					t.Errorf("expected at most %d lines (with margin), got %d", tt.maxLines+5, len(lines))
				}
			}
		})
	}
}

func TestRenderMarkdownErrorHandling(t *testing.T) {
	// エラーが発生しても生テキストを返すことを確認
	input := "# テスト"
	result, err := RenderMarkdown(input, 80)

	// エラーがあっても nil を返す（生テキストでフォールバック）
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// 何らかの結果が返ってくることを確認
	if result == "" {
		t.Error("expected non-empty result")
	}
}
