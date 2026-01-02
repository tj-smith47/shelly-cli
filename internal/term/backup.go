package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// DisplayBackupSummary prints a summary of a created backup.
func DisplayBackupSummary(ios *iostreams.IOStreams, bkp *backup.DeviceBackup) {
	ios.Println()
	ios.Printf("  Device:    %s (%s)\n", bkp.Device().ID, bkp.Device().Model)
	ios.Printf("  Firmware:  %s\n", bkp.Device().FWVersion)
	ios.Printf("  Config:    %d keys\n", len(bkp.Config))
	if len(bkp.Scripts) > 0 {
		ios.Printf("  Scripts:   %d\n", len(bkp.Scripts))
	}
	if len(bkp.Schedules) > 0 {
		ios.Printf("  Schedules: %d\n", len(bkp.Schedules))
	}
	if len(bkp.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d\n", len(bkp.Webhooks))
	}
	if bkp.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}
}

// DisplayRestorePreview prints a preview of what would be restored.
func DisplayRestorePreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	DisplayBackupSource(ios, bkp)
	ios.Printf("Will restore:\n")
	displayConfigPreview(ios, bkp, opts)
	displayScriptsPreview(ios, bkp, opts)
	displaySchedulesPreview(ios, bkp, opts)
	displayWebhooksPreview(ios, bkp, opts)
}

// DisplayBackupSource prints information about the backup source device.
func DisplayBackupSource(ios *iostreams.IOStreams, bkp *backup.DeviceBackup) {
	device := bkp.Device()
	ios.Printf("Backup source:\n")
	ios.Printf("  Device:    %s (%s)\n", device.ID, device.Model)
	ios.Printf("  Firmware:  %s\n", device.FWVersion)
	ios.Printf("  Created:   %s\n", bkp.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Println()
}

func displayConfigPreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	if len(bkp.Config) > 0 {
		if opts.SkipNetwork {
			ios.Printf("  Config:    %d keys (network config excluded)\n", len(bkp.Config))
		} else {
			ios.Printf("  Config:    %d keys\n", len(bkp.Config))
		}
	}
}

func displayScriptsPreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	if len(bkp.Scripts) > 0 {
		if opts.SkipScripts {
			ios.Printf("  Scripts:   %d (skipped)\n", len(bkp.Scripts))
		} else {
			ios.Printf("  Scripts:   %d\n", len(bkp.Scripts))
		}
	}
}

func displaySchedulesPreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	if len(bkp.Schedules) > 0 {
		if opts.SkipSchedules {
			ios.Printf("  Schedules: %d (skipped)\n", len(bkp.Schedules))
		} else {
			ios.Printf("  Schedules: %d\n", len(bkp.Schedules))
		}
	}
}

func displayWebhooksPreview(ios *iostreams.IOStreams, bkp *backup.DeviceBackup, opts backup.RestoreOptions) {
	if len(bkp.Webhooks) > 0 {
		if opts.SkipWebhooks {
			ios.Printf("  Webhooks:  %d (skipped)\n", len(bkp.Webhooks))
		} else {
			ios.Printf("  Webhooks:  %d\n", len(bkp.Webhooks))
		}
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

	if len(result.Warnings) > 0 {
		ios.Println()
		ios.Warning("Warnings:")
		for _, w := range result.Warnings {
			ios.Printf("  - %s\n", w)
		}
	}
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
