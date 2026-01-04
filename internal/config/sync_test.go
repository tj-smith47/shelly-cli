package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

const testSyncDir = "/test/sync"

//nolint:paralleltest // Test modifies global state via SetFs
func TestSaveSyncConfig(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir
	cfg := map[string]any{
		"setting1": "value1",
		"setting2": float64(42),
		"setting3": true,
	}

	if err := SaveSyncConfig(syncDir, "device1", cfg); err != nil {
		t.Fatalf("SaveSyncConfig() error: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(syncDir, "device1.json")
	if _, err := Fs().Stat(expectedPath); err != nil {
		t.Errorf("expected file %s to exist: %v", expectedPath, err)
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestLoadSyncConfig(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir

	// Save a config first
	original := map[string]any{
		"setting1": "value1",
		"setting2": float64(42),
	}
	if err := SaveSyncConfig(syncDir, "device1", original); err != nil {
		t.Fatalf("SaveSyncConfig() error: %v", err)
	}

	// Load it back
	loaded, err := LoadSyncConfig(syncDir, "device1.json")
	if err != nil {
		t.Fatalf("LoadSyncConfig() error: %v", err)
	}

	if loaded["setting1"] != "value1" {
		t.Errorf("setting1 = %v, want %q", loaded["setting1"], "value1")
	}
	if loaded["setting2"] != float64(42) {
		t.Errorf("setting2 = %v, want %v", loaded["setting2"], float64(42))
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestLoadSyncConfig_NotFound(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir

	_, err := LoadSyncConfig(syncDir, "nonexistent.json")
	if err == nil {
		t.Error("expected error loading nonexistent config")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestLoadSyncConfig_InvalidJSON(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir
	invalidPath := filepath.Join(syncDir, "invalid.json")

	// Create directory and write invalid JSON
	if err := Fs().MkdirAll(syncDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := afero.WriteFile(Fs(), invalidPath, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := LoadSyncConfig(syncDir, "invalid.json")
	if err == nil {
		t.Error("expected error loading invalid JSON")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestSaveSyncConfig_MarshalError(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir

	// Channels cannot be marshaled to JSON
	cfg := map[string]any{
		"channel": make(chan int),
	}

	err := SaveSyncConfig(syncDir, "device1", cfg)
	if err == nil {
		t.Error("expected error marshaling unmarshalable value")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestGetSyncDir(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	syncDir, err := GetSyncDir()
	if err != nil {
		t.Fatalf("GetSyncDir() error: %v", err)
	}

	if syncDir == "" {
		t.Error("GetSyncDir() returned empty string")
	}

	// Verify directory exists in afero filesystem
	if stat, err := Fs().Stat(syncDir); err != nil {
		t.Errorf("sync directory should exist: %v", err)
	} else if !stat.IsDir() {
		t.Error("sync directory should be a directory")
	}
}

//nolint:paralleltest // Test modifies global state via SetFs
func TestSaveSyncConfig_WriteError(t *testing.T) {
	// Use a read-only filesystem to force write failure
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	SetFs(roFs)
	t.Cleanup(func() { SetFs(nil) })

	syncDir := testSyncDir
	cfg := map[string]any{"test": "value"}

	err := SaveSyncConfig(syncDir, "device1", cfg)
	if err == nil {
		t.Error("expected error writing to read-only filesystem")
	}
}
