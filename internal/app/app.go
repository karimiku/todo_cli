package app

import (
	"context"
	"fmt"
	"io"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/commands"
	"github.com/kamiriku/todo_cli/internal/config"
	"github.com/kamiriku/todo_cli/internal/storage"
)

var binaryDisplayNames = []string{"todo", "tb", "td"}

// Run bootstraps the CLI application using the provided binary name and arguments.
func Run(binaryName string, args []string, stdout, stderr io.Writer) int {
	dbPath, err := config.DatabasePath()
	if err != nil {
		fmt.Fprintf(stderr, "データベースパスの決定に失敗しました: %v\n", err)
		return 1
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		fmt.Fprintf(stderr, "データベースの初期化に失敗しました: %v\n", err)
		return 1
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			fmt.Fprintf(stderr, "データベースのクローズに失敗しました: %v\n", closeErr)
		}
	}()

	app := cli.NewApp(store, stdout, stderr, binaryName, binaryDisplayNames)
	if err := registerCommands(app); err != nil {
		fmt.Fprintf(stderr, "コマンド登録に失敗しました: %v\n", err)
		return 1
	}

	return app.Run(context.Background(), args)
}

func registerCommands(app *cli.App) error {
	commandsToRegister := []cli.Command{
		commands.NewAddCommand(),
		commands.NewListCommand(),
		commands.NewNextCommand(),
		commands.NewFocusCommand(),
		commands.NewMoveCommand(),
		commands.NewEditCommand(),
		commands.NewDeleteCommand(),
		commands.NewBarCommand(),
	}

	for _, command := range commandsToRegister {
		if err := app.Register(command); err != nil {
			return err
		}
	}

	return nil
}
