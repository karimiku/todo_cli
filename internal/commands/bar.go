package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/kamiriku/todo_cli/internal/cli"
	"github.com/kamiriku/todo_cli/internal/storage"
)

type barOptions struct {
	headOnly    bool
	maxLen      int
	icon        string
	showTooltip bool
}

// BarCommand outputs data formatted for SwiftBar/xbar integration.
type BarCommand struct{}

// NewBarCommand constructs the bar command.
func NewBarCommand() cli.Command {
	return &BarCommand{}
}

func (c *BarCommand) Name() string {
	return "bar"
}

func (c *BarCommand) Aliases() []string {
	return nil
}

func (c *BarCommand) Description() string {
	return "メニューバー用にタスクを表示します"
}

func (c *BarCommand) Run(ctx *cli.CommandContext, args []string) error {
	opts, err := parseBarOptions(args)
	if err != nil {
		return err
	}
	if !opts.headOnly {
		return fmt.Errorf("usage: %s bar --head-only [--maxlen N] [--icon STR] [--no-tooltip]", ctx.BinaryName)
	}

	task, err := ctx.Store.HeadTask(ctx)
	switch {
	case err == nil:
		line := renderHeadLine(task.Text, opts)
		fmt.Fprintln(ctx.Stdout, line)
		return nil
	case errors.Is(err, storage.ErrNoPendingTasks):
		fmt.Fprintln(ctx.Stdout, "🎉 空っぽ！ | tooltip=おつかれさま")
		return nil
	default:
		reason := sanitizeTooltip(err.Error())
		fmt.Fprintf(ctx.Stdout, "⚠︎ Error | tooltip=%s\n", reason)
		return nil
	}
}

func parseBarOptions(args []string) (barOptions, error) {
	opts := barOptions{
		maxLen:      28,
		icon:        "▶️",
		showTooltip: true,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--head-only":
			opts.headOnly = true
			i++
		case strings.HasPrefix(arg, "--maxlen="):
			value := strings.TrimPrefix(arg, "--maxlen=")
			if err := setMaxLen(&opts, value); err != nil {
				return barOptions{}, err
			}
			i++
		case arg == "--maxlen":
			if i+1 >= len(args) {
				return barOptions{}, fmt.Errorf("--maxlen requires a value")
			}
			if err := setMaxLen(&opts, args[i+1]); err != nil {
				return barOptions{}, err
			}
			i += 2
		case strings.HasPrefix(arg, "--icon="):
			opts.icon = strings.TrimPrefix(arg, "--icon=")
			i++
		case arg == "--icon":
			if i+1 >= len(args) {
				return barOptions{}, fmt.Errorf("--icon requires a value")
			}
			opts.icon = args[i+1]
			i += 2
		case arg == "--no-tooltip":
			opts.showTooltip = false
			i++
		default:
			return barOptions{}, fmt.Errorf("usage: bar --head-only [--maxlen N] [--icon STR] [--no-tooltip]")
		}
	}

	if opts.maxLen < 1 {
		opts.maxLen = 1
	}

	return opts, nil
}

func setMaxLen(opts *barOptions, value string) error {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return fmt.Errorf("--maxlen requires a positive integer")
	}
	opts.maxLen = n
	return nil
}

func renderHeadLine(text string, opts barOptions) string {
	trimmed := truncateText(text, opts.maxLen)
	var builder strings.Builder
	if opts.icon != "" {
		builder.WriteString(opts.icon)
		builder.WriteByte(' ')
	}
	builder.WriteString(trimmed)
	builder.WriteString(" | ")
	if opts.showTooltip {
		builder.WriteString("tooltip=")
		builder.WriteString(sanitizeTooltip(text))
		builder.WriteByte(' ')
	}
	builder.WriteString("bash=todo param1=next terminal=false refresh=1")
	return builder.String()
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}

	truncated := runes[:maxLen]
	for len(truncated) > 0 && unicode.Is(unicode.Mn, truncated[len(truncated)-1]) {
		truncated = truncated[:len(truncated)-1]
	}
	return string(truncated) + "…"
}

func sanitizeTooltip(text string) string {
	cleaned := strings.ReplaceAll(text, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "|", "/")
	return cleaned
}
