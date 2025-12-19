// Package migrate provides migration commands.
package migrate

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/migrate/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
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
	backup, sourceType, err := loadSource(ctx, svc, ios, source)
	if err != nil {
		return err
	}

	// Check target device compatibility
	if err := checkTargetCompatibility(ctx, svc, ios, backup, target); err != nil {
		return err
	}

	if dryRunFlag {
		return showDryRunPreview(ctx, svc, ios, source, sourceType, target, backup)
	}

	// Perform migration
	ios.StartProgress("Migrating configuration...")

	opts := shelly.RestoreOptions{
		SkipNetwork: true, // Always skip network to prevent disconnection
	}
	result, err := svc.RestoreBackup(ctx, target, backup, opts)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	ios.Success("Migration completed!")
	printResult(ios, result)

	return nil
}

func loadSource(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, source string) (*shelly.DeviceBackup, string, error) {
	if fileExists(source) {
		return loadFromFile(source)
	}
	return loadFromDevice(ctx, svc, ios, source)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func loadFromFile(source string) (*shelly.DeviceBackup, string, error) {
	data, err := os.ReadFile(source) //nolint:gosec // G304: source is user-provided CLI argument
	if err != nil {
		return nil, "", fmt.Errorf("failed to read backup file: %w", err)
	}
	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		return nil, "", fmt.Errorf("invalid backup file: %w", err)
	}
	return backup, "file", nil
}

func loadFromDevice(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, source string) (*shelly.DeviceBackup, string, error) {
	ios.StartProgress("Reading source device...")
	backup, err := svc.CreateBackup(ctx, source, shelly.BackupOptions{})
	ios.StopProgress()
	if err != nil {
		return nil, "", fmt.Errorf("failed to read source device: %w", err)
	}
	return backup, "device", nil
}

func checkTargetCompatibility(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, backup *shelly.DeviceBackup, target string) error {
	targetInfo, err := svc.DeviceInfo(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to get target device info: %w", err)
	}

	if !forceFlag && backup.Device().Model != targetInfo.Model {
		ios.Warning("Source and target are different device types:")
		ios.Printf("  Source: %s\n", backup.Device().Model)
		ios.Printf("  Target: %s\n", targetInfo.Model)
		ios.Info("Use --force to migrate anyway")
		return fmt.Errorf("device type mismatch")
	}
	return nil
}

func showDryRunPreview(ctx context.Context, svc *shelly.Service, ios *iostreams.IOStreams, source, sourceType, target string, backup *shelly.DeviceBackup) error {
	ios.Title("Migration Preview (dry run)")
	ios.Println()
	ios.Printf("Source: %s (%s)\n", source, sourceType)
	ios.Printf("Target: %s\n", target)
	ios.Println()

	d, err := svc.CompareBackup(ctx, target, backup)
	if err != nil {
		return fmt.Errorf("failed to compare: %w", err)
	}

	if !d.HasDifferences() {
		ios.Info("No differences found - target already matches source")
		return nil
	}

	printDiff(ios, d)
	return nil
}

func printDiff(ios *iostreams.IOStreams, d *model.BackupDiff) {
	// Use consolidated display functions with verbose=false for concise output
	term.DisplayConfigDiffs(ios, d.ConfigDiffs, false)
	term.DisplayScriptDiffs(ios, d.ScriptDiffs, false)
	term.DisplayScheduleDiffs(ios, d.ScheduleDiffs, false)
	term.DisplayWebhookDiffs(ios, d.WebhookDiffs, false)
}

func printResult(ios *iostreams.IOStreams, result *shelly.RestoreResult) {
	ios.Println()
	if result.ConfigRestored {
		ios.Printf("  Config:    migrated\n")
	}
	if result.ScriptsRestored > 0 {
		ios.Printf("  Scripts:   %d migrated\n", result.ScriptsRestored)
	}
	if result.SchedulesRestored > 0 {
		ios.Printf("  Schedules: %d migrated\n", result.SchedulesRestored)
	}
	if result.WebhooksRestored > 0 {
		ios.Printf("  Webhooks:  %d migrated\n", result.WebhooksRestored)
	}

	if len(result.Warnings) > 0 {
		ios.Println()
		ios.Warning("Warnings:")
		for _, w := range result.Warnings {
			ios.Printf("  - %s\n", w)
		}
	}
}
