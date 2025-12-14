package roller

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CalibrateOptions holds calibrate command options.
type CalibrateOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newCalibrateCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &CalibrateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "calibrate <device>",
		Aliases: []string{"cal"},
		Short:   "Calibrate roller positioning",
		Long: `Start the calibration procedure for a Gen1 roller/cover.

Calibration measures the time required to fully open and close the roller.
This is required for position control to work accurately.

During calibration:
1. The roller will move to fully closed position
2. Then it will move to fully open position
3. Times are recorded for accurate position control

Warning: Make sure the roller has no obstructions before calibrating.`,
		Example: `  # Start calibration
  shelly gen1 roller calibrate living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runCalibrate(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")

	return cmd
}

func runCalibrate(ctx context.Context, opts *CalibrateOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	roller, err := gen1Client.Roller(opts.ID)
	if err != nil {
		return err
	}

	ios.Info("Starting roller calibration...")
	ios.Info("The roller will move to fully closed, then fully open.")

	err = roller.Calibrate(ctx)
	if err != nil {
		return err
	}

	ios.Success("Roller %d calibration started", opts.ID)
	ios.Info("Use 'shelly gen1 roller status %s' to monitor progress.", opts.Device)

	return nil
}
