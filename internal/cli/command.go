package cli

import (
	"context"
	"io"

	"github.com/kamiriku/todo_cli/internal/storage"
)

// Command defines the behaviour required by subcommands handled by the CLI.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Run(ctx *CommandContext, args []string) error
}

// CommandContext carries shared dependencies for command execution.
type CommandContext struct {
	context.Context
	Store  *storage.Store
	Stdout io.Writer
	Stderr io.Writer
}
