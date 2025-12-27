package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GetSyncDir returns the sync directory path, creating it if needed.
func GetSyncDir() (string, error) {
	configDir, err := Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	syncDir := filepath.Join(configDir, "sync")

	// Ensure directory exists
	if err := os.MkdirAll(syncDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create sync directory: %w", err)
	}

	return syncDir, nil
}

// SaveSyncConfig saves a device config to the sync directory.
func SaveSyncConfig(syncDir, device string, cfg map[string]any) error {
	filename := filepath.Join(syncDir, fmt.Sprintf("%s.json", device))
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// LoadSyncConfig loads a device config from the sync directory.
func LoadSyncConfig(syncDir, filename string) (map[string]any, error) {
	fullPath := filepath.Join(syncDir, filename)
	data, err := os.ReadFile(fullPath) //nolint:gosec // User-managed sync directory
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return cfg, nil
}
