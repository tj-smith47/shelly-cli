// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/config"
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
	fs := config.Fs()
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := fs.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := afero.WriteFile(fs, filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// LoadFromFile loads backup data from a file.
func LoadFromFile(filePath string) ([]byte, error) {
	data, err := afero.ReadFile(config.Fs(), filePath)
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

// AutoSavePath returns the auto-generated file path for a backup based on
// device info and format. It creates the backups directory if needed.
func AutoSavePath(bkp *DeviceBackup, format string) (string, error) {
	dir, err := config.BackupsDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine backups directory: %w", err)
	}
	if err := config.Fs().MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create backups directory: %w", err)
	}

	name := sanitizeForPath(bkp.Device().Name)
	if name == "" {
		name = sanitizeForPath(bkp.Device().ID)
	}
	if name == "" {
		name = "backup"
	}
	date := time.Now().Format("2006-01-02")
	return filepath.Join(dir, fmt.Sprintf("%s-%s.%s", name, date, format)), nil
}

// sanitizeForPath replaces filesystem-unsafe characters with underscores.
func sanitizeForPath(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(s)
}

// IsFile checks if the path looks like a backup file (exists and is a file).
func IsFile(path string) bool {
	info, err := config.Fs().Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// LoadAndValidate loads a DeviceBackup from a file and validates it.
func LoadAndValidate(source string) (*DeviceBackup, error) {
	data, err := afero.ReadFile(config.Fs(), source)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}
	bkp, err := Validate(data)
	if err != nil {
		return nil, fmt.Errorf("invalid backup file: %w", err)
	}
	return bkp, nil
}
