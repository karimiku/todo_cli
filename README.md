# todo - Terminal-First Todo CLI

ターミナルから即登録 → 順番厳守で片付けする超軽量 Todo。

## 特徴

- **順序厳守**: 完了は常に先頭のみ
- **シンプル**: 日付や締切なし、「今やる順番」だけに集中
- **ローカル完結**: オフラインで動作、外部送信なし
- **高速**: 起動〜完了まで体感 < 300ms
- **メニューバー表示**: SwiftBar/xbar と連携して「今やるべき 1 件」を常時表示
- **AI フレンドリー**: JSON 出力、静かな追加、Markdown 対応で AI アシスタントと連携
- **インタラクティブ表示**: vim 風の操作で詳細を快適に確認

## インストール

```bash
# ビルド
go build -o todo ./cmd/todo
go build -o tb ./cmd/tb
go build -o td ./cmd/td

# 配置（必要に応じて）
sudo cp todo /usr/local/bin/
sudo cp tb /usr/local/bin/
sudo cp td /usr/local/bin/
```

`todo` `tb` `td` は同じデータベースを共有します。好みのコマンド名で実行できます。

## 使い方

### 基本操作

```bash
# タスクを追加
todo add メール返信

# Markdown形式で詳細な情報を含めて追加
todo add "# データベース最適化

## 背景
検索が遅い（3秒以上）

## 実施内容
1. EXPLAIN ANALYZEで調査
2. インデックス追加
3. パフォーマンステスト
"

# 緊急タスクを先頭に追加
todo add "本番障害対応" --urgent

# 静かに追加（出力なし、スクリプト向け）
echo "タスク内容" | todo add --quiet

# リスト表示（長文は最初の行のみ表示）
todo list
# 1 データベース最適化
# 2 メール返信
# 3 図書館に本返却

# インタラクティブ詳細表示（vim風操作）
todo list -d
# j/k: 移動、Enter: 詳細表示、q: 終了
# Markdownが美しくレンダリングされます

# JSON形式で出力（AI/スクリプト向け）
todo list --json

# 先頭を完了（削除せず [x] 化）
todo next

# 緊急差し込み（一番上へ移動）
todo focus 2

# 任意の位置へ移動
todo move 3 1

# 削除（誤登録用）
todo delete 2

# 編集
todo edit 1 メールを確認して返信

# 完了済みタスクを日付ごとに表示・削除
todo clean
# 2025/11/04 (2件)
#   [✅] 会議メモ整理
#   [✅] 資料作成

todo clean 2025/11/03  # 指定した日付の完了済みタスクを削除

# メニューバーに 1 件だけ常時表示
todo bar --head-only --maxlen 28 --icon "▶️"
```

## AI アシスタント連携

ClaudeCode や他の AI アシスタントと組み合わせると、「あとでやる」タスクを忘れずに管理できます。

### AI からの使用例

```bash
# AI がタスクを追加
echo "ユーザー認証のリファクタリング" | todo add --quiet

# AI が現在のタスクを確認
todo list --json | jq '.pending[0].text'

# AI が詳細な作業メモを保存
todo add "# PRレビュー対応 #123

## 指摘事項
1. エラーハンドリング不足
2. テストケース追加
3. コメント追加
"
```

### 人間とAIの協働シナリオ

1. **作業中断時**: AI が現在の作業内容を todo に保存
2. **コードレビュー**: AI が指摘事項を構造化して todo に追加
3. **朝の確認**: `todo list -d` で詳細を確認し、AI と作業開始
4. **完了報告**: `todo next` で完了、AI が次のタスクを提案

## データ保存先

- デフォルト: `~/.config/todo/todo.db`
- カスタム: 環境変数 `TODO_HOME` を設定するとそのディレクトリに保存

## インタラクティブ詳細表示

長文タスクや Markdown で書かれた詳細情報を快適に確認できます。

```bash
todo list -d
```

### 操作方法

| キー | 動作 |
|------|------|
| `j` / `k` | リスト内で上下移動 |
| `Enter` | 選択したタスクを全画面表示 |
| `q` | 終了（詳細表示中は一覧に戻る） |
| `gg` / `G` | 詳細表示中に先頭/末尾へ |

### 表示の特徴

- **2ペイン表示**: 上部にリスト、下部にプレビュー
- **Markdown レンダリング**: 見出し、コードブロック、チェックボックスを美しく表示
- **リアルタイム更新**: カーソル移動と同時にプレビューが切り替わる
- **vim 風操作**: 直感的なキーバインド

### 推奨する使い方

1. 通常は `todo list` で要約のみ確認
2. 詳細が必要な時だけ `todo list -d` で確認
3. AI は詳細情報を Markdown で保存
4. 人間は快適に閲覧・確認

## メニューバー連携（SwiftBar/xbar）

```bash
# ~/Library/Application Support/SwiftBar/Plugins/todo-next.2s.sh などに配置
#!/bin/zsh
todo bar --head-only --maxlen 28 --icon "▶️"
```

- ファイルに実行権限を付与: `chmod +x todo-next.2s.sh`
- 更新間隔は 2〜5 秒がおすすめ
- 行をクリックすると `todo next` が自動実行され、即座に次のタスクへ切り替わります

## 開発

```bash
# 依存インストール
go mod download

# ビルド
go build -o todo ./cmd/todo

# テスト
go test ./...
```

### 依存ライブラリ（v2 新機能用）

- `github.com/charmbracelet/bubbletea` - TUI フレームワーク
- `github.com/charmbracelet/bubbles` - TUI コンポーネント
- `github.com/charmbracelet/glamour` - Markdown レンダリング
- `github.com/charmbracelet/lipgloss` - スタイリング

インタラクティブモードを使用しない場合、これらの依存は不要です。

## ライセンス

現状は公開ライセンスを設定していません。個人利用を前提に運用しています。
