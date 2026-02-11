// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"context"
	"encoding/json"
	"fmt"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ShellyConnector provides connectivity to Shelly devices.
// This interface is implemented by *shelly.BackupConnector.
type ShellyConnector interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
	WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
	IsGen1Device(ctx context.Context, identifier string) (bool, error)
	DeviceInfo(ctx context.Context, identifier string) (*DeviceInfoResult, error)
	GetConfig(ctx context.Context, identifier string) (map[string]any, error)
	ListScripts(ctx context.Context, identifier string) ([]ScriptInfoResult, error)
	ListSchedules(ctx context.Context, identifier string) ([]ScheduleJobResult, error)
	ListWebhooks(ctx context.Context, identifier string) ([]WebhookInfoResult, error)
}

// DeviceInfoResult holds device info returned by the connector.
type DeviceInfoResult struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
}

// ScriptInfoResult represents script information from a device.
type ScriptInfoResult struct {
	ID      int
	Name    string
	Enable  bool
	Running bool
}

// ScheduleJobResult represents a schedule job from a device.
type ScheduleJobResult struct {
	ID       int
	Enable   bool
	Timespec string
	Calls    []ScheduleCallResult
}

// ScheduleCallResult represents an RPC call in a schedule.
type ScheduleCallResult struct {
	Method string
	Params map[string]any
}

// WebhookInfoResult represents webhook information from a device.
type WebhookInfoResult struct {
	ID     int
	Cid    int
	Enable bool
	Event  string
	Name   string
	URLs   []string
}

// Service provides backup and restore operations for Shelly devices.
type Service struct {
	connector ShellyConnector
}

// NewService creates a new backup service.
func NewService(connector ShellyConnector) *Service {
	return &Service{connector: connector}
}

// CreateBackup creates a complete backup of a device using shelly-go backup.Manager.
// Gen1 devices are handled via HTTP REST calls; Gen2+ via RPC.
func (s *Service) CreateBackup(ctx context.Context, identifier string, opts Options) (*DeviceBackup, error) {
	// Check if this is a Gen1 device
	isGen1, err := s.connector.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		return s.createGen1Backup(ctx, identifier)
	}

	return s.createGen2Backup(ctx, identifier, opts)
}

// createGen1Backup creates a backup from a Gen1 device via HTTP REST.
func (s *Service) createGen1Backup(ctx context.Context, identifier string) (*DeviceBackup, error) {
	var result *DeviceBackup

	err := s.connector.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		bkp, err := exportGen1(ctx, conn)
		if err != nil {
			return err
		}
		result = bkp
		return nil
	})

	return result, err
}

// createGen2Backup creates a backup from a Gen2+ device via RPC.
func (s *Service) createGen2Backup(ctx context.Context, identifier string, opts Options) (*DeviceBackup, error) {
	var result *DeviceBackup

	err := s.connector.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := shellybackup.New(conn.RPCClient())

		// Handle encrypted vs. regular backup
		if opts.Password != "" {
			return fmt.Errorf("encrypted backups not supported via service layer; use backup create command with --encrypt flag")
		}

		// Create regular backup
		data, err := mgr.Export(ctx, opts.ToExportOptions())
		if err != nil {
			return fmt.Errorf("failed to export backup: %w", err)
		}

		// Parse the backup
		var bkp shellybackup.Backup
		if err := json.Unmarshal(data, &bkp); err != nil {
			return fmt.Errorf("failed to parse backup: %w", err)
		}

		result = &DeviceBackup{Backup: &bkp}
		return nil
	})

	return result, err
}

// RestoreBackup restores a backup to a device using shelly-go backup.Manager.
// Gen1 devices are handled via HTTP REST calls; Gen2+ via RPC.
func (s *Service) RestoreBackup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	// Check if this is a Gen1 device
	isGen1, err := s.connector.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		return s.restoreGen1Backup(ctx, identifier, deviceBackup, opts)
	}

	return s.restoreGen2Backup(ctx, identifier, deviceBackup, opts)
}

// restoreGen1Backup restores a backup to a Gen1 device via HTTP REST.
func (s *Service) restoreGen1Backup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	var result *RestoreResult

	err := s.connector.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		r, err := restoreGen1(ctx, conn, deviceBackup, opts)
		if err != nil {
			return err
		}
		result = r
		return nil
	})

	return result, err
}

// restoreGen2Backup restores a backup to a Gen2+ device via RPC.
func (s *Service) restoreGen2Backup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	var result *RestoreResult

	err := s.connector.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := shellybackup.New(conn.RPCClient())

		// Serialize the backup
		data, err := json.Marshal(deviceBackup.Backup)
		if err != nil {
			return fmt.Errorf("failed to serialize backup: %w", err)
		}

		// Restore using shelly-go
		restoreResult, err := mgr.Restore(ctx, data, opts.ToRestoreOptions())
		if err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		// Convert result
		result = &RestoreResult{
			Success:         restoreResult.Success,
			RestartRequired: restoreResult.RestartRequired,
			Warnings:        restoreResult.Warnings,
		}

		// Count restored items from backup
		if restoreResult.Success {
			UpdateResultCounts(result, deviceBackup.Backup)
		}

		return nil
	})

	return result, err
}

// CompareBackup compares a backup with a device's current state.
// Gen1 comparison is limited since Gen1 doesn't expose structured config via RPC.
func (s *Service) CompareBackup(ctx context.Context, identifier string, deviceBackup *DeviceBackup) (*model.BackupDiff, error) {
	// Check if Gen1 — comparison is limited for Gen1 devices
	isGen1, err := s.connector.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		return s.compareGen1Backup(ctx, identifier, deviceBackup)
	}

	return s.compareGen2Backup(ctx, identifier, deviceBackup)
}

// compareGen1Backup performs a basic comparison for Gen1 devices.
// Gen1 doesn't have structured RPC config, so we compare device info only.
func (s *Service) compareGen1Backup(_ context.Context, _ string, _ *DeviceBackup) (*model.BackupDiff, error) {
	diff := &model.BackupDiff{}
	diff.Warnings = append(diff.Warnings, "detailed comparison is not available for Gen1 devices; restore will apply all settings")
	return diff, nil
}

// compareGen2Backup compares a Gen2+ backup with the device's current state.
func (s *Service) compareGen2Backup(ctx context.Context, identifier string, deviceBackup *DeviceBackup) (*model.BackupDiff, error) {
	diff := &model.BackupDiff{}

	// Get current configuration
	currentConfig, err := s.connector.GetConfig(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Parse backup config
	var backupConfig map[string]any
	if err := json.Unmarshal(deviceBackup.Config, &backupConfig); err != nil {
		return nil, fmt.Errorf("failed to parse backup config: %w", err)
	}

	// Compare configurations
	diff.ConfigDiffs = CompareConfigs(currentConfig, backupConfig)

	// Get current scripts
	if deviceBackup.Scripts != nil {
		currentScripts, err := s.connector.ListScripts(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare scripts: %v", err))
		} else {
			backupScripts := ConvertBackupScripts(deviceBackup.Scripts)
			diff.ScriptDiffs = compareScripts(currentScripts, backupScripts)
		}
	}

	// Get current schedules
	if deviceBackup.Schedules != nil {
		currentSchedules, err := s.connector.ListSchedules(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare schedules: %v", err))
		} else {
			backupSchedules := ConvertBackupSchedules(deviceBackup.Schedules)
			diff.ScheduleDiffs = compareSchedules(currentSchedules, backupSchedules)
		}
	}

	// Get current webhooks
	if deviceBackup.Webhooks != nil {
		currentWebhooks, err := s.connector.ListWebhooks(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare webhooks: %v", err))
		} else {
			backupWebhooks := ConvertBackupWebhooks(deviceBackup.Webhooks)
			diff.WebhookDiffs = compareWebhooks(currentWebhooks, backupWebhooks)
		}
	}

	return diff, nil
}

// LoadMigrationSource loads a backup from either a file or device.
// Returns the backup, source type, and any error.
func (s *Service) LoadMigrationSource(ctx context.Context, source string) (bkp *DeviceBackup, sourceType MigrationSource, err error) {
	if IsFile(source) {
		bkp, err = LoadAndValidate(source)
		if err != nil {
			return nil, "", err
		}
		return bkp, SourceFile, nil
	}
	bkp, err = s.CreateBackup(ctx, source, Options{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source device: %w", err)
	}
	return bkp, SourceDevice, nil
}

// CheckMigrationCompatibility checks if the backup is compatible with the target device.
// Returns an error describing the incompatibility if force is false and devices don't match.
func (s *Service) CheckMigrationCompatibility(ctx context.Context, bkp *DeviceBackup, target string, force bool) error {
	targetInfo, err := s.connector.DeviceInfo(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to get target device info: %w", err)
	}

	if !force && bkp.Device().Model != targetInfo.Model {
		return &CompatibilityError{
			SourceModel: bkp.Device().Model,
			TargetModel: targetInfo.Model,
		}
	}
	return nil
}
