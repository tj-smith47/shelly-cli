// Package export provides export format builders for device data.
package export

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// SanitizeFilename replaces problematic characters in a filename.
func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// WriteBackupFile writes a backup to a file in the specified format.
func WriteBackupFile(bkp *backup.DeviceBackup, filePath, format string) error {
	var data []byte
	var err error

	switch format {
	case FormatYAML, "yml":
		data, err = yaml.Marshal(bkp)
	default:
		data, err = json.MarshalIndent(bkp, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	if err := afero.WriteFile(config.Fs(), filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ScanBackupFiles scans a directory for backup files and returns their info.
func ScanBackupFiles(dir string) ([]model.BackupFileInfo, error) {
	entries, err := afero.ReadDir(config.Fs(), dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	backups := make([]model.BackupFileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !IsBackupFile(name) {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, err := ParseBackupFile(filePath)
		if err != nil {
			continue
		}
		info.Filename = name
		backups = append(backups, info)
	}
	return backups, nil
}

// IsBackupFile checks if a filename has a backup file extension.
func IsBackupFile(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

// ParseBackupFile reads and parses a backup file, returning its metadata.
func ParseBackupFile(filePath string) (model.BackupFileInfo, error) {
	var info model.BackupFileInfo
	fs := config.Fs()

	data, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return info, err
	}

	bkp, err := backup.Validate(data)
	if err != nil {
		return info, err
	}

	stat, err := fs.Stat(filePath)
	if err != nil {
		return info, err
	}

	info.DeviceID = bkp.Device().ID
	info.DeviceModel = bkp.Device().Model
	info.FWVersion = bkp.Device().FWVersion
	info.CreatedAt = bkp.CreatedAt
	info.Encrypted = bkp.Encrypted()
	info.Size = stat.Size()

	return info, nil
}

// MarshalBackup serializes a backup to the specified format.
func MarshalBackup(bkp *backup.DeviceBackup, format string) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(bkp)
	default:
		return json.MarshalIndent(bkp, "", "  ")
	}
}
