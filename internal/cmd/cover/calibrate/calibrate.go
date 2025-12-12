// Package calibrate provides the cover calibrate subcommand.
package calibrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the cover calibrate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var coverID int

	cmd := &cobra.Command{
		Use:   "calibrate <device>",
		Short: "Calibrate cover",
		Long: `Start calibration for a cover/roller component.

Calibration determines the open and close times for the cover.
The cover will move to both extremes during calibration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], coverID)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &coverID, "Cover")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, coverID int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()

	spin := iostreams.NewSpinner("Starting calibration...")
	spin.Start()

	err := svc.CoverCalibrate(ctx, device, coverID)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to start calibration: %w", err)
	}

	iostreams.Success("Cover %d calibration started", coverID)
	return nil
}
