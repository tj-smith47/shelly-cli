// Package restore provides the backup restore subcommand.
package restore

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
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
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restore <device> <file>",
		Aliases: []string{"apply", "load"},
		Short:   "Restore a device from backup",
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
			return run(cmd.Context(), f, args[0], args[1])
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

func run(ctx context.Context, f *cmdutil.Factory, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*5)
	defer cancel()

	ios := f.IOStreams()

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
	if backup.Encrypted() && decryptFlag == "" {
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
		term.DisplayRestorePreview(ios, backup, opts)
		return nil
	}

	svc := f.ShellyService()
	var result *shelly.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Restoring backup...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, device, backup, opts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Print results
	ios.Success("Backup restored to %s", device)
	term.DisplayRestoreResult(ios, result)

	return nil
}
