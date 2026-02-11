// Package restore provides the backup restore subcommand.
package restore

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory       *cmdutil.Factory
	Decrypt       string
	Device        string
	DryRun        bool
	FilePath      string
	SkipAuth      bool
	SkipNetwork   bool
	SkipScripts   bool
	SkipSchedules bool
	SkipWebhooks  bool
}

// NewCommand creates the backup restore command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "restore <device> <file>",
		Aliases: []string{"apply", "load"},
		Short:   "Restore a device from backup",
		Long: `Restore a Shelly device from a backup file.

By default, everything from the backup is restored including network
and authentication settings. Use --skip-* flags to exclude specific
sections.`,
		Example: `  # Full restore from backup
  shelly backup restore living-room backup.json

  # Dry run - show what would change
  shelly backup restore living-room backup.json --dry-run

  # Restore without network config (keep current WiFi)
  shelly backup restore living-room backup.json --skip-network

  # Restore without auth config
  shelly backup restore living-room backup.json --skip-auth

  # Restore encrypted backup
  shelly backup restore living-room backup.json --decrypt mysecret

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be restored without applying")
	cmd.Flags().BoolVar(&opts.SkipAuth, "skip-auth", false, "Skip authentication configuration")
	cmd.Flags().BoolVar(&opts.SkipNetwork, "skip-network", false, "Skip network configuration (WiFi, Ethernet)")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Skip script restoration")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Skip schedule restoration")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Skip webhook restoration")
	cmd.Flags().StringVarP(&opts.Decrypt, "decrypt", "d", "", "Password to decrypt backup")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*5)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Resolve file path (check backups dir if not found as-is)
	opts.FilePath = backup.ResolveFilePath(opts.FilePath)

	// Read backup file
	data, err := afero.ReadFile(config.Fs(), opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate backup
	bkp, err := backup.Validate(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Check encryption
	if bkp.Encrypted() && opts.Decrypt == "" {
		return fmt.Errorf("backup is encrypted, use --decrypt to provide password")
	}

	restoreOpts := backup.RestoreOptions{
		DryRun:        opts.DryRun,
		SkipAuth:      opts.SkipAuth,
		SkipNetwork:   opts.SkipNetwork,
		SkipScripts:   opts.SkipScripts,
		SkipSchedules: opts.SkipSchedules,
		SkipWebhooks:  opts.SkipWebhooks,
		Password:      opts.Decrypt,
	}

	if opts.DryRun {
		ios.Title("Dry run - Restore preview")
		ios.Println()
		term.DisplayRestorePreview(ios, bkp, restoreOpts)
		return nil
	}

	svc := opts.Factory.ShellyService()
	var result *backup.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Restoring backup...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, opts.Device, bkp, restoreOpts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Print results
	ios.Success("Backup restored to %s", opts.Device)
	term.DisplayRestoreResult(ios, result)

	return nil
}
