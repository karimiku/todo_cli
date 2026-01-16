package render

import "testing"

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "通常のテキスト",
			input:    "これは通常のテキストです",
			maxLen:   20,
			expected: "これは通常のテキストです",
		},
		{
			name:     "Markdown ヘッダー (# )",
			input:    "# これはヘッダーです",
			maxLen:   20,
			expected: "これはヘッダーです",
		},
		{
			name:     "Markdown ヘッダー (## )",
			input:    "## これはヘッダー2です",
			maxLen:   20,
			expected: "これはヘッダー2です",
		},
		{
			name:     "Markdown ヘッダー (### )",
			input:    "### これはヘッダー3です",
			maxLen:   30,
			expected: "これはヘッダー3です",
		},
		{
			name:     "長文の切り詰め",
			input:    "これは非常に長いテキストで、切り詰める必要があります",
			maxLen:   10,
			expected: "これは非常に長いテキ…",
		},
		{
			name:     "複数行（最初の行のみ）",
			input:    "最初の行\n2行目\n3行目",
			maxLen:   20,
			expected: "最初の行",
		},
		{
			name:     "全角文字",
			input:    "日本語の文章をテストします",
			maxLen:   10,
			expected: "日本語の文章をテスト…",
		},
		{
			name:     "絵文字を含む",
			input:    "絵文字テスト 🎉🎊✨",
			maxLen:   10,
			expected: "絵文字テスト 🎉🎊✨",
		},
		{
			name:     "前後の空白を除去",
			input:    "  空白がある  ",
			maxLen:   20,
			expected: "空白がある",
		},
		{
			name:     "maxLen以下のテキスト",
			input:    "短い",
			maxLen:   10,
			expected: "短い",
		},
		{
			name:     "空文字列",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "複数のヘッダー記号（最初のみ除去）",
			input:    "# これは # 含むテキスト",
			maxLen:   30,
			expected: "これは # 含むテキスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSummary(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "切り詰めなし",
			input:    "短いテキスト",
			maxLen:   10,
			expected: "短いテキスト",
		},
		{
			name:     "切り詰めあり",
			input:    "これは長いテキストです",
			maxLen:   5,
			expected: "これは長い…",
		},
		{
			name:     "UTF-8マルチバイト文字",
			input:    "あいうえおかきくけこ",
			maxLen:   5,
			expected: "あいうえお…",
		},
		{
			name:     "絵文字",
			input:    "🎉🎊✨🌟⭐",
			maxLen:   3,
			expected: "🎉🎊✨…",
		},
		{
			name:     "ちょうどmaxLen",
			input:    "12345",
			maxLen:   5,
			expected: "12345",
		},
		{
			name:     "maxLen = 0",
			input:    "テキスト",
			maxLen:   0,
			expected: "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "通常のテキスト",
			input:    "これはタイトルです",
			maxLen:   20,
			expected: "これはタイトルです",
		},
		{
			name:     "Markdownヘッダー",
			input:    "# タイトル\n本文が続きます",
			maxLen:   10,
			expected: "タイトル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTitle(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
