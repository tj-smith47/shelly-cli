// Package diff provides the device config diff subcommand.
package diff

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the device config diff command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diff <source> <target>",
		Aliases: []string{"compare", "cmp"},
		Short:   "Compare device configurations",
		Long: `Compare configurations between two devices or a device and a backup file.

This command shows differences in configuration between:
  - Two live devices: shelly device config diff device1 device2
  - Device and backup: shelly device config diff device backup.json
  - Two backup files: shelly device config diff backup1.json backup2.json

Differences are shown with:
  + Added values (only in target)
  - Removed values (only in source)
  ~ Changed values (different between source and target)`,
		Example: `  # Compare two devices
  shelly device config diff kitchen-light bedroom-light

  # Compare device with backup file
  shelly device config diff living-room config-backup.json

  # Compare two backup files
  shelly device config diff backup1.json backup2.json

  # JSON output
  shelly device config diff device1 device2 -o json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, source, target string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.StartProgress("Fetching configurations...")
	sourceConfig, sourceName, err := svc.LoadConfig(ctx, source)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to get source config: %w", err)
	}

	targetConfig, targetName, err := svc.LoadConfig(ctx, target)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to get target config: %w", err)
	}
	ios.StopProgress()

	diffs := output.CompareConfigs(sourceConfig, targetConfig)

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, diffs)
	}

	ios.Printf("Comparing: %s ↔ %s\n\n", sourceName, targetName)
	if len(diffs) == 0 {
		ios.Success("Configurations are identical")
		return nil
	}

	term.DisplayConfigDiffsSummary(ios, diffs)
	return nil
}
