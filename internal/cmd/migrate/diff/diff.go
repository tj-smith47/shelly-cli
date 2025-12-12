// Package diff provides the migrate diff subcommand.
package diff

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the migrate diff command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <device> <backup-file>",
		Short: "Show differences between device and backup",
		Long: `Show the differences between a device's current state and a backup file.

This helps you understand what would change if you restored the backup.`,
		Example: `  # Show differences
  shelly migrate diff living-room backup.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], args[1])
		},
	}

	return cmd
}

func run(ctx context.Context, device, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := iostreams.System()

	// Read backup file
	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is user-provided CLI argument
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Compare with device
	spin := iostreams.NewSpinner("Comparing configurations...")
	spin.Start()

	svc := shelly.NewService()
	d, err := svc.CompareBackup(ctx, device, backup)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to compare: %w", err)
	}

	// Print header
	ios.Title("Configuration Differences")
	ios.Println()
	ios.Printf("Device: %s\n", device)
	ios.Printf("Backup: %s (%s, %s)\n", filePath, backup.Device().ID, backup.Device().Model)
	ios.Println()

	if !d.HasDifferences() {
		ios.Success("No differences found - device matches backup")
		return nil
	}

	// Print configuration differences
	if len(d.ConfigDiffs) > 0 {
		printConfigDiffs(ios, d.ConfigDiffs)
	}

	// Print script differences
	if len(d.ScriptDiffs) > 0 {
		printScriptDiffs(ios, d.ScriptDiffs)
	}

	// Print schedule differences
	if len(d.ScheduleDiffs) > 0 {
		printScheduleDiffs(ios, d.ScheduleDiffs)
	}

	// Print webhook differences
	if len(d.WebhookDiffs) > 0 {
		printWebhookDiffs(ios, d.WebhookDiffs)
	}

	return nil
}

const (
	diffAdded   = "added"
	diffRemoved = "removed"
	diffChanged = "changed"
)

func printConfigDiffs(ios *iostreams.IOStreams, diffs []shelly.ConfigDiff) {
	ios.Printf("Configuration:\n")
	for _, d := range diffs {
		switch d.DiffType {
		case diffAdded:
			ios.Printf("  + %s (will be added from backup)\n", d.Key)
		case diffRemoved:
			ios.Printf("  - %s (exists on device, not in backup)\n", d.Key)
		case diffChanged:
			ios.Printf("  ~ %s (will be updated)\n", d.Key)
		}
	}
	ios.Println()
}

func printScriptDiffs(ios *iostreams.IOStreams, diffs []shelly.ScriptDiff) {
	ios.Printf("Scripts:\n")
	for _, d := range diffs {
		switch d.DiffType {
		case diffAdded:
			ios.Printf("  + %s (will be created)\n", d.Name)
		case diffRemoved:
			ios.Printf("  - %s (exists on device, not in backup)\n", d.Name)
		case diffChanged:
			ios.Printf("  ~ %s (%s)\n", d.Name, d.Details)
		}
	}
	ios.Println()
}

func printScheduleDiffs(ios *iostreams.IOStreams, diffs []shelly.ScheduleDiff) {
	ios.Printf("Schedules:\n")
	for _, d := range diffs {
		switch d.DiffType {
		case diffAdded:
			ios.Printf("  + %s (will be created)\n", d.Timespec)
		case diffRemoved:
			ios.Printf("  - %s (exists on device, not in backup)\n", d.Timespec)
		case diffChanged:
			ios.Printf("  ~ %s (%s)\n", d.Timespec, d.Details)
		}
	}
	ios.Println()
}

func printWebhookDiffs(ios *iostreams.IOStreams, diffs []shelly.WebhookDiff) {
	ios.Printf("Webhooks:\n")
	for _, d := range diffs {
		switch d.DiffType {
		case diffAdded:
			ios.Printf("  + %s:%s (will be created)\n", d.Event, d.Name)
		case diffRemoved:
			ios.Printf("  - %s:%s (exists on device, not in backup)\n", d.Event, d.Name)
		case diffChanged:
			ios.Printf("  ~ %s:%s (%s)\n", d.Event, d.Name, d.Details)
		}
	}
	ios.Println()
}
