// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// DeviceBackup wraps the shelly-go backup.Backup for backward compatibility.
type DeviceBackup struct {
	// Backup is the underlying shelly-go backup
	*backup.Backup
}

// Device returns device info for backward compatibility with existing commands.
func (b *DeviceBackup) Device() BackupDeviceInfo {
	if b.DeviceInfo == nil {
		return BackupDeviceInfo{}
	}
	return BackupDeviceInfo{
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

// BackupDeviceInfo contains device identification for backward compatibility.
type BackupDeviceInfo struct {
	ID         string
	Name       string
	Model      string
	Generation int
	FWVersion  string
	MAC        string
}

// BackupOptions configures backup creation.
type BackupOptions struct {
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

// toExportOptions converts BackupOptions to shelly-go ExportOptions.
func (o *BackupOptions) toExportOptions() *backup.ExportOptions {
	return &backup.ExportOptions{
		IncludeWiFi:       !o.SkipWiFi,
		IncludeCloud:      true,
		IncludeAuth:       true, // Auth metadata only, no passwords
		IncludeBLE:        true,
		IncludeMQTT:       true,
		IncludeWebhooks:   !o.SkipWebhooks,
		IncludeSchedules:  !o.SkipSchedules,
		IncludeScripts:    !o.SkipScripts,
		IncludeKVS:        !o.SkipKVS,
		IncludeComponents: true,
	}
}

// RestoreOptions configures backup restoration.
type RestoreOptions struct {
	// DryRun shows what would be changed without applying.
	DryRun bool
	// SkipNetwork skips WiFi/Ethernet configuration.
	SkipNetwork bool
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

// toRestoreOptions converts RestoreOptions to shelly-go RestoreOptions.
func (o *RestoreOptions) toRestoreOptions() *backup.RestoreOptions {
	return &backup.RestoreOptions{
		RestoreWiFi:       !o.SkipNetwork,
		RestoreCloud:      true,
		RestoreAuth:       false, // Never auto-restore auth for security
		RestoreBLE:        true,
		RestoreMQTT:       true,
		RestoreWebhooks:   !o.SkipWebhooks,
		RestoreSchedules:  !o.SkipSchedules,
		RestoreScripts:    !o.SkipScripts,
		RestoreKVS:        !o.SkipKVS,
		RestoreComponents: true,
		DryRun:            o.DryRun,
		StopScripts:       true,
	}
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

// CreateBackup creates a complete backup of a device using shelly-go backup.Manager.
func (s *Service) CreateBackup(ctx context.Context, identifier string, opts BackupOptions) (*DeviceBackup, error) {
	var result *DeviceBackup

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := backup.New(conn.RPCClient())

		// Handle encrypted vs. regular backup
		if opts.Password != "" {
			// For now, encrypted backups not supported via this method
			// User should use backup create command with --encrypt flag
			return fmt.Errorf("encrypted backups not supported via service layer; use backup create command with --encrypt flag")
		}

		// Create regular backup
		data, err := mgr.Export(ctx, opts.toExportOptions())
		if err != nil {
			return fmt.Errorf("failed to export backup: %w", err)
		}

		// Parse the backup
		var bkp backup.Backup
		if err := json.Unmarshal(data, &bkp); err != nil {
			return fmt.Errorf("failed to parse backup: %w", err)
		}

		result = &DeviceBackup{Backup: &bkp}
		return nil
	})

	return result, err
}

// RestoreBackup restores a backup to a device using shelly-go backup.Manager.
func (s *Service) RestoreBackup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	var result *RestoreResult

	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := backup.New(conn.RPCClient())

		// Serialize the backup
		data, err := json.Marshal(deviceBackup.Backup)
		if err != nil {
			return fmt.Errorf("failed to serialize backup: %w", err)
		}

		// Restore using shelly-go
		restoreResult, err := mgr.Restore(ctx, data, opts.toRestoreOptions())
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
			updateRestoreResultCounts(result, deviceBackup.Backup)
		}

		return nil
	})

	return result, err
}

// updateRestoreResultCounts updates the RestoreResult with counts from the backup.
func updateRestoreResultCounts(result *RestoreResult, deviceBackup *backup.Backup) {
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

// ValidateBackup validates a backup file structure.
func ValidateBackup(data []byte) (*DeviceBackup, error) {
	var bkp backup.Backup
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

// CompareBackup compares a backup with a device's current state.
func (s *Service) CompareBackup(ctx context.Context, identifier string, deviceBackup *DeviceBackup) (*BackupDiff, error) {
	diff := &BackupDiff{}

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
		if err == nil {
			backupScripts := convertBackupScripts(deviceBackup.Scripts)
			diff.ScriptDiffs = compareScripts(currentScripts, backupScripts)
		}
	}

	// Get current schedules
	if deviceBackup.Schedules != nil {
		currentSchedules, err := s.ListSchedules(ctx, identifier)
		if err == nil {
			backupSchedules := convertBackupSchedules(deviceBackup.Schedules)
			diff.ScheduleDiffs = compareSchedules(currentSchedules, backupSchedules)
		}
	}

	// Get current webhooks
	if deviceBackup.Webhooks != nil {
		currentWebhooks, err := s.ListWebhooks(ctx, identifier)
		if err == nil {
			backupWebhooks := convertBackupWebhooks(deviceBackup.Webhooks)
			diff.WebhookDiffs = compareWebhooks(currentWebhooks, backupWebhooks)
		}
	}

	return diff, nil
}

// BackupDiff contains differences between a backup and current device state.
type BackupDiff struct {
	ConfigDiffs   []ConfigDiff
	ScriptDiffs   []ScriptDiff
	ScheduleDiffs []ScheduleDiff
	WebhookDiffs  []WebhookDiff
}

// HasDifferences returns true if there are any differences.
func (d *BackupDiff) HasDifferences() bool {
	return len(d.ConfigDiffs) > 0 || len(d.ScriptDiffs) > 0 ||
		len(d.ScheduleDiffs) > 0 || len(d.WebhookDiffs) > 0
}

// ConfigDiff represents a configuration difference.
type ConfigDiff struct {
	Key      string
	Current  any
	Backup   any
	DiffType string // "added", "removed", "changed"
}

// ScriptDiff represents a script difference.
type ScriptDiff struct {
	Name     string
	DiffType string // "added", "removed", "changed"
	Details  string
}

// ScheduleDiff represents a schedule difference.
type ScheduleDiff struct {
	Timespec string
	DiffType string // "added", "removed", "changed"
	Details  string
}

// WebhookDiff represents a webhook difference.
type WebhookDiff struct {
	Event    string
	Name     string
	DiffType string // "added", "removed", "changed"
	Details  string
}

// Helper functions for converting backup data structures

func convertBackupScripts(scripts []*backup.Script) []BackupScript {
	result := make([]BackupScript, len(scripts))
	for i, s := range scripts {
		result[i] = BackupScript{
			ID:     s.ID,
			Name:   s.Name,
			Enable: s.Enable,
			Code:   s.Code,
		}
	}
	return result
}

// BackupScript is a compatibility type.
type BackupScript struct {
	ID     int
	Name   string
	Enable bool
	Code   string
}

func convertBackupSchedules(data json.RawMessage) []BackupSchedule {
	var schedData struct {
		Jobs []struct {
			Enable   bool           `json:"enable"`
			Timespec string         `json:"timespec"`
			Calls    []ScheduleCall `json:"calls"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(data, &schedData); err != nil {
		return nil
	}

	result := make([]BackupSchedule, len(schedData.Jobs))
	for i, j := range schedData.Jobs {
		result[i] = BackupSchedule{
			Enable:   j.Enable,
			Timespec: j.Timespec,
			Calls:    j.Calls,
		}
	}
	return result
}

// BackupSchedule is a compatibility type.
type BackupSchedule struct {
	Enable   bool
	Timespec string
	Calls    []ScheduleCall
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

func compareConfigs(current, bkup map[string]any) []ConfigDiff {
	var diffs []ConfigDiff

	// Check for keys in backup that differ from current
	for key, backupVal := range bkup {
		currentVal, exists := current[key]
		if !exists {
			diffs = append(diffs, ConfigDiff{
				Key:      key,
				Backup:   backupVal,
				DiffType: "added",
			})
		} else if !deepEqualJSON(currentVal, backupVal) {
			diffs = append(diffs, ConfigDiff{
				Key:      key,
				Current:  currentVal,
				Backup:   backupVal,
				DiffType: "changed",
			})
		}
	}

	// Check for keys in current that are not in backup
	for key, currentVal := range current {
		if _, exists := bkup[key]; !exists {
			diffs = append(diffs, ConfigDiff{
				Key:      key,
				Current:  currentVal,
				DiffType: "removed",
			})
		}
	}

	return diffs
}

func compareScripts(current []ScriptInfo, bkup []BackupScript) []ScriptDiff {
	var diffs []ScriptDiff

	currentMap := make(map[string]ScriptInfo)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	backupMap := make(map[string]BackupScript)
	for _, s := range bkup {
		backupMap[s.Name] = s
	}

	// Check backup scripts
	for name, backupScript := range backupMap {
		if _, exists := currentMap[name]; !exists {
			diffs = append(diffs, ScriptDiff{
				Name:     name,
				DiffType: "added",
				Details:  "script will be created",
			})
		} else {
			diffs = append(diffs, ScriptDiff{
				Name:     name,
				DiffType: "changed",
				Details:  fmt.Sprintf("enable: %v", backupScript.Enable),
			})
		}
	}

	// Check for scripts not in backup
	for name := range currentMap {
		if _, exists := backupMap[name]; !exists {
			diffs = append(diffs, ScriptDiff{
				Name:     name,
				DiffType: "removed",
				Details:  "script not in backup (will not be deleted)",
			})
		}
	}

	return diffs
}

func compareSchedules(current []ScheduleJob, bkup []BackupSchedule) []ScheduleDiff {
	var diffs []ScheduleDiff

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
			diffs = append(diffs, ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: "added",
				Details:  fmt.Sprintf("enable: %v", s.Enable),
			})
		}
	}

	for _, s := range current {
		if !backupTimespecs[s.Timespec] {
			diffs = append(diffs, ScheduleDiff{
				Timespec: s.Timespec,
				DiffType: "removed",
				Details:  "schedule not in backup",
			})
		}
	}

	return diffs
}

func compareWebhooks(current, bkup []WebhookInfo) []WebhookDiff {
	var diffs []WebhookDiff

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
			diffs = append(diffs, WebhookDiff{
				Event:    backupWh.Event,
				Name:     backupWh.Name,
				DiffType: "added",
				Details:  fmt.Sprintf("urls: %v", backupWh.URLs),
			})
		}
	}

	for key, currentWh := range currentMap {
		if _, exists := backupMap[key]; !exists {
			diffs = append(diffs, WebhookDiff{
				Event:    currentWh.Event,
				Name:     currentWh.Name,
				DiffType: "removed",
				Details:  "webhook not in backup",
			})
		}
	}

	return diffs
}

// deepEqualJSON compares two values for equality using JSON serialization.
func deepEqualJSON(a, b any) bool {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aJSON, bJSON)
}

// SaveBackupToFile saves backup data to a file.
func (s *Service) SaveBackupToFile(data []byte, filePath string) error {
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

// LoadBackupFromFile loads backup data from a file.
func (s *Service) LoadBackupFromFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec // User-provided file path
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}
	return data, nil
}

// GenerateBackupFilename generates a backup filename based on device info and timestamp.
func (s *Service) GenerateBackupFilename(deviceName, deviceID string, encrypted bool) string {
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
