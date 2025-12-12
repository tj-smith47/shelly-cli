// Package status provides the monitor status subcommand for real-time device monitoring.
package status

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var (
	intervalFlag time.Duration
	countFlag    int
)

// NewCommand creates the monitor status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <device>",
		Short: "Monitor device status in real-time",
		Long: `Monitor a device's status in real-time with automatic refresh.

Status includes switches, covers, lights, energy meters, and other components.
Press Ctrl+C to stop monitoring.`,
		Example: `  # Monitor device status every 2 seconds
  shelly monitor status living-room

  # Monitor with custom interval
  shelly monitor status living-room --interval 5s

  # Monitor for a specific number of updates
  shelly monitor status living-room --count 10`,
		Args: cobra.ExactArgs(1),
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
		clearScreen(ios)
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
		displaySingleEM(ios, em, findPreviousEM(em.ID, previous))
	}
	ios.Println()
}

func displaySingleEM(ios *iostreams.IOStreams, em, prev *shelly.EMStatus) {
	ios.Printf("  EM %d:\n", em.ID)

	// Phase A
	powerA := formatPower(em.AActivePower)
	if prev != nil && em.AActivePower != prev.AActivePower {
		powerA = theme.StatusWarn().Render(powerA + " ↑")
	}
	ios.Printf("    Phase A: %s  %.1fV  %.2fA\n", powerA, em.AVoltage, em.ACurrent)

	// Phase B
	powerB := formatPower(em.BActivePower)
	if prev != nil && em.BActivePower != prev.BActivePower {
		powerB = theme.StatusWarn().Render(powerB + " ↑")
	}
	ios.Printf("    Phase B: %s  %.1fV  %.2fA\n", powerB, em.BVoltage, em.BCurrent)

	// Phase C
	powerC := formatPower(em.CActivePower)
	if prev != nil && em.CActivePower != prev.CActivePower {
		powerC = theme.StatusWarn().Render(powerC + " ↑")
	}
	ios.Printf("    Phase C: %s  %.1fV  %.2fA\n", powerC, em.CVoltage, em.CCurrent)

	// Totals
	ios.Printf("    Total:   %.1f W\n", em.TotalActivePower)
}

func displayEM1Status(ios *iostreams.IOStreams, statuses []shelly.EM1Status, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Energy Meters (single-phase):\n")
	for i := range statuses {
		em1 := &statuses[i]
		displaySingleEM1(ios, em1, findPreviousEM1(em1.ID, previous))
	}
	ios.Println()
}

func displaySingleEM1(ios *iostreams.IOStreams, em1, prev *shelly.EM1Status) {
	power := formatPower(em1.ActPower)
	if prev != nil && em1.ActPower != prev.ActPower {
		power = theme.StatusWarn().Render(power + " ↑")
	}

	ios.Printf("  EM1 %d: %s  %.1fV  %.2fA\n",
		em1.ID, power, em1.Voltage, em1.Current)
}

func displayPMStatus(ios *iostreams.IOStreams, statuses []shelly.PMStatus, previous *shelly.MonitoringSnapshot) {
	if len(statuses) == 0 {
		return
	}

	ios.Printf("Power Meters:\n")
	for i := range statuses {
		pm := &statuses[i]
		displaySinglePM(ios, pm, findPreviousPM(pm.ID, previous))
	}
}

func displaySinglePM(ios *iostreams.IOStreams, pm, prev *shelly.PMStatus) {
	power := formatPower(pm.APower)
	if prev != nil && pm.APower != prev.APower {
		power = theme.StatusWarn().Render(power + " ↑")
	}

	energyStr := ""
	if pm.AEnergy != nil {
		energyStr = fmt.Sprintf("  %.2f Wh", pm.AEnergy.Total)
	}

	ios.Printf("  PM %d: %s  %.1fV  %.2fA%s\n",
		pm.ID, power, pm.Voltage, pm.Current, energyStr)
}

func formatPower(w float64) string {
	if w >= 1000 {
		return fmt.Sprintf("%.2f kW", w/1000)
	}
	return fmt.Sprintf("%.1f W", w)
}

func clearScreen(ios *iostreams.IOStreams) {
	ios.Printf("\033[H\033[2J")
}

func findPreviousEM(id int, prev *shelly.MonitoringSnapshot) *shelly.EMStatus {
	if prev == nil {
		return nil
	}
	for i := range prev.EM {
		if prev.EM[i].ID == id {
			return &prev.EM[i]
		}
	}
	return nil
}

func findPreviousEM1(id int, prev *shelly.MonitoringSnapshot) *shelly.EM1Status {
	if prev == nil {
		return nil
	}
	for i := range prev.EM1 {
		if prev.EM1[i].ID == id {
			return &prev.EM1[i]
		}
	}
	return nil
}

func findPreviousPM(id int, prev *shelly.MonitoringSnapshot) *shelly.PMStatus {
	if prev == nil {
		return nil
	}
	for i := range prev.PM {
		if prev.PM[i].ID == id {
			return &prev.PM[i]
		}
	}
	return nil
}
