.PHONY: help build test lint fmt clean install dev-setup ci

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make build       - 全てのバイナリをビルド"
	@echo "  make test        - テストを実行"
	@echo "  make lint        - リントを実行"
	@echo "  make fmt         - コードフォーマット"
	@echo "  make clean       - ビルド成果物を削除"
	@echo "  make install     - バイナリをインストール"
	@echo "  make dev-setup   - 開発環境をセットアップ"
	@echo "  make ci          - CI相当のチェックを実行"

# ビルド
build:
	@echo "Building..."
	go build -v -o todo ./cmd/todo
	go build -v -o tb ./cmd/tb
	go build -v -o td ./cmd/td
	@echo "Build complete!"

# テスト
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete!"

# テストカバレッジ表示
coverage: test
	go tool cover -html=coverage.out

# リント
lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m
	@echo "Lint complete!"

# フォーマット
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .
	@echo "Format complete!"

# フォーマットチェック（CIで使用）
fmt-check:
	@echo "Checking format..."
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Format check complete!"

# go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete!"

# クリーンアップ
clean:
	@echo "Cleaning..."
	rm -f todo tb td
	rm -f coverage.out
	@echo "Clean complete!"

# インストール
install: build
	@echo "Installing..."
	sudo cp todo /usr/local/bin/
	sudo cp tb /usr/local/bin/
	sudo cp td /usr/local/bin/
	@echo "Install complete!"

# 開発環境セットアップ
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go mod download
	@echo "Development environment setup complete!"

# CI相当のチェック（ローカルで実行）
ci: fmt-check vet lint test
	@echo "All CI checks passed!"

# 依存関係の更新
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "Dependencies updated!"

# 脆弱性チェック
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...
	@echo "Vulnerability check complete!"
