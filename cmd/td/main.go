package main

import (
	"os"
	"path/filepath"

	app "github.com/kamiriku/todo_cli/internal/app"
)

func main() {
	binary := filepath.Base(os.Args[0])
	os.Exit(app.Run(binary, os.Args[1:], os.Stdout, os.Stderr))
}
