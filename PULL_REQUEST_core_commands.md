# feature/core-commands → main

## 概要

Todo CLI の基本 3 コマンド（add/list/next）を実装しました。タスクの追加・表示・完了処理が可能になります。

## 変更内容

### 実装したコマンド

#### 1. `todo add <text>`

- タスクを未完了リストの末尾に追加
- `order_key` を自動計算（MAX + 1.0）
- 空文字チェック付き

#### 2. `todo list`

- 未完了タスクを番号付きで表示（1..N）
- 完了タスクは下部に `[x]` 表示
- `order_key` で並び替え

#### 3. `todo next`

- 先頭タスク（最小 order_key）を完了状態に変更
- 削除せずに `[x]` マーク
- 未完了タスクが 0 件の場合のメッセージ対応

### 追加されたファイル

- **internal/commands/add.go**: add コマンドの実装
- **internal/commands/list.go**: list コマンドの実装
- **internal/commands/next.go**: next コマンドの実装

### 改善点

- CLI ディスパッチャーの実装強化（`internal/cli/app.go`）
- エラーメッセージの統一
- トランザクション管理の追加

## 技術詳細

### データベース操作

- **add**: `INSERT` + `SELECT MAX(order_key)` をトランザクション実行
- **list**: `SELECT` WHERE is_done=0 ORDER BY order_key, id
- **next**: `UPDATE` WHERE is_done=0 AND order_key=MIN を 1 件のみ更新

### エラーハンドリング

- 空のテキスト: 「空のテキストは不可」を表示
- 未完了 0 件: 「未完了はありません 🎉」を表示

## コミット履歴

```
e02183f feat: wire core commands
70188bc feat: add core todo commands
8583101 feat: create cli dispatcher
e649fc7 feat: implement sqlite task storage
bb3dffa feat: add config helpers for database path
5ed3fb6 chore: initialize go module
```

## 動作確認

```bash
# ビルド
go build -o todo ./cmd/todo

# タスク追加
./todo add メール返信
./todo add 図書館に本返却

# リスト表示
./todo list
# 1 メール返信
# 2 図書館に本返却

# 先頭完了
./todo next

# 再表示
./todo list
# 1 図書館に本返却
#
# ---
#
# [x] メール返信
```

## 次のステップ

- `feature/extended-commands`: focus/move/edit/delete コマンドの実装


