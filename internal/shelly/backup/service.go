// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"context"
	"encoding/json"
	"fmt"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/gen2/components"

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
		bkp, err := shellybackup.ExportGen1(ctx, conn.Device())
		if err != nil {
			return err
		}

		// Enrich DeviceInfo from the connector's probe, which carries the full
		// identity (ID, App) that the minimal /shelly-derived DeviceInfo omits.
		info := conn.Info()
		bkp.DeviceInfo = &shellybackup.DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: info.Generation,
			Version:    info.Firmware,
			App:        info.App,
		}

		result = &DeviceBackup{Backup: bkp}
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

// RestoreBackupGen restores a backup using an explicitly supplied generation,
// skipping the device-probing that RestoreBackup performs. It is used when the
// target's generation is already known from the backup — e.g. a device sitting
// at its factory WiFi AP, where generation auto-detection is unreliable (a Gen1
// device's bare AP IP carries no generation hint and can be misrouted to the
// Gen2 RPC path). generation < 2 selects the Gen1 REST path; otherwise Gen2 RPC.
func (s *Service) RestoreBackupGen(ctx context.Context, identifier string, generation int, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	if generation == 1 {
		return s.restoreGen1Backup(ctx, identifier, deviceBackup, opts)
	}
	return s.restoreGen2Backup(ctx, identifier, deviceBackup, opts)
}

// restoreGen1Backup restores a backup to a Gen1 device via HTTP REST.
func (s *Service) restoreGen1Backup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	var result *RestoreResult

	err := s.connector.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		r, err := shellybackup.RestoreGen1(ctx, conn.Device(), deviceBackup.Backup, toGen1RestoreOptions(opts))
		if err != nil {
			return err
		}
		result = &RestoreResult{
			Success:          r.Success,
			ConfigRestored:   r.Success,
			RestartRequired:  r.RestartRequired,
			Warnings:         r.Warnings,
			Errors:           gen1ErrorStrings(r.Errors),
			DestabilizedStep: r.DestabilizedStep,
		}
		// The library result carries no per-section counts; recover the webhook
		// count the same way the old in-CLI restore did (number of action entries
		// in the backup), gated on the webhook restore actually running and the
		// restore reaching the webhook step (a halted restore never does).
		if !opts.SkipWebhooks && r.Success {
			result.WebhooksRestored = countGen1Actions(deviceBackup.Backup)
		}
		// A destabilizing step is a hard failure: the restore halted with the device
		// in a reboot loop. Surface it as an error so the command never reports a
		// false success, naming the breaking step for diagnosis.
		if r.DestabilizedStep != "" {
			return fmt.Errorf(
				"restore halted: device became unstable after the %q step — a write drove it "+
					"into a reboot loop; capture the per-step trace with --trace-file to confirm",
				r.DestabilizedStep)
		}
		return nil
	})

	return result, err
}

// gen1ErrorStrings renders the library restore errors as display strings for the
// CLI result, so per-setting failures surface to the user instead of being
// dropped. Returns nil for an empty slice (a clean restore).
func gen1ErrorStrings(errs []error) []string {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, len(errs))
	for i, err := range errs {
		out[i] = err.Error()
	}
	return out
}

// toGen1RestoreOptions translates the CLI RestoreOptions into the shelly-go
// Gen1RestoreOptions consumed by shellybackup.RestoreGen1.
func toGen1RestoreOptions(opts RestoreOptions) *shellybackup.Gen1RestoreOptions {
	out := &shellybackup.Gen1RestoreOptions{
		Name:                   opts.Name,
		SkipNetwork:            opts.SkipNetwork,
		SkipAuth:               opts.SkipAuth,
		SkipState:              opts.SkipState,
		SkipMeters:             opts.SkipMeters,
		SkipWebhooks:           opts.SkipWebhooks,
		ClockDependentOnly:     opts.ClockDependentOnly,
		AllowFirmwareDowngrade: opts.AllowFirmwareDowngrade,
		UpdateFirmware:         opts.UpdateFirmware,
		FirmwareURL:            opts.FirmwareURL,
		NetworkOnly:            opts.NetworkOnly,
		StepTrace:              opts.StepTrace,
	}
	if opts.NetworkOverride != nil {
		out.NetworkOverride = &shellybackup.Gen1NetworkOverride{
			SSID:     opts.NetworkOverride.SSID,
			Password: opts.NetworkOverride.Password,
			StaticIP: opts.NetworkOverride.StaticIP,
			Gateway:  opts.NetworkOverride.Gateway,
			Netmask:  opts.NetworkOverride.Netmask,
			DNS:      opts.NetworkOverride.DNS,
		}
	}
	return out
}

// countGen1Actions reports the number of action-URL entries stored in a Gen1
// backup's webhook blob, matching the WebhooksRestored value the previous
// in-CLI restore reported. A nil or unparseable blob counts as zero.
func countGen1Actions(bkp *shellybackup.Backup) int {
	if bkp.Webhooks == nil {
		return 0
	}
	var actions gen1.ActionSettings
	if err := json.Unmarshal(bkp.Webhooks, &actions); err != nil {
		return 0
	}
	return len(actions.Actions)
}

// restoreGen2Backup restores a backup to a Gen2+ device via RPC.
func (s *Service) restoreGen2Backup(ctx context.Context, identifier string, deviceBackup *DeviceBackup, opts RestoreOptions) (*RestoreResult, error) {
	var result *RestoreResult

	err := s.connector.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := shellybackup.New(conn.RPCClient())

		// Apply a network override (if any) onto a shallow copy so the caller's
		// backup is left untouched; the rewritten WiFi blob is what gets restored.
		toRestore := deviceBackup.Backup
		if opts.NetworkOverride != nil && !opts.SkipNetwork {
			wifi, overrideErr := applyGen2WiFiOverride(toRestore.WiFi, opts.NetworkOverride)
			if overrideErr != nil {
				return overrideErr
			}
			clone := *toRestore
			clone.WiFi = wifi
			toRestore = &clone
		}

		// Serialize the backup
		data, err := json.Marshal(toRestore)
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

		// The shelly-go restore does not apply Sys config, so a name override
		// must be set explicitly. Failures are non-fatal warnings.
		if opts.Name != "" && !opts.DryRun {
			name := opts.Name
			sys := components.NewSys(conn.RPCClient())
			if nameErr := sys.SetConfig(ctx, &components.SysConfig{
				Device: &components.SysDeviceConfig{Name: &name},
			}); nameErr != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("set device name: %v", nameErr))
			}
		}

		return nil
	})

	return result, err
}

// applyGen2WiFiOverride overlays a NetworkOverride onto a Gen2 WiFi config blob
// (the raw WiFi.GetConfig result) and returns the rewritten blob. SSID and pass
// are replaced only when explicitly provided; a static IP switches the station
// to static IPv4 addressing. The blob shape ({ap, sta, sta1}) is preserved so it
// round-trips through WiFi.SetConfig unchanged apart from the override.
func applyGen2WiFiOverride(wifiBlob json.RawMessage, ov *NetworkOverride) (json.RawMessage, error) {
	cfg := map[string]any{}
	if len(wifiBlob) > 0 {
		if err := json.Unmarshal(wifiBlob, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse WiFi config for override: %w", err)
		}
	}

	sta, ok := cfg["sta"].(map[string]any)
	if !ok || sta == nil {
		sta = map[string]any{}
	}
	sta["enable"] = true
	if ov.SSID != "" {
		sta["ssid"] = ov.SSID
	}
	if ov.Password != "" {
		sta["pass"] = ov.Password
	}
	if ov.IsStatic() {
		sta["ipv4mode"] = ipv4ModeStatic
		sta["ip"] = ov.StaticIP
		sta["netmask"] = ov.Netmask
		sta["gw"] = ov.Gateway
		if ov.DNS != "" {
			sta["nameserver"] = ov.DNS
		}
	}
	cfg["sta"] = sta

	return json.Marshal(cfg)
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
	resolved := ResolveFilePath(source)
	if IsFile(resolved) {
		bkp, err = LoadAndValidate(resolved)
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
