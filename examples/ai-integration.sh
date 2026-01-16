#!/bin/bash
# AI連携のサンプルスクリプト
# ClaudeCodeや他のAIアシスタントから使用する例

# 1. 静かにタスクを追加（出力なし）
echo "ユーザー認証のリファクタリング" | todo add --quiet

# 2. 緊急タスクを先頭に追加
echo "本番環境でエラー発生 - 調査が必要" | todo add --urgent --quiet

# 3. JSON形式で現在のタスクを取得
current_task=$(todo list --json | jq -r '.pending[0].Text')
echo "現在のタスク: $current_task"

# 4. タスク数を確認
pending_count=$(todo list --json | jq '.count.pending')
echo "未完了タスク数: $pending_count"

# 5. 詳細な作業メモをMarkdownで保存
todo add --quiet <<EOF
# PRレビュー対応 #123

## 指摘事項
1. エラーハンドリング不足
   - try-catch を追加
2. テストケース追加
   - ユニットテストが不足
3. ドキュメント更新
   - API仕様書を更新

## 期限
明日の午前中まで
EOF

# 6. 全タスクをJSON形式でエクスポート
todo list --json > tasks_backup.json
echo "タスクをバックアップしました: tasks_backup.json"
