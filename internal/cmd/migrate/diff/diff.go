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

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	FilePath string
}

// NewCommand creates the migrate diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			opts.FilePath = args[1]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Read backup file (user-provided path from CLI argument)
	data, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	bkp, err := backup.Validate(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Compare with device
	svc := opts.Factory.ShellyService()
	var d *model.BackupDiff
	err = cmdutil.RunWithSpinner(ctx, ios, "Comparing configurations...", func(ctx context.Context) error {
		var cmpErr error
		d, cmpErr = svc.CompareBackup(ctx, opts.Device, bkp)
		return cmpErr
	})
	if err != nil {
		return fmt.Errorf("failed to compare: %w", err)
	}

	// Print header
	ios.Title("Configuration Differences")
	ios.Println()
	ios.Printf("Device: %s\n", opts.Device)
	ios.Printf("Backup: %s (%s, %s)\n", opts.FilePath, bkp.Device().ID, bkp.Device().Model)
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
