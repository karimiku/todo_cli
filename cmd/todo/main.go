package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/config"
	"github.com/kamiriku/todo_cli/internal/storage"
)

func main() {
	os.Exit(run())
}

func run() int {
	dbPath, err := config.DatabasePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "データベースパスの決定に失敗しました: %v\n", err)
		return 1
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "データベースの初期化に失敗しました: %v\n", err)
		return 1
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "データベースのクローズに失敗しました: %v\n", closeErr)
		}
	}()

	app := cli.NewApp(store, os.Stdout, os.Stderr)

	ctx := context.Background()
	args := os.Args[1:]

	return app.Run(ctx, args)
}
