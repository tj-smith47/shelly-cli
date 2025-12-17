// Package status provides the monitor status subcommand for real-time device monitoring.
package status

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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

	opts := shelly.MonitorOptions{
		Interval: intervalFlag,
		Count:    countFlag,
	}

	ios.Title("Monitoring %s", device)
	ios.Printf("Press Ctrl+C to stop\n\n")

	var lastSnapshot *shelly.MonitoringSnapshot
	return svc.MonitorDevice(ctx, device, opts, func(snapshot shelly.MonitoringSnapshot) error {
		displaySnapshot(ios, &snapshot, lastSnapshot)
		lastSnapshot = &snapshot
		return nil
	})
}

func displaySnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates (simple approach)
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Device Status")
	ios.Printf("  Timestamp: %s\n\n", current.Timestamp.Format(time.RFC3339))

	// Display energy meters
	displayEMStatus(ios, current.EM, previous)
	displayEM1Status(ios, current.EM1, previous)

	// Display power meters
	displayPMStatus(ios, current.PM, previous)

	ios.Println()
}

func displayEMStatus(ios *iostreams.IOStreams, statuses []shelly.EMStatus, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Energy Meters (3-phase):\n")
	for i := range statuses {
		em := &statuses[i]
		prev := output.FindPreviousEM(em.ID, previous)
		for _, line := range output.FormatEMLines(em, prev) {
			ios.Println(line)
		}
	}
	ios.Println()
}

func displayEM1Status(ios *iostreams.IOStreams, statuses []shelly.EM1Status, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Energy Meters (single-phase):\n")
	for i := range statuses {
		em1 := &statuses[i]
		prev := output.FindPreviousEM1(em1.ID, previous)
		ios.Println(output.FormatEM1Line(em1, prev))
	}
	ios.Println()
}

func displayPMStatus(ios *iostreams.IOStreams, statuses []shelly.PMStatus, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Power Meters:\n")
	for i := range statuses {
		pm := &statuses[i]
		prev := output.FindPreviousPM(pm.ID, previous)
		ios.Println(output.FormatPMLine(pm, prev))
	}
}

