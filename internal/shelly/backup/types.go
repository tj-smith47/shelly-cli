// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/reprovision"
)

// ipv4ModeStatic is the value used to request static IPv4 addressing on both
// Gen1 (ipv4_method) and Gen2 (ipv4mode) WiFi station config.
const ipv4ModeStatic = "static"

// DeviceBackup wraps the shelly-go backup.Backup with CLI-specific methods.
type DeviceBackup struct {
	*shellybackup.Backup
	// encrypted records whether this backup was loaded from, or written as, an
	// AES-encrypted envelope. The in-memory Backup is always plaintext; this flag
	// tracks the on-disk protection state for summaries and listings.
	encrypted bool
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

// Encrypted reports whether the backup was loaded from, or written as, an
// AES-encrypted envelope. See Load and Encrypt for how the flag is set.
func (b *DeviceBackup) Encrypted() bool {
	return b.encrypted
}

// ConfigKeyCount returns the number of top-level keys in the backed-up device
// configuration. Config holds the device's raw JSON, so its length is a byte
// count and says nothing about how much configuration it carries.
func (b *DeviceBackup) ConfigKeyCount() int {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(b.Config, &keys); err != nil {
		return 0
	}
	return len(keys)
}

// WebhookCount returns the number of webhooks in the backup: Gen1 action
// entries, or Gen2 hooks. Webhooks holds the device's raw JSON, so its length
// is a byte count.
func (b *DeviceBackup) WebhookCount() int {
	if b.Device().Generation == 1 {
		return countGen1Actions(b.Backup)
	}
	return countListed(b.Webhooks, "hooks")
}

// ScheduleCount returns the number of schedule jobs in the backup.
func (b *DeviceBackup) ScheduleCount() int {
	return countListed(b.Schedules, "jobs")
}

// countListed returns the length of the list stored under field in a raw
// device reply, or zero when the reply is absent or has another shape.
func countListed(raw json.RawMessage, field string) int {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return 0
	}
	var items []json.RawMessage
	if err := json.Unmarshal(doc[field], &items); err != nil {
		return 0
	}
	return len(items)
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
// SSID and Password are optional. A restore writes a station key only when one
// is known: Password, or the key a Gen1 backup holds. With neither, no key is
// written and the device keeps the one it has.
type NetworkOverride struct {
	// SSID overrides the station SSID; empty keeps the backup's SSID.
	SSID string
	// Password is written as the station key. Empty writes the key a Gen1
	// backup holds when SSID leaves the backup's network alone, and otherwise
	// no key at all, so the device keeps its own.
	Password string
	// StaticIP, when set, switches the station to a static IPv4 address. An
	// empty Gateway, Netmask or DNS falls back to the backup's static settings.
	StaticIP string
	// Gateway is the static IPv4 default gateway.
	Gateway string
	// Netmask is the static IPv4 subnet mask.
	Netmask string
	// DNS is the static IPv4 nameserver (optional).
	DNS string
	// Open joins the network with no passphrase; Password must be empty.
	Open bool
}

// Network returns the override as the SDK's reprovision.Network; a nil
// override is an empty one.
func (n *NetworkOverride) Network() reprovision.Network {
	if n == nil {
		return reprovision.Network{}
	}
	return reprovision.Network{
		SSID: n.SSID, Password: n.Password, StaticIP: n.StaticIP,
		Gateway: n.Gateway, Netmask: n.Netmask, DNS: n.DNS, Open: n.Open,
	}
}

// IsStatic reports whether a static IPv4 address was requested.
func (n *NetworkOverride) IsStatic() bool {
	return n != nil && n.StaticIP != ""
}

// StaticNetwork returns the static addressing a restore of the backup with
// override writes: the override's address with any missing gateway, netmask or
// DNS taken from the backup's own static settings, or the backup's addressing
// when the override sets no address. IP is empty for a DHCP station. A static
// address left without a gateway or netmask is refused with the
// StaticNetworkError wording, which matches backup.ErrIncompleteStaticNetwork.
func (d *DeviceBackup) StaticNetwork(override *NetworkOverride) (shellybackup.StaticNetwork, error) {
	static, err := resolveStatic(override, d.Backup)
	return static, StaticNetworkError(err, override)
}

// resolveStatic applies shelly-go's static address rule to an override and the
// station network bkp records.
func resolveStatic(override *NetworkOverride, bkp *shellybackup.Backup) (shellybackup.StaticNetwork, error) {
	n := reprovision.NetworkFromBackup(bkp)
	fromBackup := shellybackup.StaticNetwork{IP: n.StaticIP, Gateway: n.Gateway, Netmask: n.Netmask, DNS: n.DNS}
	var ov shellybackup.StaticNetwork
	if override != nil {
		ov = shellybackup.StaticNetwork{
			IP: override.StaticIP, Gateway: override.Gateway, Netmask: override.Netmask, DNS: override.DNS,
		}
	}
	return shellybackup.ResolveStaticNetwork(ov, fromBackup)
}

// MsgIncompleteBackupStatic reports a backup whose static address has no
// gateway or netmask, when no --static-ip replaces it.
const MsgIncompleteBackupStatic = "the backup's static address has no gateway or netmask — pass --static-ip " +
	"with --gateway and --netmask"

// CLIError is an error from shelly-go reworded for the CLI: flag names where the
// SDK names its options, and the target's name where the SDK says "the device".
// errors.Is and errors.As still reach the SDK error.
type CLIError struct {
	Msg string
	Err error
}

// Error returns the CLI wording.
func (e *CLIError) Error() string { return e.Msg }

// Unwrap returns the SDK error.
func (e *CLIError) Unwrap() error { return e.Err }

// StaticNetworkError words a static address that shelly-go refused for lacking a
// gateway or netmask (backup.ErrIncompleteStaticNetwork) in terms of the CLI's
// flags, given the network override the restore used. Any other error, nil
// included, is returned unchanged.
func StaticNetworkError(err error, override *NetworkOverride) error {
	if !errors.Is(err, shellybackup.ErrIncompleteStaticNetwork) {
		return err
	}
	if override.IsStatic() {
		return &CLIError{
			Msg: fmt.Sprintf("static address %s needs a gateway and a netmask, and the backup has none — pass "+
				"--gateway and --netmask", override.StaticIP),
			Err: err,
		}
	}
	return &CLIError{Msg: MsgIncompleteBackupStatic, Err: err}
}

// RestoreOptions configures backup restoration.
type RestoreOptions struct {
	// DryRun shows what would be changed without applying.
	DryRun bool
	// Name is an explicit device name (--name). It is always written.
	Name string
	// AliasName is the target's registry alias, applied as the device name
	// only when the backup's MAC differs from the target's (see ResolveName).
	AliasName string
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
	// Station1 says how the backup's secondary WiFi station is written. The
	// zero value writes it as the backup records it, without a key on Gen2+.
	Station1 Station1Write
	// StepTrace, when non-nil, receives a per-step diagnostic line during a Gen1
	// restore (each setting group's warnings/errors and the device's post-write
	// uptime/stability). It is the debug seam behind --trace-file for pinpointing
	// which setting destabilizes a fragile device; nil on a normal restore.
	StepTrace io.Writer
}

// Station1Write is how a restore writes the backup's secondary WiFi station
// (sta1).
type Station1Write struct {
	// Omit leaves the secondary station out of the write; the restore result
	// carries a warning naming SSID.
	Omit bool
	// SSID is the secondary station's network, for the Omit warning.
	SSID string
	// Unread says the Omit is because the device's stations could not be
	// read, for the warning's wording.
	Unread bool
	// Password is written as its key.
	Password string
	// Open writes it as an open network.
	Open bool
}

// omitWarning is the restore warning for a secondary station left out.
func (w Station1Write) omitWarning() string {
	if w.Unread {
		return fmt.Sprintf("secondary WiFi station %q was not restored: the device's WiFi stations could not be "+
			"read to compare networks; set it with `shelly wifi set`", w.SSID)
	}
	return fmt.Sprintf("secondary WiFi station %q was not restored: the device has a different network there "+
		"and no password for %q was found on this host; set it with `shelly wifi set`", w.SSID, w.SSID)
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

// Failed reports whether the device rejected any part of the restore. A restore
// can fail per-section (WiFi/Cloud/MQTT/auth/…) while the call returns a nil
// top-level error, so callers must consult this rather than err alone.
func (r *RestoreResult) Failed() bool {
	return r != nil && !r.Success
}

// Err summarizes why the restore did not fully succeed, or nil when it did.
// Callers that surface a restore as a plain error (provisioning, the TUI import
// flow) use this directly; the command layer wraps it with the target name.
func (r *RestoreResult) Err() error {
	if !r.Failed() {
		return nil
	}
	if r.DestabilizedStep != "" {
		return fmt.Errorf("the device entered a reboot loop after the %q step", r.DestabilizedStep)
	}
	if n := len(r.Errors); n > 0 {
		return fmt.Errorf("%d section(s) rejected: %s", n, strings.Join(r.Errors, "; "))
	}
	return errors.New("restore did not complete")
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
