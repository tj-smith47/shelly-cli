// Package open provides the cover open subcommand.
package open

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover open command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int
	var duration int

	cmd := &cobra.Command{
		Use:     "open <device>",
		Aliases: []string{"up", "raise"},
		Short:   "Open cover",
		Long:    `Open a cover/roller component on the specified device.`,
		Example: `  # Open cover fully
  shelly cover open bedroom

  # Open cover for 5 seconds
  shelly cover up bedroom --duration 5`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID, duration)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds (0 = full open)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID, duration int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Opening cover...")

	var dur *int
	if duration > 0 {
		dur = &duration
	}

	err := svc.CoverOpen(ctx, device, coverID, dur)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to open cover: %w", err)
	}

	ios.Success("Cover %d opening", coverID)
	return nil
}
