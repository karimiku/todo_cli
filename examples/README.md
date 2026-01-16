# Examples

todo CLI の使用例とサンプルスクリプト集です。

## ファイル一覧

### 1. `ai-integration.sh`

AI アシスタント（ClaudeCode など）との連携例を示すスクリプトです。

**使用方法:**
```bash
chmod +x ai-integration.sh
./ai-integration.sh
```

**含まれる例:**
- 静かにタスクを追加（`--quiet`）
- 緊急タスクの追加（`--urgent`）
- JSON 形式でのタスク取得
- 詳細な作業メモの保存
- タスクのバックアップ

### 2. `long-task-example.md`

長文タスクのサンプルです。Markdown で詳細な情報を含むタスクの例を示します。

**使用方法:**
```bash
# ファイル全体をタスクとして追加
cat long-task-example.md | todo add

# または
todo add < long-task-example.md

# インタラクティブモードで確認
todo list -d
```

**含まれる情報:**
- プロジェクトの背景
- 現状分析
- 実施内容（チェックリスト付き）
- 期待される効果
- 参考資料

### 3. `batch-operations.sh`

バッチ処理のサンプルスクリプトです。

**使用方法:**
```bash
chmod +x batch-operations.sh
./batch-operations.sh
```

**含まれる例:**
- 複数タスクの一括追加
- ファイルからのタスク読み込み
- タスク状況の確認
- CSV 形式でのエクスポート
- 上位N件のタスク表示

## 実践的な使用例

### 朝のルーティン

```bash
# 今日のタスクを確認
todo list -d

# AIに今日の計画を立ててもらう
# AIが重要なタスクを先頭に移動
todo focus 3
```

### コードレビュー時

```bash
# レビューコメントをタスクとして保存
todo add --urgent <<EOF
# PR #123 レビュー対応

## 修正が必要な箇所
1. エラーハンドリング（line 45）
2. テストケース追加
3. ドキュメント更新

## 優先度: 高
明日の朝までに対応
EOF
```

### 作業終了時

```bash
# 今日完了したタスクを確認
todo clean

# 明日のタスクを確認
todo list --json | jq -r '.pending[:5][] | "- \(.Text | split("\n")[0])"'
```

## Tips

1. **スクリプトから使う場合は `--quiet` を使う**
   ```bash
   result=$(some_command)
   echo "Result: $result" | todo add --quiet
   ```

2. **JSON 出力と jq を組み合わせる**
   ```bash
   # タスクIDのリスト
   todo list --json | jq '.pending[].ID'
   
   # 最初の行だけ取得
   todo list --json | jq -r '.pending[].Text' | head -1
   ```

3. **ヒアドキュメントで複数行タスク**
   ```bash
   todo add <<EOF
   複数行の
   タスクを
   追加
   EOF
   ```

4. **条件付きで追加**
   ```bash
   if [ $exit_code -ne 0 ]; then
       echo "エラー発生: $error_message" | todo add --urgent
   fi
   ```
