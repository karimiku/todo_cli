# infra/foundation → main

## 概要

Todo CLI の基盤となるインフラストラクチャを構築しました。データベース接続、設定管理、CLI フレームワークの土台を実装しています。

## 変更内容

### 追加されたファイル

- **README.md**: プロジェクト概要・使い方・開発方法を記載
- **LICENSE**: MIT License
- **.gitignore**: Go プロジェクト用の gitignore
- **go.mod/go.sum**: Go モジュール定義と依存関係

### 実装した機能

#### 1. 設定管理（`internal/config/`）

- データベースファイルの保存パス解決ロジック
- 環境変数 `TODO_HOME` のサポート
- デフォルトは `~/.config/todo/todo.db`
- 単体テスト完備

#### 2. SQLite ストレージ（`internal/storage/`）

- GORM を使った SQLite 接続
- Task 構造体の定義
- PRAGMA 設定（WAL、busy_timeout 等）
- オートマイグレーション対応
- 単体テスト完備

#### 3. CLI フレームワーク（`internal/cli/`）

- サブコマンドディスパッチの基盤
- エラーハンドリング
- コマンド登録システム

#### 4. エントリーポイント（`cmd/todo/`）

- main.go: アプリケーションの起動処理
- CLI の初期化と実行

### テスト

- `internal/config/path_test.go`: 設定パス解決のテスト
- `internal/storage/storage_test.go`: ストレージ操作のテスト

## 技術スタック

- Go 1.21+
- GORM (gorm.io/gorm)
- SQLite ドライバ (gorm.io/driver/sqlite)

## コミット履歴

```
dd1ba10 test: cover config and storage foundations
0c28226 feat: scaffold CLI foundation
ba39e63 docs: clarify extended command specs
```

## 次のステップ

この基盤を元に、以下を実装予定:

- `feature/core-commands`: add/list/next コマンド
- `feature/extended-commands`: focus/move/edit/delete コマンド

## 動作確認

```bash
# ビルド
go build -o todo ./cmd/todo

# 実行（現時点ではまだコマンドは動作しません）
./todo
```




