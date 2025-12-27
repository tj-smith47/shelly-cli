// Package diff provides the migrate diff subcommand.
package diff

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the migrate diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diff <device> <backup-file>",
		Aliases: []string{"compare", "cmp"},
		Short:   "Show differences between device and backup",
		Long: `Show the differences between a device's current state and a backup file.

This helps you understand what would change if you restored the backup.`,
		Example: `  # Show differences
  shelly migrate diff living-room backup.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := f.IOStreams()

	// Read backup file
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	bkp, err := backup.Validate(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Compare with device
	svc := f.ShellyService()
	var d *model.BackupDiff
	err = cmdutil.RunWithSpinner(ctx, ios, "Comparing configurations...", func(ctx context.Context) error {
		var cmpErr error
		d, cmpErr = svc.CompareBackup(ctx, device, bkp)
		return cmpErr
	})
	if err != nil {
		return fmt.Errorf("failed to compare: %w", err)
	}

	// Print header
	ios.Title("Configuration Differences")
	ios.Println()
	ios.Printf("Device: %s\n", device)
	ios.Printf("Backup: %s (%s, %s)\n", filePath, bkp.Device().ID, bkp.Device().Model)
	ios.Println()

	if !d.HasDifferences() {
		ios.Success("No differences found - device matches backup")
		return nil
	}

	// Print differences using consolidated display helpers
	term.DisplayConfigDiffs(ios, d.ConfigDiffs, true)
	term.DisplayScriptDiffs(ios, d.ScriptDiffs, true)
	term.DisplayScheduleDiffs(ios, d.ScheduleDiffs, true)
	term.DisplayWebhookDiffs(ios, d.WebhookDiffs, true)

	return nil
}
