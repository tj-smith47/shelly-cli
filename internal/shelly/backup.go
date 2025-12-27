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
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// CreateBackup creates a complete backup of a device using shelly-go backup.Manager.
func (s *Service) CreateBackup(ctx context.Context, identifier string, opts backup.Options) (*backup.DeviceBackup, error) {
	var result *backup.DeviceBackup

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := shellybackup.New(conn.RPCClient())

		// Handle encrypted vs. regular backup
		if opts.Password != "" {
			// For now, encrypted backups not supported via this method
			// User should use backup create command with --encrypt flag
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

		result = &backup.DeviceBackup{Backup: &bkp}
		return nil
	})

	return result, err
}

// RestoreBackup restores a backup to a device using shelly-go backup.Manager.
func (s *Service) RestoreBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup, opts backup.RestoreOptions) (*backup.RestoreResult, error) {
	var result *backup.RestoreResult

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
		result = &backup.RestoreResult{
			Success:         restoreResult.Success,
			RestartRequired: restoreResult.RestartRequired,
			Warnings:        restoreResult.Warnings,
		}

		// Count restored items from backup
		if restoreResult.Success {
			backup.UpdateResultCounts(result, deviceBackup.Backup)
		}

		return nil
	})

	return result, err
}

// CompareBackup compares a backup with a device's current state.
func (s *Service) CompareBackup(ctx context.Context, identifier string, deviceBackup *backup.DeviceBackup) (*model.BackupDiff, error) {
	diff := &model.BackupDiff{}

	// Get current configuration
	currentConfig, err := s.GetConfig(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Parse backup config
	var backupConfig map[string]any
	if err := json.Unmarshal(deviceBackup.Config, &backupConfig); err != nil {
		return nil, fmt.Errorf("failed to parse backup config: %w", err)
	}

	// Compare configurations
	diff.ConfigDiffs = compareConfigs(currentConfig, backupConfig)

	// Get current scripts
	if deviceBackup.Scripts != nil {
		currentScripts, err := s.ListScripts(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare scripts: %v", err))
		} else {
			backupScripts := convertBackupScripts(deviceBackup.Scripts)
			diff.ScriptDiffs = compareScripts(currentScripts, backupScripts)
		}
	}

	// Get current schedules
	if deviceBackup.Schedules != nil {
		currentSchedules, err := s.ListSchedules(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare schedules: %v", err))
		} else {
			backupSchedules := convertBackupSchedules(deviceBackup.Schedules)
			diff.ScheduleDiffs = compareSchedules(currentSchedules, backupSchedules)
		}
	}

	// Get current webhooks
	if deviceBackup.Webhooks != nil {
		currentWebhooks, err := s.ListWebhooks(ctx, identifier)
		if err != nil {
			diff.Warnings = append(diff.Warnings, fmt.Sprintf("could not compare webhooks: %v", err))
		} else {
			backupWebhooks := convertBackupWebhooks(deviceBackup.Webhooks)
			diff.WebhookDiffs = compareWebhooks(currentWebhooks, backupWebhooks)
		}
	}

	return diff, nil
}

// Helper functions for converting backup data structures

func convertBackupScripts(scripts []*shellybackup.Script) []backup.Script {
	result := make([]backup.Script, len(scripts))
	for i, s := range scripts {
		result[i] = backup.Script{
			ID:     s.ID,
			Name:   s.Name,
			Enable: s.Enable,
			Code:   s.Code,
		}
	}
	return result
}

func convertBackupSchedules(data json.RawMessage) []backup.Schedule {
	var schedData struct {
		Jobs []struct {
			Enable   bool                  `json:"enable"`
			Timespec string                `json:"timespec"`
			Calls    []backup.ScheduleCall `json:"calls"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(data, &schedData); err != nil {
		return nil
	}

	result := make([]backup.Schedule, len(schedData.Jobs))
	for i, j := range schedData.Jobs {
		result[i] = backup.Schedule{
			Enable:   j.Enable,
			Timespec: j.Timespec,
			Calls:    j.Calls,
		}
	}
	return result
}

func convertBackupWebhooks(data json.RawMessage) []WebhookInfo {
	var whData struct {
		Hooks []WebhookInfo `json:"hooks"`
	}
	if err := json.Unmarshal(data, &whData); err != nil {
		return nil
	}
	return whData.Hooks
}

// Comparison functions

func compareConfigs(current, bkup map[string]any) []model.ConfigDiff {
	var diffs []model.ConfigDiff

	// Check for keys in backup that differ from current
	for key, backupVal := range bkup {
		currentVal, exists := current[key]
		if !exists {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				NewValue: backupVal,
				DiffType: model.DiffAdded,
			})
		} else if !utils.DeepEqualJSON(currentVal, backupVal) {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				OldValue: currentVal,
				NewValue: backupVal,
				DiffType: model.DiffChanged,
			})
		}
	}

	// Check for keys in current that are not in backup
	for key, currentVal := range current {
		if _, exists := bkup[key]; !exists {
			diffs = append(diffs, model.ConfigDiff{
				Path:     key,
				OldValue: currentVal,
				DiffType: model.DiffRemoved,
			})
		}
	}

	return diffs
}

func compareScripts(current []ScriptInfo, bkup []backup.Script) []model.ScriptDiff {
	var diffs []model.ScriptDiff

	currentMap := make(map[string]ScriptInfo)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	backupMap := make(map[string]backup.Script)
	for _, s := range bkup {
		backupMap[s.Name] = s
	}

	// Check backup scripts
	for name, backupScript := range backupMap {
		if _, exists := currentMap[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffAdded,
				Details:  "script will be created",
			})
		} else {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffChanged,
				Details:  fmt.Sprintf("enable: %v", backupScript.Enable),
			})
		}
	}

	// Check for scripts not in backup
	for name := range currentMap {
		if _, exists := backupMap[name]; !exists {
			diffs = append(diffs, model.ScriptDiff{
				Name:     name,
				DiffType: model.DiffRemoved,
				Details:  "script not in backup (will not be deleted)",
			})
		}
	}

	return diffs
}

func compareSchedules(current []ScheduleJob, bkup []backup.Schedule) []model.ScheduleDiff {
	var diffs []model.ScheduleDiff

	// Simple comparison by timespec
	currentTimespecs := make(map[string]bool)
	for _, s := range current {
		currentTimespecs[s.Timespec] = true
	}

	backupTimespecs := make(map[string]bool)
	for _, s := range bkup {
		backupTimespecs[s.Timespec] = true
	}

	for _, s := range bkup {
		if !currentTimespecs[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffAdded,
				Details:  fmt.Sprintf("enable: %v", s.Enable),
			})
		}
	}

	for _, s := range current {
		if !backupTimespecs[s.Timespec] {
			diffs = append(diffs, model.ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: model.DiffRemoved,
				Details:  "schedule not in backup",
			})
		}
	}

	return diffs
}

func compareWebhooks(current, bkup []WebhookInfo) []model.WebhookDiff {
	var diffs []model.WebhookDiff

	currentMap := make(map[string]WebhookInfo)
	for _, w := range current {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		currentMap[key] = w
	}

	backupMap := make(map[string]WebhookInfo)
	for _, w := range bkup {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		backupMap[key] = w
	}

	for key, backupWh := range backupMap {
		if _, exists := currentMap[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    backupWh.Event,
				Name:     backupWh.Name,
				DiffType: model.DiffAdded,
				Details:  fmt.Sprintf("urls: %v", backupWh.URLs),
			})
		}
	}

	for key, currentWh := range currentMap {
		if _, exists := backupMap[key]; !exists {
			diffs = append(diffs, model.WebhookDiff{
				Event:    currentWh.Event,
				Name:     currentWh.Name,
				DiffType: model.DiffRemoved,
				Details:  "webhook not in backup",
			})
		}
	}

	return diffs
}

// LoadMigrationSource loads a backup from either a file or device.
// Returns the backup, source type, and any error.
func (s *Service) LoadMigrationSource(ctx context.Context, source string) (bkp *backup.DeviceBackup, sourceType backup.MigrationSource, err error) {
	if backup.IsFile(source) {
		bkp, err = backup.LoadAndValidate(source)
		if err != nil {
			return nil, "", err
		}
		return bkp, backup.SourceFile, nil
	}
	bkp, err = s.CreateBackup(ctx, source, backup.Options{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source device: %w", err)
	}
	return bkp, backup.SourceDevice, nil
}

// CheckMigrationCompatibility checks if the backup is compatible with the target device.
// Returns an error describing the incompatibility if force is false and devices don't match.
func (s *Service) CheckMigrationCompatibility(ctx context.Context, bkp *backup.DeviceBackup, target string, force bool) error {
	targetInfo, err := s.DeviceInfo(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to get target device info: %w", err)
	}

	if !force && bkp.Device().Model != targetInfo.Model {
		return &backup.CompatibilityError{
			SourceModel: bkp.Device().Model,
			TargetModel: targetInfo.Model,
		}
	}
	return nil
}
