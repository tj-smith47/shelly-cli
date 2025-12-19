package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// DisplayBackupSummary prints a summary of a created backup.
func DisplayBackupSummary(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	ios.Println()
	ios.Printf("  Device:    %s (%s)\n", backup.Device().ID, backup.Device().Model)
	ios.Printf("  Firmware:  %s\n", backup.Device().FWVersion)
	ios.Printf("  Config:    %d keys\n", len(backup.Config))
	if len(backup.Scripts) > 0 {
		ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
	}
	if len(backup.Schedules) > 0 {
		ios.Printf("  Schedules: %d\n", len(backup.Schedules))
	}
	if len(backup.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
	}
	if backup.Encrypted() {
		ios.Printf("  Encrypted: yes\n")
	}
}

// DisplayRestorePreview prints a preview of what would be restored.
func DisplayRestorePreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	DisplayBackupSource(ios, backup)
	ios.Printf("Will restore:\n")
	displayConfigPreview(ios, backup, opts)
	displayScriptsPreview(ios, backup, opts)
	displaySchedulesPreview(ios, backup, opts)
	displayWebhooksPreview(ios, backup, opts)
}

// DisplayBackupSource prints information about the backup source device.
func DisplayBackupSource(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	device := backup.Device()
	ios.Printf("Backup source:\n")
	ios.Printf("  Device:    %s (%s)\n", device.ID, device.Model)
	ios.Printf("  Firmware:  %s\n", device.FWVersion)
	ios.Printf("  Created:   %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Println()
}

func displayConfigPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Config) > 0 {
		if opts.SkipNetwork {
			ios.Printf("  Config:    %d keys (network config excluded)\n", len(backup.Config))
		} else {
			ios.Printf("  Config:    %d keys\n", len(backup.Config))
		}
	}
}

func displayScriptsPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Scripts) > 0 {
		if opts.SkipScripts {
			ios.Printf("  Scripts:   %d (skipped)\n", len(backup.Scripts))
		} else {
			ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
		}
	}
}

func displaySchedulesPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Schedules) > 0 {
		if opts.SkipSchedules {
			ios.Printf("  Schedules: %d (skipped)\n", len(backup.Schedules))
		} else {
			ios.Printf("  Schedules: %d\n", len(backup.Schedules))
		}
	}
}

func displayWebhooksPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Webhooks) > 0 {
		if opts.SkipWebhooks {
			ios.Printf("  Webhooks:  %d (skipped)\n", len(backup.Webhooks))
		} else {
			ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
		}
	}
}

// DisplayRestoreResult prints the results of a restore operation.
func DisplayRestoreResult(ios *iostreams.IOStreams, result *shelly.RestoreResult) {
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
	table := output.FormatBackupsTable(backups)
	printTable(ios, table)
}

// DisplayBackupExportResults prints the results of a backup export operation.
func DisplayBackupExportResults(ios *iostreams.IOStreams, results []export.BackupResult) {
	for _, r := range results {
		if r.Success {
			ios.Printf("  Backing up %s (%s)... OK\n", r.DeviceName, r.Address)
		} else {
			ios.Printf("  Backing up %s (%s)... FAILED\n", r.DeviceName, r.Address)
		}
	}
}
