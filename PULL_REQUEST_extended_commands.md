# feature/extended-commands → main

## 概要

Todo CLI の拡張コマンド（focus/move/edit/delete）と UX 改善を実装しました。全コマンドが完成し、エイリアス対応も追加しています。

## 実装予定のコマンド

### 1. `todo focus <id>`

- 指定した ID のタスクを最上位に移動
- 表示順位から内部 ID への解決処理
- `order_key = MIN - 1.0` で実現

### 2. `todo move <id> <pos>`

- 任意の位置へタスクを移動
- 先頭/末尾/中間位置への対応
- `order_key` の再計算（平均値アルゴリズム）

### 3. `todo delete <id>`

- 未完了タスクの物理削除
- 誤登録の削除に利用

### 4. `todo edit <id> <text>`

- タスクの文言編集
- 空文字チェック付き

### 5. エイリアス対応

- `tb` → `todo`
- `td` → `todo`
- 同じデータベースを共有

## UX 改善

- ヘルプメッセージの改善
- 成功メッセージの統一
- エラーメッセージの明確化

## 技術詳細

### ID 解決ロジック

表示 ID（1..N）を内部 ID に変換する処理を実装:

1. 未完了タスクを `order_key` 順に取得
2. 表示順位を付与してマッピング
3. 指定 ID の内部 ID を特定

### 並び替えアルゴリズム

`move` コマンドの位置移動:

- 先頭へ: `order_key = MIN - 1.0`
- 末尾へ: `order_key = MAX + 1.0`
- 中間へ: `order_key = (前の値 + 後の値) / 2.0`

## テスト計画

- 各コマンドの統合テスト
- ID 解決ロジックのテスト
- 並び替えアルゴリズムのテスト

## 動作確認（実装後）

```bash
# ビルド
go build -o todo ./cmd/todo

# 拡張コマンドの使用例
./todo add タスク1
./todo add タスク2
./todo add タスク3

# focus: 2番目を最優先に
./todo focus 2

# move: 3番目を1番目に
./todo move 3 1

# edit: タスク編集
./todo edit 2 編集後のタスク

# delete: 削除
./todo delete 2

# エイリアス
tb list
td add テストタスク
```

## 注意事項

このブランチは `infra/foundation` と `feature/core-commands` のマージ後に開発を進める必要があります。


