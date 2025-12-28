package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShowUpdateNotification_EnvDisabled(t *testing.T) {
	// Set the disable env var
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "1")

	// This should return early without doing anything
	// We just verify it doesn't panic
	ShowUpdateNotification()
}

func TestShowUpdateNotification_NoCache(t *testing.T) {
	// Create a temp directory with no cache
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	// Should not panic with no cache
	ShowUpdateNotification()
}

func TestShowUpdateNotification_DevBuild(t *testing.T) {
	// Create a temp directory with a cache file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	// Create cache with a version
	cachePath := filepath.Join(tmpDir, ".config", "shelly", "cache")
	if err := os.MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cachePath, "latest-version"), []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original version and set to dev
	originalVersion := Version
	Version = "dev"
	defer func() { Version = originalVersion }()

	// Should return early for dev builds without panic
	ShowUpdateNotification()
}

func TestShowUpdateNotification_SkippedCommands(t *testing.T) {
	// Create a temp directory with a cache file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	})

	// Create cache with a newer version
	cachePath := filepath.Join(tmpDir, ".config", "shelly", "cache")
	if err := os.MkdirAll(cachePath, 0o750); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cachePath, "latest-version"), []byte("2.0.0"), 0o600); err != nil {
		t.Fatalf("failed to write cache: %v", err)
	}

	// Save original args
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	// Test skipped commands
	skippedCommands := []string{"version", "update", "completion", "help"}
	//nolint:paralleltest // Subtests cannot run in parallel as they modify global os.Args
	for _, cmd := range skippedCommands {
		t.Run("skip_"+cmd, func(t *testing.T) {
			os.Args = []string{"shelly", cmd}
			// Should return early without panic
			ShowUpdateNotification()
		})
	}
}

func TestShowUpdateNotification_NoArgs(t *testing.T) {
	// Create a temp directory with a cache file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELLY_NO_UPDATE_CHECK", "")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("warning: failed to restore HOME: %v", err)
		}
	}()

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with no args (just program name)
	os.Args = []string{"shelly"}

	// Should not panic with no args
	ShowUpdateNotification()
}
