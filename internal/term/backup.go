package term

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// DisplayBackupSummary prints a summary of a created backup.
func DisplayBackupSummary(ios *iostreams.IOStreams, bkp *backup.DeviceBackup) {
	ios.Println()
	ios.Printf("  Device:     %s (%s, Gen%d)\n", bkp.Device().ID, bkp.Device().Model, bkp.Device().Generation)
	ios.Printf("  Firmware:   %s\n", bkp.Device().FWVersion)
	ios.Printf("  Config:     %d keys\n", bkp.ConfigKeyCount())
	if len(bkp.Components) > 0 {
		ios.Printf("  Components: %d\n", len(bkp.Components))
	}
	if len(bkp.Scripts) > 0 {
		ios.Printf("  Scripts:    %d\n", len(bkp.Scripts))
	}
	if n := bkp.ScheduleCount(); n > 0 {
		ios.Printf("  Schedules:  %d\n", n)
	}
	if n := bkp.WebhookCount(); n > 0 {
		ios.Printf("  Webhooks:   %d\n", n)
	}
	if len(bkp.KVS) > 0 {
		ios.Printf("  KVS:        %d\n", len(bkp.KVS))
	}
	if bkp.Encrypted() {
		ios.Printf("  Encrypted:  yes\n")
	}
}

// DisplayRestorePreview prints a preview of what would be restored.
func DisplayRestorePreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	DisplayBackupSource(ios, bkp)
	ios.Printf("Will restore:\n")
	displayConfigPreview(ios, bkp, opts)
	displayCountPreview(ios, "Scripts", len(bkp.Scripts), opts.SkipScripts)
	displayCountPreview(ios, "Schedules", bkp.ScheduleCount(), opts.SkipSchedules)
	displayCountPreview(ios, "Webhooks", bkp.WebhookCount(), opts.SkipWebhooks)
	displayCountPreview(ios, "KVS", len(bkp.KVS), opts.SkipKVS)
}

// DisplayBackupSource prints information about the backup source device.
func DisplayBackupSource(ios *iostreams.IOStreams, bkp *backup.DeviceBackup) {
	device := bkp.Device()
	ios.Printf("Backup source:\n")
	ios.Printf("  Device:    %s (%s, Gen%d)\n", device.ID, device.Model, device.Generation)
	ios.Printf("  Firmware:  %s\n", device.FWVersion)
	ios.Printf("  Created:   %s\n", bkp.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Println()
}

func displayConfigPreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	if len(bkp.Config) == 0 {
		return
	}

	var excluded []string
	if opts.SkipNetwork {
		excluded = append(excluded, "network")
	}
	if opts.SkipAuth {
		excluded = append(excluded, "auth")
	}

	switch len(excluded) {
	case 0:
		ios.Printf("  Config:    %d keys\n", bkp.ConfigKeyCount())
	default:
		ios.Printf("  Config:    %d keys (%s excluded)\n", bkp.ConfigKeyCount(), strings.Join(excluded, ", "))
	}
}

// displayCountPreview prints one restore-preview line for a counted section,
// and nothing when the backup holds none of it.
func displayCountPreview(ios *iostreams.IOStreams, label string, count int, skipped bool) {
	switch {
	case count == 0:
	case skipped:
		ios.Printf("  %-10s %d (skipped)\n", label+":", count)
	default:
		ios.Printf("  %-10s %d\n", label+":", count)
	}
}

// DisplayRestoreResult prints the results of a restore operation.
func DisplayRestoreResult(ios *iostreams.IOStreams, result *backup.RestoreResult) {
	ios.Println()
	if result.ConfigRestored {
		ios.Printf("  Config:    restored\n")
	}
	if result.ScriptsRestored > 0 {
		ios.Printf("  Scripts:   %d restored\n", result.ScriptsRestored)
	}
	if result.SchedulesRestored > 0 {
		ios.Printf("  Schedules: %d restored\n", result.SchedulesRestored)
	}
	if result.WebhooksRestored > 0 {
		ios.Printf("  Webhooks:  %d restored\n", result.WebhooksRestored)
	}

	// The list follows its header onto stderr so a piped run keeps the block together.
	if len(result.Warnings) > 0 {
		ios.Errorln()
		ios.Warning("Warnings:")
		for _, w := range result.Warnings {
			ios.Errorf("  - %s\n", w)
		}
	}

	if len(result.Errors) > 0 {
		ios.Errorln()
		ios.Error("Errors:")
		for _, e := range result.Errors {
			ios.Errorf("  - %s\n", e)
		}
	}

	if result.DestabilizedStep != "" {
		ios.Errorln()
		ios.Error("Restore halted: the device entered a reboot loop after the %q step.", result.DestabilizedStep)
	}
}

// RestoreResultError returns a non-nil error describing why a restore did not
// fully succeed, or nil when the device accepted the whole restore (or there is
// no result, for callers that short-circuit before restoring). Commands use it
// to gate success messaging and any destructive follow-up — a migrate source
// factory-reset must never run after a partial or failed target restore.
func RestoreResultError(target string, result *backup.RestoreResult) error {
	if err := result.Err(); err != nil {
		return fmt.Errorf("restore to %s failed: %w", target, err)
	}
	return nil
}

// ReportRestoreResult prints the outcome of a restore and returns a non-nil
// error when the device rejected any part of it, so the command exits non-zero
// instead of printing a false success. On success it prints the standard
// "Backup restored to <target>" line; either way it renders the per-section
// detail via DisplayRestoreResult.
func ReportRestoreResult(ios *iostreams.IOStreams, target string, result *backup.RestoreResult) error {
	if err := RestoreResultError(target, result); err != nil {
		ios.Error("Restore to %s did not complete cleanly:", target)
		DisplayRestoreResult(ios, result)
		return err
	}
	ios.Success("Backup restored to %s", target)
	if result != nil {
		DisplayRestoreResult(ios, result)
	}
	return nil
}

// DisplayBackupsTable prints a table of backup files.
func DisplayBackupsTable(ios *iostreams.IOStreams, backups []model.BackupFileInfo) {
	builder := output.FormatBackupsTable(backups)
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print backups table", err)
	}
}

// DisplayBackupExportResults prints the results of a backup export operation.
func DisplayBackupExportResults(ios *iostreams.IOStreams, results []shelly.BackupResult) {
	for _, r := range results {
		if r.Success {
			ios.Printf("  Backing up %s (%s)... OK\n", r.DeviceName, r.Address)
		} else {
			ios.Printf("  Backing up %s (%s)... FAILED\n", r.DeviceName, r.Address)
		}
	}
}
