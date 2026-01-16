# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - 2026-01-16

### 🎉 Major Release: AI-Friendly & Interactive Mode

v2.0.0 は AI アシスタントとの連携と、人間にとっての使いやすさを大幅に向上させるメジャーリリースです。

### ✨ 新機能

#### AI フレンドリー機能
- **JSON 出力** (`--json`): 構造化されたデータ出力で AI/スクリプトから利用可能
- **静かな追加モード** (`--quiet`, `-q`): 出力を抑制してスクリプト向けに最適化
- **標準入力対応**: パイプやヒアドキュメントからタスクを追加可能
- **緊急追加モード** (`--urgent`): タスクを先頭に追加して即座に対応

#### インタラクティブモード
- **詳細表示モード** (`-d`, `--detail`): TUI でタスクを快適に閲覧
  - 2ペイン表示（リスト + プレビュー）
  - リアルタイムMarkdownレンダリング
  - vim風キーバインド（j/k で移動、Enter で詳細表示）
- **全画面表示**: 長文タスクをスクロール可能な全画面で表示
  - j/k: スクロール
  - g/G: 先頭/末尾へジャンプ
  - d/u: 半ページスクロール
  - q/Esc: 戻る

#### Markdown サポート
- **要約表示**: 長文タスクは自動的に要約（最初の行のみ、60文字制限）
- **Markdownレンダリング**: 見出し、コードブロック、リストを美しく表示
  - シンタックスハイライト対応
  - 自動的にターミナルテーマに適応

### 🔧 改善

- **ヘルプメッセージ**: 新機能のフラグ情報を追加
- **パフォーマンス**: レンダリングとTUIの最適化
- **エラーハンドリング**: より詳細なエラーメッセージ

### 📦 依存ライブラリ

v2 では以下のライブラリを追加：
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI フレームワーク
- `github.com/charmbracelet/bubbles` v0.21.0 - TUI コンポーネント
- `github.com/charmbracelet/glamour` v0.10.0 - Markdown レンダリング
- `github.com/charmbracelet/lipgloss` - スタイリング

### 🔄 破壊的変更

なし。v1 の全機能は引き続き動作します。

### 📝 マイグレーション

既存のデータベースはそのまま使用できます。アップグレードに特別な手順は不要です。

### 🎯 使用例

```bash
# JSON出力でAIと連携
todo list --json | jq '.pending[0].Text'

# 静かにタスク追加
echo "タスク" | todo add --quiet

# 緊急タスクを先頭に
todo add "本番障害対応" --urgent

# インタラクティブモードで快適に閲覧
todo list -d
```

### 🐛 既知の制約

- インタラクティブモードは ANSI カラー対応ターミナルが必要です
- Windows での動作は未検証（Linux/macOS で動作確認済み）

---

## [1.0.0] - 2025-11-XX

### 初回リリース

基本的なTodo管理機能：
- タスクの追加・完了・削除・編集
- 順序管理（focus, move）
- メニューバー連携（SwiftBar/xbar）
- 完了済みタスクの管理（clean）
