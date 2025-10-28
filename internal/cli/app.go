package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kamiriku/todo_cli/internal/storage"
)

// App manages command registration and dispatch.
type App struct {
	store  *storage.Store
	stdout io.Writer
	stderr io.Writer
	byName map[string]Command
	lookup map[string]string
}

// NewApp constructs a new App with the provided dependencies.
func NewApp(store *storage.Store, stdout, stderr io.Writer) *App {
	return &App{
		store:  store,
		stdout: stdout,
		stderr: stderr,
		byName: make(map[string]Command),
		lookup: make(map[string]string),
	}
}

// Register adds a command to the dispatch table.
func (a *App) Register(cmd Command) error {
	name := strings.TrimSpace(cmd.Name())
	if name == "" {
		return fmt.Errorf("command name is required")
	}

	canonical := strings.ToLower(name)
	if _, exists := a.lookup[canonical]; exists {
		return fmt.Errorf("command %q already registered", name)
	}

	a.byName[canonical] = cmd
	a.lookup[canonical] = canonical

	for _, alias := range cmd.Aliases() {
		trimmed := strings.TrimSpace(alias)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := a.lookup[key]; exists {
			return fmt.Errorf("command alias %q already registered", alias)
		}
		a.lookup[key] = canonical
	}

	return nil
}

// Run resolves and executes the requested command.
func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 1
	}

	name := strings.ToLower(args[0])
	if name == "help" {
		a.printUsage()
		return 0
	}

	canonical, ok := a.lookup[name]
	if !ok {
		fmt.Fprintf(a.stderr, "不明なコマンドです: %s\n", args[0])
		a.printUsage()
		return 1
	}

	cmd := a.byName[canonical]
	cmdCtx := &CommandContext{
		Context: ctx,
		Store:   a.store,
		Stdout:  a.stdout,
		Stderr:  a.stderr,
	}

	if err := cmd.Run(cmdCtx, args[1:]); err != nil {
		fmt.Fprintf(a.stderr, "%s: %v\n", cmd.Name(), err)
		return 1
	}

	return 0
}

func (a *App) printUsage() {
	fmt.Fprintln(a.stdout, "usage: todo <command> [arguments]")

	if len(a.byName) == 0 {
		fmt.Fprintln(a.stdout, "利用可能なコマンドはまだ登録されていません。")
		return
	}

	fmt.Fprintln(a.stdout, "\navailable commands:")
	names := make([]string, 0, len(a.byName))
	for _, cmd := range a.byName {
		names = append(names, cmd.Name())
	}

	sort.Strings(names)
	for _, name := range names {
		cmd := a.byName[strings.ToLower(name)]
		fmt.Fprintf(a.stdout, "  %-10s %s\n", name, cmd.Description())
	}
}
