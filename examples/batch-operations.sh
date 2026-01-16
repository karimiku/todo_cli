#!/bin/bash
# バッチ処理のサンプルスクリプト

# 複数のタスクを一括追加
echo "=== 複数タスクの一括追加 ==="
tasks=(
    "メール返信"
    "会議資料作成"
    "コードレビュー"
    "ドキュメント更新"
    "テスト実施"
)

for task in "${tasks[@]}"; do
    todo add "$task" --quiet
done

echo "✅ ${#tasks[@]}件のタスクを追加しました"

# ファイルからタスクを読み込んで追加
echo -e "\n=== ファイルからタスク追加 ==="
if [ -f "tasks.txt" ]; then
    while IFS= read -r line; do
        if [ -n "$line" ]; then
            echo "$line" | todo add --quiet
        fi
    done < tasks.txt
    echo "✅ tasks.txt からタスクを追加しました"
fi

# Markdownファイルをタスクとして追加
echo -e "\n=== Markdownファイルをタスクとして追加 ==="
if [ -f "long-task-example.md" ]; then
    cat long-task-example.md | todo add --quiet
    echo "✅ Markdownファイルをタスクとして追加しました"
fi

# 現在のタスク状況を表示
echo -e "\n=== 現在のタスク状況 ==="
pending=$(todo list --json | jq '.count.pending')
completed=$(todo list --json | jq '.count.completed')

echo "未完了: $pending 件"
echo "完了済: $completed 件"
echo "合計: $((pending + completed)) 件"

# タスクをCSV形式でエクスポート
echo -e "\n=== CSV形式でエクスポート ==="
todo list --json | jq -r '.pending[] | [.ID, .Text, .CreatedAt] | @csv' > tasks.csv
echo "✅ tasks.csv にエクスポートしました"

# 先頭から3つのタスクを表示
echo -e "\n=== 次にやるべき3つのタスク ==="
todo list --json | jq -r '.pending[:3][] | "- \(.Text | split("\n")[0])"'
