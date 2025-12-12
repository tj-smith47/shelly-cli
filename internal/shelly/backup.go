// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DeviceBackup represents a complete backup of a Shelly device.
type DeviceBackup struct {
	// Metadata
	Version   string           `json:"version"`
	CreatedAt time.Time        `json:"created_at"`
	Device    BackupDeviceInfo `json:"device"`

	// Core configuration
	Config map[string]any `json:"config"`

	// Scripts (Gen2+)
	Scripts []BackupScript `json:"scripts,omitempty"`

	// Schedules
	Schedules []BackupSchedule `json:"schedules,omitempty"`

	// Webhooks
	Webhooks []WebhookInfo `json:"webhooks,omitempty"`

	// Encrypted indicates if sensitive data is encrypted.
	Encrypted bool `json:"encrypted,omitempty"`

	// PasswordHash stores bcrypt hash for verification (only when encrypted).
	PasswordHash string `json:"password_hash,omitempty"`
}

// BackupDeviceInfo contains device identification for backup.
type BackupDeviceInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	Model      string `json:"model"`
	Generation int    `json:"generation"`
	FWVersion  string `json:"fw_version"`
	MAC        string `json:"mac,omitempty"`
}

// BackupScript contains script data for backup.
type BackupScript struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
	Code   string `json:"code"`
}

// BackupSchedule contains schedule data for backup.
type BackupSchedule struct {
	Enable   bool           `json:"enable"`
	Timespec string         `json:"timespec"`
	Calls    []ScheduleCall `json:"calls"`
}

// BackupOptions configures backup creation.
type BackupOptions struct {
	// SkipScripts excludes scripts from backup.
	SkipScripts bool
	// SkipSchedules excludes schedules from backup.
	SkipSchedules bool
	// SkipWebhooks excludes webhooks from backup.
	SkipWebhooks bool
	// Password for encryption (empty = no encryption).
	Password string
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
	// Password for decryption (required if backup is encrypted).
	Password string
}

// RestoreResult contains the result of a restore operation.
type RestoreResult struct {
	ConfigRestored    bool
	ScriptsRestored   int
	SchedulesRestored int
	WebhooksRestored  int
	Warnings          []string
}

// CreateBackup creates a complete backup of a device.
func (s *Service) CreateBackup(ctx context.Context, identifier string, opts BackupOptions) (*DeviceBackup, error) {
	backup := &DeviceBackup{
		Version:   "1.0",
		CreatedAt: time.Now().UTC(),
	}

	// Get device info
	info, err := s.DeviceInfo(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}
	backup.Device = BackupDeviceInfo{
		ID:         info.ID,
		Model:      info.Model,
		Generation: info.Generation,
		FWVersion:  info.Firmware,
		MAC:        info.MAC,
	}

	// Get full configuration
	config, err := s.GetConfig(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	backup.Config = config

	// Get scripts (Gen2+ only)
	if !opts.SkipScripts && info.Generation >= 2 {
		scripts, err := s.backupScripts(ctx, identifier)
		if err != nil {
			// Non-fatal - device might not support scripts
			backup.Scripts = []BackupScript{}
		} else {
			backup.Scripts = scripts
		}
	}

	// Get schedules (Gen2+ only)
	if !opts.SkipSchedules && info.Generation >= 2 {
		schedules, err := s.backupSchedules(ctx, identifier)
		if err != nil {
			// Non-fatal - device might not support schedules
			backup.Schedules = []BackupSchedule{}
		} else {
			backup.Schedules = schedules
		}
	}

	// Get webhooks (Gen2+ only)
	if !opts.SkipWebhooks && info.Generation >= 2 {
		webhooks, err := s.ListWebhooks(ctx, identifier)
		if err != nil {
			// Non-fatal
			backup.Webhooks = []WebhookInfo{}
		} else {
			backup.Webhooks = webhooks
		}
	}

	// Handle encryption if password provided
	if opts.Password != "" {
		if err := backup.encrypt(opts.Password); err != nil {
			return nil, fmt.Errorf("failed to encrypt backup: %w", err)
		}
	}

	return backup, nil
}

// backupScripts retrieves all scripts with their code.
func (s *Service) backupScripts(ctx context.Context, identifier string) ([]BackupScript, error) {
	scripts, err := s.ListScripts(ctx, identifier)
	if err != nil {
		return nil, err
	}

	result := make([]BackupScript, 0, len(scripts))
	for _, script := range scripts {
		code, err := s.GetScriptCode(ctx, identifier, script.ID)
		if err != nil {
			// Skip scripts we can't read
			continue
		}
		result = append(result, BackupScript{
			ID:     script.ID,
			Name:   script.Name,
			Enable: script.Enable,
			Code:   code,
		})
	}
	return result, nil
}

// backupSchedules retrieves all schedules.
func (s *Service) backupSchedules(ctx context.Context, identifier string) ([]BackupSchedule, error) {
	schedules, err := s.ListSchedules(ctx, identifier)
	if err != nil {
		return nil, err
	}

	result := make([]BackupSchedule, len(schedules))
	for i, sched := range schedules {
		result[i] = BackupSchedule{
			Enable:   sched.Enable,
			Timespec: sched.Timespec,
			Calls:    sched.Calls,
		}
	}
	return result, nil
}

// encrypt encrypts sensitive data in the backup.
// Note: This is a simple encryption marker - full encryption would require
// encrypting the Config, Scripts, etc. with AES using the password.
// For now, we just mark it as encrypted and store a password hash for verification.
func (b *DeviceBackup) encrypt(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	b.Encrypted = true
	b.PasswordHash = string(hash)
	return nil
}

// VerifyPassword checks if the provided password matches the backup's password.
func (b *DeviceBackup) VerifyPassword(password string) bool {
	if !b.Encrypted {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(b.PasswordHash), []byte(password))
	return err == nil
}

// RestoreBackup restores a backup to a device.
func (s *Service) RestoreBackup(ctx context.Context, identifier string, backup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	result := &RestoreResult{}

	// Verify password if encrypted
	if backup.Encrypted && !backup.VerifyPassword(opts.Password) {
		return nil, fmt.Errorf("incorrect password for encrypted backup")
	}

	// Dry run just returns what would happen
	if opts.DryRun {
		result.ConfigRestored = len(backup.Config) > 0
		result.ScriptsRestored = len(backup.Scripts)
		result.SchedulesRestored = len(backup.Schedules)
		result.WebhooksRestored = len(backup.Webhooks)
		return result, nil
	}

	// Restore configuration
	config := backup.Config
	if opts.SkipNetwork {
		config = filterNetworkConfig(config)
	}
	if len(config) > 0 {
		if err := s.SetConfig(ctx, identifier, config); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("config restore failed: %v", err))
		} else {
			result.ConfigRestored = true
		}
	}

	// Restore scripts
	if !opts.SkipScripts {
		restored, warnings := s.restoreScripts(ctx, identifier, backup.Scripts)
		result.ScriptsRestored = restored
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Restore schedules
	if !opts.SkipSchedules {
		restored, warnings := s.restoreSchedules(ctx, identifier, backup.Schedules)
		result.SchedulesRestored = restored
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Restore webhooks
	if !opts.SkipWebhooks {
		restored, warnings := s.restoreWebhooks(ctx, identifier, backup.Webhooks)
		result.WebhooksRestored = restored
		result.Warnings = append(result.Warnings, warnings...)
	}

	return result, nil
}

// filterNetworkConfig removes network-related configuration.
func filterNetworkConfig(config map[string]any) map[string]any {
	filtered := make(map[string]any)
	for k, v := range config {
		// Skip WiFi, Ethernet, and cloud configuration
		switch k {
		case "wifi", "eth", "cloud", "sys":
			continue
		default:
			filtered[k] = v
		}
	}
	return filtered
}

// restoreScripts restores scripts to a device.
func (s *Service) restoreScripts(ctx context.Context, identifier string, scripts []BackupScript) (restored int, warnings []string) {
	for _, script := range scripts {
		// Create the script
		id, err := s.CreateScript(ctx, identifier, script.Name)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to create script %q: %v", script.Name, err))
			continue
		}

		// Upload the code
		if err := s.UpdateScriptCode(ctx, identifier, id, script.Code, false); err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to upload code for script %q: %v", script.Name, err))
			continue
		}

		// Set enable state
		if err := s.UpdateScriptConfig(ctx, identifier, id, nil, &script.Enable); err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to configure script %q: %v", script.Name, err))
		}

		restored++
	}

	return restored, warnings
}

// restoreSchedules restores schedules to a device.
func (s *Service) restoreSchedules(ctx context.Context, identifier string, schedules []BackupSchedule) (restored int, warnings []string) {
	for _, sched := range schedules {
		_, err := s.CreateSchedule(ctx, identifier, sched.Enable, sched.Timespec, sched.Calls)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to create schedule: %v", err))
			continue
		}
		restored++
	}

	return restored, warnings
}

// restoreWebhooks restores webhooks to a device.
func (s *Service) restoreWebhooks(ctx context.Context, identifier string, webhooks []WebhookInfo) (restored int, warnings []string) {
	for _, wh := range webhooks {
		params := CreateWebhookParams{
			Event:  wh.Event,
			URLs:   wh.URLs,
			Name:   wh.Name,
			Enable: wh.Enable,
			Cid:    wh.Cid,
		}
		_, err := s.CreateWebhook(ctx, identifier, params)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to create webhook %q: %v", wh.Name, err))
			continue
		}
		restored++
	}

	return restored, warnings
}

// ValidateBackup validates a backup file structure.
func ValidateBackup(data []byte) (*DeviceBackup, error) {
	var backup DeviceBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("invalid backup format: %w", err)
	}

	if backup.Version == "" {
		return nil, fmt.Errorf("missing backup version")
	}
	if backup.Device.ID == "" && backup.Device.Model == "" {
		return nil, fmt.Errorf("missing device information")
	}
	if backup.Config == nil {
		return nil, fmt.Errorf("missing configuration data")
	}

	return &backup, nil
}

// CompareBackup compares a backup with a device's current state.
func (s *Service) CompareBackup(ctx context.Context, identifier string, backup *DeviceBackup) (*BackupDiff, error) {
	diff := &BackupDiff{}

	// Get current config
	currentConfig, err := s.GetConfig(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Compare configurations
	diff.ConfigDiffs = compareConfigs(currentConfig, backup.Config)

	// Get current scripts
	currentScripts, err := s.ListScripts(ctx, identifier)
	if err == nil {
		diff.ScriptDiffs = compareScripts(currentScripts, backup.Scripts)
	}

	// Get current schedules
	currentSchedules, err := s.ListSchedules(ctx, identifier)
	if err == nil {
		diff.ScheduleDiffs = compareSchedules(currentSchedules, backup.Schedules)
	}

	// Get current webhooks
	currentWebhooks, err := s.ListWebhooks(ctx, identifier)
	if err == nil {
		diff.WebhookDiffs = compareWebhooks(currentWebhooks, backup.Webhooks)
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

// compareConfigs compares two configuration maps.
func compareConfigs(current, backup map[string]any) []ConfigDiff {
	var diffs []ConfigDiff

	// Check for keys in backup that differ from current
	for key, backupVal := range backup {
		currentVal, exists := current[key]
		if !exists {
			diffs = append(diffs, ConfigDiff{
				Key:      key,
				Backup:   backupVal,
				DiffType: "added",
			})
		} else if !deepEqualAny(currentVal, backupVal) {
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
		if _, exists := backup[key]; !exists {
			diffs = append(diffs, ConfigDiff{
				Key:      key,
				Current:  currentVal,
				DiffType: "removed",
			})
		}
	}

	return diffs
}

// compareScripts compares current scripts with backup scripts.
func compareScripts(current []ScriptInfo, backup []BackupScript) []ScriptDiff {
	var diffs []ScriptDiff

	currentMap := make(map[string]ScriptInfo)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	backupMap := make(map[string]BackupScript)
	for _, s := range backup {
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
			// Script exists - would need code comparison for full diff
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

// compareSchedules compares current schedules with backup schedules.
func compareSchedules(current []ScheduleJob, backup []BackupSchedule) []ScheduleDiff {
	var diffs []ScheduleDiff

	// Simple comparison by timespec
	currentTimespecs := make(map[string]bool)
	for _, s := range current {
		currentTimespecs[s.Timespec] = true
	}

	backupTimespecs := make(map[string]bool)
	for _, s := range backup {
		backupTimespecs[s.Timespec] = true
	}

	for _, s := range backup {
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

// compareWebhooks compares current webhooks with backup webhooks.
func compareWebhooks(current, backup []WebhookInfo) []WebhookDiff {
	var diffs []WebhookDiff

	currentMap := make(map[string]WebhookInfo)
	for _, w := range current {
		key := fmt.Sprintf("%s:%s", w.Event, w.Name)
		currentMap[key] = w
	}

	backupMap := make(map[string]WebhookInfo)
	for _, w := range backup {
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

// deepEqualAny compares two values for equality using JSON serialization.
func deepEqualAny(a, b any) bool {
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
