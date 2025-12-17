// Package power provides the monitor power subcommand for real-time power monitoring.
package power

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
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
		displayPowerSnapshot(ios, &snapshot, lastSnapshot)
		lastSnapshot = &snapshot
		return nil
	})
}

func displayPowerSnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Power Consumption")
	ios.Printf("  Updated: %s\n\n", current.Timestamp.Format("15:04:05"))

	// Calculate totals
	totalPower := 0.0
	totalEnergy := 0.0

	// Display 3-phase energy meters
	for i := range current.EM {
		em := &current.EM[i]
		ios.Printf("EM %d (3-phase):\n", em.ID)
		displayPhase(ios, "A", em.AActivePower, em.AVoltage, em.ACurrent, em.APowerFactor, getPrevEMPowerA(em.ID, previous))
		displayPhase(ios, "B", em.BActivePower, em.BVoltage, em.BCurrent, em.BPowerFactor, getPrevEMPowerB(em.ID, previous))
		displayPhase(ios, "C", em.CActivePower, em.CVoltage, em.CCurrent, em.CPowerFactor, getPrevEMPowerC(em.ID, previous))
		ios.Printf("  Total: %s\n\n", output.FormatPowerColored(em.TotalActivePower))
		totalPower += em.TotalActivePower
	}

	// Display single-phase energy meters
	for i := range current.EM1 {
		em1 := &current.EM1[i]
		ios.Printf("EM1 %d:\n", em1.ID)
		prevPower := getPrevEM1Power(em1.ID, previous)
		displayMeter(ios, em1.ActPower, em1.Voltage, em1.Current, em1.PF, prevPower)
		totalPower += em1.ActPower
	}

	// Display power meters
	for i := range current.PM {
		pm := &current.PM[i]
		ios.Printf("PM %d:\n", pm.ID)
		prevPower := getPrevPMPower(pm.ID, previous)
		displayMeter(ios, pm.APower, pm.Voltage, pm.Current, nil, prevPower)
		totalPower += pm.APower
		if pm.AEnergy != nil {
			totalEnergy += pm.AEnergy.Total
		}
	}

	// Display totals
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(output.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

func displayPhase(ios *iostreams.IOStreams, phase string, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerColored(power)
	if prevPower != nil && power != *prevPower {
		if power > *prevPower {
			powerStr = theme.StatusWarn().Render(output.FormatPower(power) + " ↑")
		} else {
			powerStr = theme.StatusOK().Render(output.FormatPower(power) + " ↓")
		}
	}
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Phase %s: %s  %.1fV  %.2fA%s\n", phase, powerStr, voltage, current, pfStr)
}

func displayMeter(ios *iostreams.IOStreams, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerColored(power)
	if prevPower != nil && power != *prevPower {
		if power > *prevPower {
			powerStr = theme.StatusWarn().Render(output.FormatPower(power) + " ↑")
		} else {
			powerStr = theme.StatusOK().Render(output.FormatPower(power) + " ↓")
		}
	}
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Power: %s  Voltage: %.1fV  Current: %.2fA%s\n", powerStr, voltage, current, pfStr)
}

func getPrevEMPowerA(id int, prev *shelly.MonitoringSnapshot) *float64 {
	if prev == nil {
		return nil
	}
	for i := range prev.EM {
		if prev.EM[i].ID == id {
			return &prev.EM[i].AActivePower
		}
	}
	return nil
}

func getPrevEMPowerB(id int, prev *shelly.MonitoringSnapshot) *float64 {
	if prev == nil {
		return nil
	}
	for i := range prev.EM {
		if prev.EM[i].ID == id {
			return &prev.EM[i].BActivePower
		}
	}
	return nil
}

func getPrevEMPowerC(id int, prev *shelly.MonitoringSnapshot) *float64 {
	if prev == nil {
		return nil
	}
	for i := range prev.EM {
		if prev.EM[i].ID == id {
			return &prev.EM[i].CActivePower
		}
	}
	return nil
}

func getPrevEM1Power(id int, prev *shelly.MonitoringSnapshot) *float64 {
	if prev == nil {
		return nil
	}
	for i := range prev.EM1 {
		if prev.EM1[i].ID == id {
			return &prev.EM1[i].ActPower
		}
	}
	return nil
}

func getPrevPMPower(id int, prev *shelly.MonitoringSnapshot) *float64 {
	if prev == nil {
		return nil
	}
	for i := range prev.PM {
		if prev.PM[i].ID == id {
			return &prev.PM[i].APower
		}
	}
	return nil
}
