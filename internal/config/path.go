package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	envTodoHome    = "TODO_HOME"
	envXDGConfig   = "XDG_CONFIG_HOME"
	envAppData     = "APPDATA"
	defaultDirName = "todo"
	defaultDBName  = "todo.db"
	defaultDirPerm = 0o755
)

// DataDir determines the directory used to store application data.
func DataDir() (string, error) {
	if home, ok := os.LookupEnv(envTodoHome); ok {
		trimmed := strings.TrimSpace(home)
		if trimmed != "" {
			return normalizePath(trimmed)
		}
	}

	if xdg := strings.TrimSpace(os.Getenv(envXDGConfig)); xdg != "" {
		dir, err := normalizePath(xdg)
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, defaultDirName), nil
	}

	if runtime.GOOS == "windows" {
		if appData := strings.TrimSpace(os.Getenv(envAppData)); appData != "" {
			dir, err := normalizePath(appData)
			if err != nil {
				return "", err
			}
			return filepath.Join(dir, defaultDirName), nil
		}
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}

	return filepath.Join(configDir, defaultDirName), nil
}

// EnsureDataDir creates the data directory if it does not already exist.
func EnsureDataDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, defaultDirPerm); err != nil {
		return "", fmt.Errorf("create data dir: %w", err)
	}

	return dir, nil
}

// DatabasePath returns the path to the SQLite database file, ensuring the
// parent directory exists.
func DatabasePath() (string, error) {
	dir, err := EnsureDataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, defaultDBName), nil
}

func normalizePath(p string) (string, error) {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve ~ in path: %w", err)
		}

		if p == "~" {
			return home, nil
		}

		trimmed := strings.TrimPrefix(p, "~/")
		return filepath.Join(home, trimmed), nil
	}

	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("make path absolute: %w", err)
	}

	return abs, nil
}
