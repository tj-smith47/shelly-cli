// Package power provides the monitor power subcommand for real-time power monitoring.
package power

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Count    int
	Device   string
	Interval time.Duration
}

// NewCommand creates the monitor power command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		Interval: 2 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "power <device>",
		Aliases: []string{"pwr", "watt"},
		Short:   "Monitor power consumption in real-time",
		Long: `Monitor a device's power consumption in real-time.

Shows power (W), voltage (V), current (A), and energy (Wh) for all
energy meters and power meters on the device.
Press Ctrl+C to stop monitoring.`,
		Example: `  # Monitor power consumption
  shelly monitor power living-room

  # Monitor with 1-second interval
  shelly monitor power living-room --interval 1s

  # Monitor for a specific number of updates
  shelly monitor power living-room --count 10`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Interval, "interval", "i", 2*time.Second, "Refresh interval")
	cmd.Flags().IntVarP(&opts.Count, "count", "n", 0, "Number of updates (0 = unlimited)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	monitorOpts := shelly.MonitoringOptions{
		Interval: opts.Interval,
		Count:    opts.Count,
	}

	ios.Title("Power Monitoring: %s", opts.Device)
	ios.Printf("Press Ctrl+C to stop\n\n")

	var lastSnapshot *model.MonitoringSnapshot
	return svc.MonitorDevice(ctx, opts.Device, monitorOpts, func(snapshot model.MonitoringSnapshot) error {
		term.DisplayPowerSnapshot(ios, &snapshot, lastSnapshot)
		lastSnapshot = &snapshot
		return nil
	})
}
