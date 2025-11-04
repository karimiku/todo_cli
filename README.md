# todo - Terminal-First Todo CLI

ターミナルから即登録 → 順番厳守で片付けする超軽量 Todo。

## 特徴

- **順序厳守**: 完了は常に先頭のみ
- **シンプル**: 日付や締切なし、「今やる順番」だけに集中
- **ローカル完結**: オフラインで動作、外部送信なし
- **高速**: 起動〜完了まで体感 < 300ms
- **メニューバー表示**: SwiftBar/xbar と連携して「今やるべき 1 件」を常時表示

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

```bash
# タスクを追加
todo add メール返信

# リスト表示（完了済みは日付ごとにグルーピング）
todo list
# 1 メール返信
# 2 図書館に本返却
# 3 買い出し
#
# ---
#
# 2025/11/04 (2件)
#   [✅] 会議メモ整理
#   [✅] 資料作成
#
# 2025/11/03 (1件)
#   [✅] メール確認

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
# ...
# 削除する場合は: clean YYYY/MM/DD

todo clean 2025/11/03  # 指定した日付の完了済みタスクを削除

# メニューバーに 1 件だけ常時表示
todo bar --head-only --maxlen 28 --icon "▶️"
```

## データ保存先

- デフォルト: `~/.config/todo/todo.db`
- カスタム: 環境変数 `TODO_HOME` を設定するとそのディレクトリに保存

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
go build

# テスト
go test ./...
```

## ライセンス

現状は公開ライセンスを設定していません。個人利用を前提に運用しています。
