// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"io"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
)

// ipv4ModeStatic is the value used to request static IPv4 addressing on both
// Gen1 (ipv4_method) and Gen2 (ipv4mode) WiFi station config.
const ipv4ModeStatic = "static"

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

// NetworkOverride replaces the WiFi station settings applied during restore.
// It lets you clone one device's configuration onto another without copying the
// source's IP address, so both devices stay online with distinct addresses.
//
// Identity fields (MAC, device ID, serial) are never written by restore on any
// generation, so only network settings ever need overriding for a safe clone.
// SSID and Password are optional: when empty, the backup's own credentials are
// kept (a Gen1 backup includes the station key; a Gen2 backup omits it).
type NetworkOverride struct {
	// SSID overrides the station SSID; empty keeps the backup's SSID.
	SSID string
	// Password overrides the station key; empty keeps the backup's key.
	Password string
	// StaticIP, when set, switches the station to a static IPv4 address.
	// Gateway and Netmask are required alongside it.
	StaticIP string
	// Gateway is the static IPv4 default gateway.
	Gateway string
	// Netmask is the static IPv4 subnet mask.
	Netmask string
	// DNS is the static IPv4 nameserver (optional).
	DNS string
}

// IsStatic reports whether a static IPv4 address was requested.
func (n *NetworkOverride) IsStatic() bool {
	return n != nil && n.StaticIP != ""
}

// RestoreOptions configures backup restoration.
type RestoreOptions struct {
	// DryRun shows what would be changed without applying.
	DryRun bool
	// Name overrides the device's stored display name. Empty leaves the name as
	// the backup's. Used so a cloned device is named distinctly from its source.
	Name string
	// SkipNetwork skips WiFi/Ethernet configuration.
	SkipNetwork bool
	// NetworkOverride, when non-nil, replaces the backup's WiFi station settings
	// before they are applied. Ignored when SkipNetwork is true.
	NetworkOverride *NetworkOverride
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
	// SkipState skips restoring captured live component state — color temperature
	// and brightness — so a restore leaves the target's current light look intact
	// and applies configuration only.
	SkipState bool
	// SkipMeters skips restoring meter / energy-meter configuration (e.g. Gen1
	// overpower limits), leaving the target's protection settings untouched.
	SkipMeters bool
	// Password for decryption (required if backup is encrypted).
	Password string
	// ClockDependentOnly restores only the clock-gated configuration (Gen1 light
	// component config and captured light state). Set for the LAN second pass of a
	// --to-ap restore, where everything else already applied at the factory AP and
	// only the time-dependent writes the clockless AP rejected need re-applying.
	ClockDependentOnly bool
	// AllowFirmwareDowngrade forces the older-firmware config write instead of the
	// automatic firmware update. By default, when the backup is from newer firmware
	// than the target runs, the device is OTA-updated to matched firmware first (the
	// safe resolution of a downgrade — a proven reboot-loop trigger otherwise). Set
	// this only to skip that update and force the downgrade, accepting the risk.
	AllowFirmwareDowngrade bool
	// FirmwareURL overrides the firmware image the automatic downgrade-recovery update
	// flashes. Empty derives the official current-stable URL from the device's model.
	FirmwareURL string
	// NetworkOnly writes only the WiFi station configuration and returns, bypassing
	// every other step and the firmware-downgrade gate. It is the factory-AP pass of
	// a --to-ap restore: after any needed firmware update is flashed at the AP, only
	// the station config is written there so the device joins the LAN cleanly, and the
	// full configuration is then applied on the LAN — where the device has a clock and
	// is stable, and where writes cannot be misread as a reboot loop when the device
	// reboots to join the network.
	NetworkOnly bool
	// SkipClockWait disables the bounded wait for the device's NTP clock that
	// otherwise precedes writing time-based schedule rules. Set it for the full-config
	// pass run at a clockless factory AP, where the device can never sync time and the
	// LAN second pass re-applies those rules once it has joined the network.
	SkipClockWait bool
	// StepTrace, when non-nil, receives a per-step diagnostic line during a Gen1
	// restore (each setting group's warnings/errors and the device's post-write
	// uptime/stability). It is the debug seam behind --trace-file for pinpointing
	// which setting destabilizes a fragile device; nil on a normal restore.
	StepTrace io.Writer
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
	// DestabilizedStep names the restore step after which the device failed to
	// restabilize — a write drove it into a reboot loop and the restore halted.
	// Empty on a clean restore.
	DestabilizedStep  string
	Warnings          []string
	Errors            []string
	Success           bool
	ConfigRestored    bool
	ScriptsRestored   int
	SchedulesRestored int
	WebhooksRestored  int
	RestartRequired   bool
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
