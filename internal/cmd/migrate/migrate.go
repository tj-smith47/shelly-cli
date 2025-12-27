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
	"github.com/tj-smith47/shelly-cli/internal/term"
)

var (
	dryRunFlag bool
	forceFlag  bool
)

// NewCommand creates the migrate command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate <source> <target>",
		Aliases: []string{"mig"},
		Short:   "Migrate configuration between devices",
		Long: `Migrate configuration from a source device or backup file to a target device.

Source can be a device name/address or a backup file path.
The --dry-run flag shows what would be changed without applying.`,
		Example: `  # Migrate from one device to another
  shelly migrate living-room bedroom --dry-run

  # Migrate from backup file to device
  shelly migrate backup.json bedroom

  # Force migration between different device types
  shelly migrate backup.json bedroom --force`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be changed without applying")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force migration between different device types")

	cmd.AddCommand(validate.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, source, target string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*3)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Load source backup (from file or device)
	var backup *shelly.DeviceBackup
	var sourceType shelly.MigrationSource
	var err error

	if shelly.IsBackupFile(source) {
		err = cmdutil.RunWithSpinner(ctx, ios, "Reading backup file...", func(_ context.Context) error {
			backup, sourceType, err = svc.LoadMigrationSource(ctx, source)
			return err
		})
	} else {
		err = cmdutil.RunWithSpinner(ctx, ios, "Reading source device...", func(ctx context.Context) error {
			backup, sourceType, err = svc.LoadMigrationSource(ctx, source)
			return err
		})
	}
	if err != nil {
		return err
	}

	// Check target device compatibility
	if err := svc.CheckMigrationCompatibility(ctx, backup, target, forceFlag); err != nil {
		var compErr *shelly.MigrationCompatibilityError
		if errors.As(err, &compErr) {
			ios.Warning("Source and target are different device types:")
			ios.Printf("  Source: %s\n", compErr.SourceModel)
			ios.Printf("  Target: %s\n", compErr.TargetModel)
			ios.Info("Use --force to migrate anyway")
			return fmt.Errorf("device type mismatch")
		}
		return err
	}

	if dryRunFlag {
		d, err := svc.CompareBackup(ctx, target, backup)
		if err != nil {
			return fmt.Errorf("failed to compare: %w", err)
		}
		term.DisplayMigrationPreview(ios, source, string(sourceType), target, d)
		return nil
	}

	// Perform migration
	opts := shelly.RestoreOptions{
		SkipNetwork: true, // Always skip network to prevent disconnection
	}
	var result *shelly.RestoreResult
	err = cmdutil.RunWithSpinner(ctx, ios, "Migrating configuration...", func(ctx context.Context) error {
		var restoreErr error
		result, restoreErr = svc.RestoreBackup(ctx, target, backup, opts)
		return restoreErr
	})
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ios.Success("Migration completed!")
	term.DisplayMigrationResult(ios, result)

	return nil
}
