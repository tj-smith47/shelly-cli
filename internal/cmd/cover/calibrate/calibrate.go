// Package calibrate provides the cover calibrate subcommand.
package calibrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the cover calibrate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:     "calibrate <device>",
		Aliases: []string{"cal"},
		Short:   "Calibrate cover",
		Long: `Start calibration for a cover/roller component.

Calibration determines the open and close times for the cover.
The cover will move to both extremes during calibration.`,
		Example: `  # Calibrate a cover
  shelly cover calibrate bedroom

  # Calibrate specific cover ID
  shelly cover cal bedroom --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	ios.StartProgress("Starting calibration...")

	err := svc.CoverCalibrate(ctx, device, coverID)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to start calibration: %w", err)
	}

	ios.Success("Cover %d calibration started", coverID)
	return nil
}
