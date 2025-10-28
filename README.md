# todo - Terminal-First Todo CLI

ターミナルから即登録 → 順番厳守で片付けする超軽量 Todo。

## 特徴

- **順序厳守**: 完了は常に先頭のみ
- **シンプル**: 日付や締切なし、「今やる順番」だけに集中
- **ローカル完結**: オフラインで動作、外部送信なし
- **高速**: 起動〜完了まで体感 < 300ms

## インストール

```bash
# ビルド
go build -o todo .

# 配置
sudo cp todo /usr/local/bin/

# エイリアス設定 (.zshrc に追加)
alias tb='todo'
alias td='todo'
```

## 使い方

```bash
# タスクを追加
todo add メール返信

# リスト表示
todo list
# 1 メール返信
# 2 図書館に本返却
# 3 買い出し
#
# ---
#
# [x] 会議メモ整理

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
```

## データ保存先

- デフォルト: `~/.config/todo/todo.db`
- カスタム: 環境変数 `TODO_HOME` を設定するとそのディレクトリに保存

## ライセンス

MIT License (LICENSE ファイルを参照)

## 開発

```bash
# 依存インストール
go mod download

# ビルド
go build

# テスト
go test ./...
```

## 目的

このプロジェクトは個人利用のみを目的としています。配布・リリースは予定していません。
