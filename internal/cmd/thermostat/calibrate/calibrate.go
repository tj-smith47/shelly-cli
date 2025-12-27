// Package calibrate provides the thermostat calibrate command.
package calibrate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the thermostat calibrate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "calibrate <device>",
		Aliases: []string{"cal"},
		Short:   "Calibrate thermostat valve",
		Long: `Initiate valve calibration on a thermostat.

Calibration helps the thermostat learn the full range of valve
movement. This should be performed:
- After initial installation
- If the valve behavior seems incorrect
- After physical maintenance on the valve

The calibration process takes a few minutes. The valve will
move through its full range to determine open/close positions.`,
		Example: `  # Calibrate thermostat
  shelly thermostat calibrate gateway

  # Calibrate specific thermostat
  shelly thermostat calibrate gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Thermostat")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		thermostat := dev.Gen2().Thermostat(opts.ID)

		err := cmdutil.RunWithSpinner(ctx, ios, "Starting valve calibration...", func(ctx context.Context) error {
			return thermostat.Calibrate(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to start calibration: %w", err)
		}

		ios.Success("Calibration started on thermostat %d", opts.ID)
		ios.Info("The valve will move through its full range.")
		ios.Info("This process takes a few minutes to complete.")
		ios.Info("Check status with: shelly thermostat status %s", opts.Device)

		return nil
	})
}
