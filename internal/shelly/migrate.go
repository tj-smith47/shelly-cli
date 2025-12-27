// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// MigrationOptions configures migration between devices.
type MigrationOptions struct {
	// IncludeWiFi migrates WiFi configuration (security consideration).
	IncludeWiFi bool
	// IncludeCloud migrates Cloud configuration.
	IncludeCloud bool
	// IncludeMQTT migrates MQTT configuration.
	IncludeMQTT bool
	// IncludeBLE migrates BLE configuration.
	IncludeBLE bool
	// IncludeSchedules migrates schedules.
	IncludeSchedules bool
	// IncludeWebhooks migrates webhooks.
	IncludeWebhooks bool
	// IncludeScripts migrates scripts.
	IncludeScripts bool
	// IncludeKVS migrates KVS data.
	IncludeKVS bool
	// RebootAfter reboots the target device after migration.
	RebootAfter bool
	// DryRun simulates the migration without making changes.
	DryRun bool
	// AllowDifferentModels allows migration between different device models.
	AllowDifferentModels bool
	// AllowDifferentGenerations allows migration between different generations.
	AllowDifferentGenerations bool
}

// DefaultMigrationOptions returns default migration options.
func DefaultMigrationOptions() *MigrationOptions {
	return &MigrationOptions{
		IncludeWiFi:               false, // Security: requires explicit opt-in
		IncludeCloud:              true,
		IncludeMQTT:               true,
		IncludeBLE:                true,
		IncludeSchedules:          true,
		IncludeWebhooks:           true,
		IncludeScripts:            true,
		IncludeKVS:                true,
		RebootAfter:               true,
		DryRun:                    false,
		AllowDifferentModels:      false,
		AllowDifferentGenerations: false,
	}
}

// toMigrationOptions converts to shelly-go migration options.
func (o *MigrationOptions) toMigrationOptions() *shellybackup.MigrationOptions {
	return &shellybackup.MigrationOptions{
		IncludeWiFi:      o.IncludeWiFi,
		IncludeCloud:     o.IncludeCloud,
		IncludeMQTT:      o.IncludeMQTT,
		IncludeBLE:       o.IncludeBLE,
		IncludeSchedules: o.IncludeSchedules,
		IncludeWebhooks:  o.IncludeWebhooks,
		IncludeScripts:   o.IncludeScripts,
		IncludeKVS:       o.IncludeKVS,
		RebootAfter:      o.RebootAfter,
		DryRun:           o.DryRun,
	}
}

// MigrationResult contains the result of a migration operation.
type MigrationResult struct {
	SourceDevice       *backup.DeviceInfo
	TargetDevice       *backup.DeviceInfo
	ComponentsMigrated []string
	Warnings           []string
	Errors             []error
	Success            bool
	RestartRequired    bool
	DurationSeconds    float64
}

// Migrate migrates configuration from source device to target device.
func (s *Service) Migrate(ctx context.Context, sourceIdentifier, targetIdentifier string, opts *MigrationOptions, progressCallback func(string, float64)) (*MigrationResult, error) {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	var sourceConn, targetConn *client.Client
	var result *MigrationResult

	// Get source connection
	err := s.WithConnection(ctx, sourceIdentifier, func(conn *client.Client) error {
		sourceConn = conn
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source device: %w", err)
	}

	// Get target connection
	err = s.WithConnection(ctx, targetIdentifier, func(conn *client.Client) error {
		targetConn = conn
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target device: %w", err)
	}

	// Create migrator
	migrator := shellybackup.NewMigrator(sourceConn.RPCClient(), targetConn.RPCClient())
	migrator.AllowDifferentModels = opts.AllowDifferentModels
	migrator.AllowDifferentGenerations = opts.AllowDifferentGenerations

	// Set progress callback if provided
	if progressCallback != nil {
		migrator.OnProgress = progressCallback
	}

	// Perform migration
	migResult, err := migrator.Migrate(ctx, opts.toMigrationOptions())
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Convert result
	result = &MigrationResult{
		ComponentsMigrated: migResult.ComponentsMigrated,
		Warnings:           migResult.Warnings,
		Errors:             migResult.Errors,
		Success:            migResult.Success,
		RestartRequired:    migResult.RestartRequired,
		DurationSeconds:    migResult.Duration().Seconds(),
	}

	// Convert device info
	if migResult.SourceDevice != nil {
		result.SourceDevice = &backup.DeviceInfo{
			ID:         migResult.SourceDevice.ID,
			Name:       migResult.SourceDevice.Name,
			Model:      migResult.SourceDevice.Model,
			Generation: migResult.SourceDevice.Generation,
			FWVersion:  migResult.SourceDevice.Version,
			MAC:        migResult.SourceDevice.MAC,
		}
	}
	if migResult.TargetDevice != nil {
		result.TargetDevice = &backup.DeviceInfo{
			ID:         migResult.TargetDevice.ID,
			Name:       migResult.TargetDevice.Name,
			Model:      migResult.TargetDevice.Model,
			Generation: migResult.TargetDevice.Generation,
			FWVersion:  migResult.TargetDevice.Version,
			MAC:        migResult.TargetDevice.MAC,
		}
	}

	return result, nil
}

// MigrationValidation contains migration validation results.
type MigrationValidation struct {
	SourceDevice *backup.DeviceInfo
	TargetDevice *backup.DeviceInfo
	Warnings     []string
	Errors       []string
	Valid        bool
}

// ValidateMigration checks if migration between two devices is possible.
func (s *Service) ValidateMigration(ctx context.Context, sourceIdentifier, targetIdentifier string, opts *MigrationOptions) (*MigrationValidation, error) {
	if opts == nil {
		opts = DefaultMigrationOptions()
	}

	var sourceConn, targetConn *client.Client

	// Get source connection
	err := s.WithConnection(ctx, sourceIdentifier, func(conn *client.Client) error {
		sourceConn = conn
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to source device: %w", err)
	}

	// Get target connection
	err = s.WithConnection(ctx, targetIdentifier, func(conn *client.Client) error {
		targetConn = conn
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to target device: %w", err)
	}

	// Create migrator
	migrator := shellybackup.NewMigrator(sourceConn.RPCClient(), targetConn.RPCClient())
	migrator.AllowDifferentModels = opts.AllowDifferentModels
	migrator.AllowDifferentGenerations = opts.AllowDifferentGenerations

	// Validate
	validation, err := migrator.ValidateMigration(ctx)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert result
	result := &MigrationValidation{
		Warnings: validation.Warnings,
		Errors:   validation.Errors,
		Valid:    validation.Valid,
	}

	// Convert device info
	if validation.SourceDevice != nil {
		result.SourceDevice = &backup.DeviceInfo{
			ID:         validation.SourceDevice.ID,
			Name:       validation.SourceDevice.Name,
			Model:      validation.SourceDevice.Model,
			Generation: validation.SourceDevice.Generation,
			FWVersion:  validation.SourceDevice.Version,
			MAC:        validation.SourceDevice.MAC,
		}
	}
	if validation.TargetDevice != nil {
		result.TargetDevice = &backup.DeviceInfo{
			ID:         validation.TargetDevice.ID,
			Name:       validation.TargetDevice.Name,
			Model:      validation.TargetDevice.Model,
			Generation: validation.TargetDevice.Generation,
			FWVersion:  validation.TargetDevice.Version,
			MAC:        validation.TargetDevice.MAC,
		}
	}

	return result, nil
}

// MigrateFromBackup migrates configuration from a backup file to a device.
func (s *Service) MigrateFromBackup(ctx context.Context, backupData []byte, targetIdentifier string, opts *backup.RestoreOptions) (*backup.RestoreResult, error) {
	// Parse backup
	bkup, err := backup.Validate(backupData)
	if err != nil {
		return nil, fmt.Errorf("invalid backup: %w", err)
	}

	// Restore to target device
	return s.RestoreBackup(ctx, targetIdentifier, bkup, *opts)
}

// DiffBackups compares two backups and returns differences.
func (s *Service) DiffBackups(backup1, backup2 *backup.DeviceBackup) (*model.BackupDiff, error) {
	diff := &model.BackupDiff{}

	// Parse configs
	var config1, config2 map[string]any
	if err := json.Unmarshal(backup1.Config, &config1); err != nil {
		return nil, fmt.Errorf("failed to parse backup1 config: %w", err)
	}
	if err := json.Unmarshal(backup2.Config, &config2); err != nil {
		return nil, fmt.Errorf("failed to parse backup2 config: %w", err)
	}

	// Compare configurations
	diff.ConfigDiffs = compareConfigs(config1, config2)

	// Compare scripts
	if backup1.Scripts != nil && backup2.Scripts != nil {
		scripts1 := convertBackupScripts(backup1.Scripts)
		scripts2 := convertBackupScripts(backup2.Scripts)
		diff.ScriptDiffs = compareScripts2(scripts1, scripts2)
	}

	// Compare schedules
	if backup1.Schedules != nil && backup2.Schedules != nil {
		schedules1 := convertBackupSchedules(backup1.Schedules)
		schedules2 := convertBackupSchedules(backup2.Schedules)
		diff.ScheduleDiffs = compareSchedules2(schedules1, schedules2)
	}

	// Compare webhooks
	if backup1.Webhooks != nil && backup2.Webhooks != nil {
		webhooks1 := convertBackupWebhooks(backup1.Webhooks)
		webhooks2 := convertBackupWebhooks(backup2.Webhooks)
		diff.WebhookDiffs = compareWebhooks2(webhooks1, webhooks2)
	}

	return diff, nil
}

// Helper comparison functions for backup-to-backup comparisons

func compareScripts2(scripts1, scripts2 []backup.Script) []model.ScriptDiff {
	var diffs []model.ScriptDiff

	map1 := make(map[string]backup.Script)
	for _, s := range scripts1 {
		map1[s.Name] = s
	}

	map2 := make(map[string]backup.Script)
	for _, s := range scripts2 {
		map2[s.Name] = s
	}

	// Check scripts in backup2
	for name, script2 := range map2 {
		if _, exists := map1[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffAdded,
				Details:  "script in backup2 only",
			})
		} else {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffChanged,
				Details:  fmt.Sprintf("enable: %v", script2.Enable),
			})
		}
	}

	// Check for scripts in backup1 not in backup2
	for name := range map1 {
		if _, exists := map2[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffRemoved,
				Details:  "script in backup1 only",
			})
		}
	}

	return diffs
}

func compareSchedules2(schedules1, schedules2 []backup.Schedule) []model.ScheduleDiff {
	var diffs []model.ScheduleDiff

	timespecs1 := make(map[string]bool)
	for _, s := range schedules1 {
		timespecs1[s.Timespec] = true
	}

	timespecs2 := make(map[string]bool)
	for _, s := range schedules2 {
		timespecs2[s.Timespec] = true
	}

	for _, s := range schedules2 {
		if !timespecs1[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffAdded,
				Details:  "schedule in backup2 only",
			})
		}
	}

	for _, s := range schedules1 {
		if !timespecs2[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffRemoved,
				Details:  "schedule in backup1 only",
			})
		}
	}

	return diffs
}

func compareWebhooks2(webhooks1, webhooks2 []WebhookInfo) []model.WebhookDiff {
	var diffs []model.WebhookDiff

	map1 := make(map[string]WebhookInfo)
	for _, w := range webhooks1 {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		map1[key] = w
	}

	map2 := make(map[string]WebhookInfo)
	for _, w := range webhooks2 {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		map2[key] = w
	}

	for key, wh2 := range map2 {
		if _, exists := map1[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    wh2.Event,
				Name:     wh2.Name,
				DiffType: model.DiffAdded,
				Details:  "webhook in backup2 only",
			})
		}
	}

	for key, wh1 := range map1 {
		if _, exists := map2[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    wh1.Event,
				Name:     wh1.Name,
				DiffType: model.DiffRemoved,
				Details:  "webhook in backup1 only",
			})
		}
	}

	return diffs
}
