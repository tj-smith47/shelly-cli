// Package calibrate provides the cover calibrate subcommand.
package calibrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the cover calibrate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Cover")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, "Starting calibration...", func(ctx context.Context) error {
		return svc.CoverCalibrate(ctx, opts.Device, opts.ID)
	})
	if err != nil {
		return fmt.Errorf("failed to start calibration: %w", err)
	}

	ios.Success("Cover %d calibration started", opts.ID)
	return nil
}
