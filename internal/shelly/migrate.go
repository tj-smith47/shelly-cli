// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
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
	return backup.DiffBackups(backup1, backup2)
}
