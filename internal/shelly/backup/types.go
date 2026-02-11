// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

// DeviceBackup wraps the shelly-go backup.Backup with CLI-specific methods.
type DeviceBackup struct {
	*shellybackup.Backup
}

// Device returns device info from the backup.
func (b *DeviceBackup) Device() DeviceInfo {
	if b.DeviceInfo == nil {
		return DeviceInfo{}
	}
	return DeviceInfo{
		ID:         b.DeviceInfo.ID,
		Name:       b.DeviceInfo.Name,
		Model:      b.DeviceInfo.Model,
		Generation: b.DeviceInfo.Generation,
		FWVersion:  b.DeviceInfo.Version,
		MAC:        b.DeviceInfo.MAC,
	}
}

// Encrypted returns true if backup is encrypted (always false for regular backups).
func (b *DeviceBackup) Encrypted() bool {
	return false
}

// DeviceInfo contains device identification from a backup.
type DeviceInfo struct {
	ID         string
	Name       string
	Model      string
	Generation int
	FWVersion  string
	MAC        string
}

// Options configures backup creation.
type Options struct {
	// SkipScripts excludes scripts from backup.
	SkipScripts bool
	// SkipSchedules excludes schedules from backup.
	SkipSchedules bool
	// SkipWebhooks excludes webhooks from backup.
	SkipWebhooks bool
	// SkipKVS excludes KVS data from backup.
	SkipKVS bool
	// SkipWiFi excludes WiFi configuration from backup (security).
	SkipWiFi bool
	// Password for encryption (empty = no encryption).
	Password string
}

// ToExportOptions converts Options to shelly-go ExportOptions.
// Builds on library defaults and overrides only CLI-controlled fields.
func (o *Options) ToExportOptions() *shellybackup.ExportOptions {
	opts := shellybackup.DefaultExportOptions()
	opts.IncludeWiFi = !o.SkipWiFi
	opts.IncludeWebhooks = !o.SkipWebhooks
	opts.IncludeSchedules = !o.SkipSchedules
	opts.IncludeScripts = !o.SkipScripts
	opts.IncludeKVS = !o.SkipKVS
	return opts
}

// RestoreOptions configures backup restoration.
type RestoreOptions struct {
	// DryRun shows what would be changed without applying.
	DryRun bool
	// SkipNetwork skips WiFi/Ethernet configuration.
	SkipNetwork bool
	// SkipAuth skips authentication configuration.
	SkipAuth bool
	// SkipScripts skips script restoration.
	SkipScripts bool
	// SkipSchedules skips schedule restoration.
	SkipSchedules bool
	// SkipWebhooks skips webhook restoration.
	SkipWebhooks bool
	// SkipKVS skips KVS data restoration.
	SkipKVS bool
	// Password for decryption (required if backup is encrypted).
	Password string
}

// ToRestoreOptions converts RestoreOptions to shelly-go RestoreOptions.
// Builds on library defaults and overrides only CLI-controlled fields.
func (o *RestoreOptions) ToRestoreOptions() *shellybackup.RestoreOptions {
	opts := shellybackup.DefaultRestoreOptions()
	opts.RestoreWiFi = !o.SkipNetwork
	opts.RestoreAuth = !o.SkipAuth
	opts.RestoreWebhooks = !o.SkipWebhooks
	opts.RestoreSchedules = !o.SkipSchedules
	opts.RestoreScripts = !o.SkipScripts
	opts.RestoreKVS = !o.SkipKVS
	opts.DryRun = o.DryRun
	return opts
}

// RestoreResult contains the result of a restore operation.
type RestoreResult struct {
	Success           bool
	ConfigRestored    bool
	ScriptsRestored   int
	SchedulesRestored int
	WebhooksRestored  int
	RestartRequired   bool
	Warnings          []string
}

// Script is a compatibility type for backup scripts.
type Script struct {
	ID     int
	Name   string
	Enable bool
	Code   string
}

// Schedule is a compatibility type for backup schedules.
type Schedule struct {
	Enable   bool
	Timespec string
	Calls    []ScheduleCall
}

// ScheduleCall represents a single call in a schedule.
type ScheduleCall struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// MigrationSource represents where a migration backup came from.
type MigrationSource string

// Migration source constants.
const (
	SourceFile   MigrationSource = "file"
	SourceDevice MigrationSource = "device"
)

// CompatibilityError represents a device type mismatch during migration.
type CompatibilityError struct {
	SourceModel string
	TargetModel string
}

// Error implements error interface.
func (e *CompatibilityError) Error() string {
	return "device type mismatch"
}

// UpdateResultCounts updates the RestoreResult with counts from the backup.
func UpdateResultCounts(result *RestoreResult, deviceBackup *shellybackup.Backup) {
	result.ConfigRestored = true

	// Count scripts
	if deviceBackup.Scripts != nil {
		result.ScriptsRestored = len(deviceBackup.Scripts)
	}

	// Parse and count schedules
	if deviceBackup.Schedules != nil {
		var schedData struct {
			Jobs []json.RawMessage `json:"jobs"`
		}
		if err := json.Unmarshal(deviceBackup.Schedules, &schedData); err == nil {
			result.SchedulesRestored = len(schedData.Jobs)
		}
	}

	// Parse and count webhooks
	if deviceBackup.Webhooks != nil {
		var whData struct {
			Hooks []json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(deviceBackup.Webhooks, &whData); err == nil {
			result.WebhooksRestored = len(whData.Hooks)
		}
	}
}
