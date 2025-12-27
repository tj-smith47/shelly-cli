// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

// Validate validates a backup file structure.
func Validate(data []byte) (*DeviceBackup, error) {
	var bkp shellybackup.Backup
	if err := json.Unmarshal(data, &bkp); err != nil {
		return nil, fmt.Errorf("invalid backup format: %w", err)
	}

	if bkp.Version == 0 {
		return nil, fmt.Errorf("missing or invalid backup version")
	}
	if bkp.DeviceInfo == nil {
		return nil, fmt.Errorf("missing device information")
	}
	if bkp.Config == nil {
		return nil, fmt.Errorf("missing configuration data")
	}

	return &DeviceBackup{Backup: &bkp}, nil
}

// SaveToFile saves backup data to a file.
func SaveToFile(data []byte, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// LoadFromFile loads backup data from a file.
func LoadFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec // User-provided file path
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}
	return data, nil
}

// GenerateFilename generates a backup filename based on device info and timestamp.
func GenerateFilename(deviceName, deviceID string, encrypted bool) string {
	timestamp := time.Now().Format("20060102-150405")
	safeName := strings.ReplaceAll(deviceName, " ", "-")
	if safeName == "" {
		safeName = deviceID
	}

	suffix := ".json"
	if encrypted {
		suffix = ".enc.json"
	}

	return fmt.Sprintf("backup-%s-%s%s", safeName, timestamp, suffix)
}

// IsFile checks if the path looks like a backup file (exists and is a file).
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// LoadAndValidate loads a DeviceBackup from a file and validates it.
func LoadAndValidate(source string) (*DeviceBackup, error) {
	data, err := os.ReadFile(source) //nolint:gosec // G304: source is user-provided CLI argument
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}
	bkp, err := Validate(data)
	if err != nil {
		return nil, fmt.Errorf("invalid backup file: %w", err)
	}
	return bkp, nil
}
