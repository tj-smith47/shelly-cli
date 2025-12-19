// Package power provides the monitor power subcommand for real-time power monitoring.
package power

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

var (
	intervalFlag time.Duration
	countFlag    int
)

// NewCommand creates the monitor power command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().DurationVarP(&intervalFlag, "interval", "i", 2*time.Second, "Refresh interval")
	cmd.Flags().IntVarP(&countFlag, "count", "n", 0, "Number of updates (0 = unlimited)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	opts := shelly.MonitorOptions{
		Interval: intervalFlag,
		Count:    countFlag,
	}

	ios.Title("Power Monitoring: %s", device)
	ios.Printf("Press Ctrl+C to stop\n\n")

	var lastSnapshot *shelly.MonitoringSnapshot
	return svc.MonitorDevice(ctx, device, opts, func(snapshot shelly.MonitoringSnapshot) error {
		term.DisplayPowerSnapshot(ios, &snapshot, lastSnapshot)
		lastSnapshot = &snapshot
		return nil
	})
}
