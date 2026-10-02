// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// DiffBackups compares two backups and returns differences.
func (s *Service) DiffBackups(backup1, backup2 *backup.DeviceBackup) (*model.BackupDiff, error) {
	return backup.DiffBackups(backup1, backup2)
}
