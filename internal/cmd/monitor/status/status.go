// Package status provides the monitor status subcommand for real-time device monitoring.
package status

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

var (
	intervalFlag time.Duration
	countFlag    int
)

// NewCommand creates the monitor status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "watch"},
		Short:   "Monitor device status in real-time",
		Long: `Monitor a device's status in real-time with automatic refresh.

Status includes switches, covers, lights, energy meters, and other components.
Press Ctrl+C to stop monitoring.`,
		Example: `  # Monitor device status every 2 seconds
  shelly monitor status living-room

  # Monitor with custom interval
  shelly monitor status living-room --interval 5s

  # Monitor for a specific number of updates
  shelly monitor status living-room --count 10`,
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

	opts := shelly.MonitoringOptions{
		Interval: intervalFlag,
		Count:    countFlag,
	}

	ios.Title("Monitoring %s", device)
	ios.Printf("Press Ctrl+C to stop\n\n")

	var lastSnapshot *model.MonitoringSnapshot
	return svc.MonitorDevice(ctx, device, opts, func(snapshot model.MonitoringSnapshot) error {
		term.DisplayStatusSnapshot(ios, &snapshot, lastSnapshot)
		lastSnapshot = &snapshot
		return nil
	})
}
