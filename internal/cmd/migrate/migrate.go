// Package migrate provides migration commands.
package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory       *cmdutil.Factory
	Source        string
	Target        string
	DryRun        bool
	Force         bool
	SkipAuth      bool
	SkipNetwork   bool
	SkipScripts   bool
	SkipSchedules bool
	SkipWebhooks  bool
}

// NewCommand creates the migrate command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "migrate <source> <target>",
		Aliases: []string{"mig"},
		Short:   "Migrate configuration between devices",
		Long: `Migrate configuration from a source device or backup file to a target device.

By default, everything is migrated including network and authentication
settings. Use --skip-* flags to exclude specific sections.

Source can be a device name/address or a backup file path.
The --dry-run flag shows what would be changed without applying.`,
		Example: `  # Migrate from one device to another
  shelly migrate living-room bedroom --dry-run

  # Migrate from backup file to device
  shelly migrate backup.json bedroom

  # Migrate without network config (keep current WiFi)
  shelly migrate living-room bedroom --skip-network

  # Force migration between different device types
  shelly migrate backup.json bedroom --force`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Source = args[0]
			opts.Target = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be changed without applying")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force migration between different device types")
	cmd.Flags().BoolVar(&opts.SkipAuth, "skip-auth", false, "Skip authentication configuration")
	cmd.Flags().BoolVar(&opts.SkipNetwork, "skip-network", false, "Skip network configuration (WiFi, Ethernet)")
	cmd.Flags().BoolVar(&opts.SkipScripts, "skip-scripts", false, "Skip script migration")
	cmd.Flags().BoolVar(&opts.SkipSchedules, "skip-schedules", false, "Skip schedule migration")
	cmd.Flags().BoolVar(&opts.SkipWebhooks, "skip-webhooks", false, "Skip webhook migration")

	cmd.AddCommand(validate.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*3)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Load source backup (from file or device)
	var bkp *backup.DeviceBackup
	var sourceType backup.MigrationSource
	var err error

	if backup.IsFile(opts.Source) {
		err = cmdutil.RunWithSpinner(ctx, ios, "Reading backup file...", func(_ context.Context) error {
			bkp, sourceType, err = svc.LoadMigrationSource(ctx, opts.Source)
			return err
		})
	} else {
		err = cmdutil.RunWithSpinner(ctx, ios, "Reading source device...", func(ctx context.Context) error {
			bkp, sourceType, err = svc.LoadMigrationSource(ctx, opts.Source)
			return err
		})
	}
	if err != nil {
		return err
	}

	// Check target device compatibility
	if err := svc.CheckMigrationCompatibility(ctx, bkp, opts.Target, opts.Force); err != nil {
		var compErr *backup.CompatibilityError
		if errors.As(err, &compErr) {
			ios.Warning("Source and target are different device types:")
			ios.Printf("  Source: %s\n", compErr.SourceModel)
			ios.Printf("  Target: %s\n", compErr.TargetModel)
			ios.Info("Use --force to migrate anyway")
			return fmt.Errorf("device type mismatch")
		}
		return err
	}

	if opts.DryRun {
		d, err := svc.CompareBackup(ctx, opts.Target, bkp)
		if err != nil {
			return fmt.Errorf("failed to compare: %w", err)
		}
		term.DisplayMigrationPreview(ios, opts.Source, string(sourceType), opts.Target, d)
		return nil
	}

	// Perform migration
	restoreOpts := backup.RestoreOptions{
		SkipAuth:      opts.SkipAuth,
		SkipNetwork:   opts.SkipNetwork,
		SkipScripts:   opts.SkipScripts,
		SkipSchedules: opts.SkipSchedules,
		SkipWebhooks:  opts.SkipWebhooks,
	}
	var result *backup.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Migrating configuration...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, opts.Target, bkp, restoreOpts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ios.Success("Migration completed!")
	term.DisplayMigrationResult(ios, result)

	return nil
}
