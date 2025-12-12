// Package restore provides the backup restore subcommand.
package restore

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	dryRunFlag        bool
	skipNetworkFlag   bool
	skipScriptsFlag   bool
	skipSchedulesFlag bool
	skipWebhooksFlag  bool
	decryptFlag       string
)

// NewCommand creates the backup restore command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <device> <file>",
		Short: "Restore a device from backup",
		Long: `Restore a Shelly device from a backup file.

By default, all data from the backup is restored. Use --skip-* flags
to exclude specific sections.

Network configuration (WiFi, Ethernet) is skipped by default with
--skip-network to prevent losing connectivity.`,
		Example: `  # Restore from backup (skip network config)
  shelly backup restore living-room backup.json

  # Dry run - show what would change
  shelly backup restore living-room backup.json --dry-run

  # Restore everything including network config
  shelly backup restore living-room backup.json --skip-network=false

  # Restore encrypted backup
  shelly backup restore living-room backup.json --decrypt mysecret

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be restored without applying")
	cmd.Flags().BoolVar(&skipNetworkFlag, "skip-network", true, "Skip network configuration (WiFi, Ethernet)")
	cmd.Flags().BoolVar(&skipScriptsFlag, "skip-scripts", false, "Skip script restoration")
	cmd.Flags().BoolVar(&skipSchedulesFlag, "skip-schedules", false, "Skip schedule restoration")
	cmd.Flags().BoolVar(&skipWebhooksFlag, "skip-webhooks", false, "Skip webhook restoration")
	cmd.Flags().StringVarP(&decryptFlag, "decrypt", "d", "", "Password to decrypt backup")

	return cmd
}

func run(ctx context.Context, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*5)
	defer cancel()

	ios := iostreams.System()

	// Read backup file
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate backup
	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Check encryption
	if backup.Encrypted && decryptFlag == "" {
		return fmt.Errorf("backup is encrypted, use --decrypt to provide password")
	}

	opts := shelly.RestoreOptions{
		DryRun:        dryRunFlag,
		SkipNetwork:   skipNetworkFlag,
		SkipScripts:   skipScriptsFlag,
		SkipSchedules: skipSchedulesFlag,
		SkipWebhooks:  skipWebhooksFlag,
		Password:      decryptFlag,
	}

	if dryRunFlag {
		ios.Title("Dry run - Restore preview")
		ios.Println()
		printRestorePreview(ios, backup, opts)
		return nil
	}

	spin := iostreams.NewSpinner("Restoring backup...")
	spin.Start()

	svc := shelly.NewService()
	result, err := svc.RestoreBackup(ctx, device, backup, opts)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Print results
	ios.Success("Backup restored to %s", device)
	printRestoreResult(ios, result)

	return nil
}

const (
	statusEnabled  = "enabled"
	statusDisabled = "disabled"
)

func enabledStatus(enabled bool) string {
	if enabled {
		return statusEnabled
	}
	return statusDisabled
}

func printRestorePreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	printBackupSource(ios, backup)
	ios.Printf("Will restore:\n")
	printConfigPreview(ios, backup, opts)
	printScriptsPreview(ios, backup, opts)
	printSchedulesPreview(ios, backup, opts)
	printWebhooksPreview(ios, backup, opts)
}

func printBackupSource(ios *iostreams.IOStreams, backup *shelly.DeviceBackup) {
	ios.Printf("Backup source:\n")
	ios.Printf("  Device:    %s (%s)\n", backup.Device.ID, backup.Device.Model)
	ios.Printf("  Firmware:  %s\n", backup.Device.FWVersion)
	ios.Printf("  Created:   %s\n", backup.CreatedAt.Format("2006-01-02 15:04:05"))
	ios.Println()
}

func printConfigPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if len(backup.Config) > 0 {
		if opts.SkipNetwork {
			ios.Printf("  Config:    %d keys (network config excluded)\n", len(backup.Config))
		} else {
			ios.Printf("  Config:    %d keys\n", len(backup.Config))
		}
	}
}

func printScriptsPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if !opts.SkipScripts && len(backup.Scripts) > 0 {
		ios.Printf("  Scripts:   %d\n", len(backup.Scripts))
		for _, s := range backup.Scripts {
			ios.Printf("    - %s (%s)\n", s.Name, enabledStatus(s.Enable))
		}
	} else if opts.SkipScripts && len(backup.Scripts) > 0 {
		ios.Printf("  Scripts:   %d (skipped)\n", len(backup.Scripts))
	}
}

func printSchedulesPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if !opts.SkipSchedules && len(backup.Schedules) > 0 {
		ios.Printf("  Schedules: %d\n", len(backup.Schedules))
		for _, s := range backup.Schedules {
			ios.Printf("    - %s (%s)\n", s.Timespec, enabledStatus(s.Enable))
		}
	} else if opts.SkipSchedules && len(backup.Schedules) > 0 {
		ios.Printf("  Schedules: %d (skipped)\n", len(backup.Schedules))
	}
}

func printWebhooksPreview(ios *iostreams.IOStreams, backup *shelly.DeviceBackup, opts shelly.RestoreOptions) {
	if !opts.SkipWebhooks && len(backup.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d\n", len(backup.Webhooks))
		for _, w := range backup.Webhooks {
			ios.Printf("    - %s: %s (%s)\n", w.Event, w.Name, enabledStatus(w.Enable))
		}
	} else if opts.SkipWebhooks && len(backup.Webhooks) > 0 {
		ios.Printf("  Webhooks:  %d (skipped)\n", len(backup.Webhooks))
	}
}

func printRestoreResult(ios *iostreams.IOStreams, result *shelly.RestoreResult) {
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
