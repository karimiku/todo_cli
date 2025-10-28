package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDataDirUsesTodoHome(t *testing.T) {
	t.Helper()

	temp := t.TempDir()
	custom := filepath.Join(temp, "mydata")
	t.Setenv("TODO_HOME", "  "+custom+"  ")

	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir returned error: %v", err)
	}

	if dir != custom {
		t.Fatalf("DataDir = %q, want %q", dir, custom)
	}
}

func TestDatabasePathExpandsHome(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tempHome)
	}

	t.Setenv("TODO_HOME", "~/todospace")

	path, err := DatabasePath()
	if err != nil {
		t.Fatalf("DatabasePath returned error: %v", err)
	}

	wantDir := filepath.Join(tempHome, "todospace")
	if !strings.HasPrefix(path, wantDir) {
		t.Fatalf("DatabasePath = %q, expected directory prefix %q", path, wantDir)
	}

	if filepath.Base(path) != defaultDBName {
		t.Fatalf("DatabasePath base = %q, want %q", filepath.Base(path), defaultDBName)
	}

	info, err := os.Stat(wantDir)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", wantDir, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", wantDir)
	}
}

func TestDataDirFallsBackToUserConfigDir(t *testing.T) {
	temp := t.TempDir()
	configRoot := filepath.Join(temp, "config")
	if err := os.MkdirAll(configRoot, 0o755); err != nil {
		t.Fatalf("failed to create fake config dir: %v", err)
	}

	t.Setenv("TODO_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("HOME", temp)
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", configRoot)
	}

	dir, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir returned error: %v", err)
	}

	want := filepath.Join(configRoot, defaultDirName)
	if dir != want {
		t.Fatalf("DataDir = %q, want %q", dir, want)
	}
}

func TestEnsureDataDirCreatesDirectory(t *testing.T) {
	temp := t.TempDir()
	custom := filepath.Join(temp, "nested", "storage")
	t.Setenv("TODO_HOME", custom)

	dir, err := EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir returned error: %v", err)
	}

	if dir != custom {
		t.Fatalf("EnsureDataDir = %q, want %q", dir, custom)
	}

	info, err := os.Stat(custom)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", custom, err)
	}

	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", custom)
	}
}
