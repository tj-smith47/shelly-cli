package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveSyncConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	cfg := map[string]any{
		"setting1": "value1",
		"setting2": float64(42),
		"setting3": true,
	}

	if err := SaveSyncConfig(tmpDir, "device1", cfg); err != nil {
		t.Fatalf("SaveSyncConfig() error: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tmpDir, "device1.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", expectedPath)
	}
}

func TestLoadSyncConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Save a config first
	original := map[string]any{
		"setting1": "value1",
		"setting2": float64(42),
	}
	if err := SaveSyncConfig(tmpDir, "device1", original); err != nil {
		t.Fatalf("SaveSyncConfig() error: %v", err)
	}

	// Load it back
	loaded, err := LoadSyncConfig(tmpDir, "device1.json")
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

func TestLoadSyncConfig_NotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	_, err := LoadSyncConfig(tmpDir, "nonexistent.json")
	if err == nil {
		t.Error("expected error loading nonexistent config")
	}
}

func TestLoadSyncConfig_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := LoadSyncConfig(tmpDir, "invalid.json")
	if err == nil {
		t.Error("expected error loading invalid JSON")
	}
}

func TestSaveSyncConfig_MarshalError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Channels cannot be marshaled to JSON
	cfg := map[string]any{
		"channel": make(chan int),
	}

	err := SaveSyncConfig(tmpDir, "device1", cfg)
	if err == nil {
		t.Error("expected error marshaling unmarshalable value")
	}
}
