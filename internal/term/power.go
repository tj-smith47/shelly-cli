package term

import (
	"fmt"
	"sort"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayPowerMetrics outputs power, voltage, and current with units.
// Nil values are skipped.
func DisplayPowerMetrics(ios *iostreams.IOStreams, power, voltage, current *float64) {
	displayPowerMetricsWithWidth(ios, power, voltage, current, 9)
}

// DisplayPowerMetricsWide outputs power metrics with wider alignment for cover status.
func DisplayPowerMetricsWide(ios *iostreams.IOStreams, power, voltage, current *float64) {
	displayPowerMetricsWithWidth(ios, power, voltage, current, 10)
}

func displayPowerMetricsWithWidth(ios *iostreams.IOStreams, power, voltage, current *float64, width int) {
	if power != nil {
		ios.Printf("  %-*s%.1f W\n", width, "Power:", *power)
	}
	if voltage != nil {
		ios.Printf("  %-*s%.1f V\n", width, "Voltage:", *voltage)
	}
	if current != nil {
		ios.Printf("  %-*s%.3f A\n", width, "Current:", *current)
	}
}

// DisplayPowerSnapshot displays a power monitoring snapshot with all energy meters.
// Handles 3-phase EM, single-phase EM1, and power meters with change indicators.
func DisplayPowerSnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Power Consumption")
	ios.Printf("  Updated: %s\n\n", current.Timestamp.Format("15:04:05"))

	// Display 3-phase energy meters
	for i := range current.EM {
		em := &current.EM[i]
		prevA, prevB, prevC := output.GetPrevEMPhasePower(em.ID, previous)
		ios.Printf("EM %d (3-phase):\n", em.ID)
		displayPhase(ios, "A", em.AActivePower, em.AVoltage, em.ACurrent, em.APowerFactor, prevA)
		displayPhase(ios, "B", em.BActivePower, em.BVoltage, em.BCurrent, em.BPowerFactor, prevB)
		displayPhase(ios, "C", em.CActivePower, em.CVoltage, em.CCurrent, em.CPowerFactor, prevC)
		ios.Printf("  Total: %s\n\n", output.FormatPowerColored(em.TotalActivePower))
	}

	// Display single-phase energy meters
	for i := range current.EM1 {
		em1 := &current.EM1[i]
		ios.Printf("EM1 %d:\n", em1.ID)
		prevPower := output.GetPrevEM1Power(em1.ID, previous)
		displayMeter(ios, em1.ActPower, em1.Voltage, em1.Current, em1.PF, prevPower)
	}

	// Display power meters
	for i := range current.PM {
		pm := &current.PM[i]
		ios.Printf("PM %d:\n", pm.ID)
		prevPower := output.GetPrevPMPower(pm.ID, previous)
		displayMeter(ios, pm.APower, pm.Voltage, pm.Current, nil, prevPower)
	}

	// Calculate and display totals
	totalPower, totalEnergy := output.CalculateSnapshotTotals(current)
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(output.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

// displayPhase prints a single phase line for power monitoring.
func displayPhase(ios *iostreams.IOStreams, phase string, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Phase %s: %s  %.1fV  %.2fA%s\n", phase, powerStr, voltage, current, pfStr)
}

// displayMeter prints a meter line for power monitoring.
func displayMeter(ios *iostreams.IOStreams, power, voltage, current float64, pf, prevPower *float64) {
	powerStr := output.FormatPowerWithChange(power, prevPower)
	pfStr := ""
	if pf != nil {
		pfStr = fmt.Sprintf("  PF:%.2f", *pf)
	}
	ios.Printf("  Power: %s  Voltage: %.1fV  Current: %.2fA%s\n", powerStr, voltage, current, pfStr)
}

// DisplayStatusSnapshot displays a device status monitoring snapshot.
// Shows energy meters in a simpler format than power monitoring.
func DisplayStatusSnapshot(ios *iostreams.IOStreams, current, previous *shelly.MonitoringSnapshot) {
	// Clear screen for non-first updates
	if previous != nil {
		ios.ClearScreen()
	}

	ios.Title("Device Status")
	ios.Printf("  Timestamp: %s\n\n", current.Timestamp.Format("2006-01-02T15:04:05Z07:00"))

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

// DisplayDashboard prints the energy dashboard with summary and device breakdown.
func DisplayDashboard(ios *iostreams.IOStreams, data model.DashboardData) {
	ios.Printf("%s\n", theme.Bold().Render("Energy Dashboard"))
	ios.Printf("Timestamp: %s\n\n", data.Timestamp.Format(time.RFC3339))

	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Devices:     %d total (%d online, %d offline)\n",
		data.DeviceCount, data.OnlineCount, data.OfflineCount)
	ios.Printf("  Total Power: %s\n", theme.StyledPower(data.TotalPower))

	if data.TotalEnergy > 0 {
		ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy))
	}

	if data.EstimatedCost != nil {
		ios.Printf("  Est. Cost:   %s %.2f/kWh = %s %.4f\n",
			data.CostCurrency, data.CostPerKwh,
			data.CostCurrency, *data.EstimatedCost)
	}

	ios.Printf("\n%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Device", "Status", "Power", "Components")
	for _, dev := range data.Devices {
		statusStr := theme.StatusOK().Render("online")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		}

		powerStr := output.FormatPower(dev.TotalPower)
		if !dev.Online {
			powerStr = "-"
		}

		table.AddRow(dev.Device, statusStr, powerStr, formatComponentSummary(dev.Components))
	}

	printTable(ios, table)
}

func formatComponentSummary(components []model.ComponentPower) string {
	if len(components) == 0 {
		return "-"
	}

	counts := make(map[string]int)
	for _, c := range components {
		counts[c.Type]++
	}

	parts := make([]string, 0, len(counts))
	for typ, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, typ))
	}

	return fmt.Sprintf("%d (%s)", len(components), joinStrings(parts, ", "))
}

// DisplayComparison prints energy comparison results with summary and bar chart.
func DisplayComparison(ios *iostreams.IOStreams, data model.ComparisonData) {
	ios.Printf("%s\n", theme.Bold().Render("Energy Comparison"))
	ios.Printf("Period: %s\n", data.Period)
	if !data.From.IsZero() {
		ios.Printf("From:   %s\n", data.From.Format("2006-01-02 15:04:05"))
	}
	if !data.To.IsZero() {
		ios.Printf("To:     %s\n", data.To.Format("2006-01-02 15:04:05"))
	}
	ios.Printf("\n")

	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy*1000))
	ios.Printf("  Max Device:   %s\n", theme.StyledEnergy(data.MaxEnergy*1000))
	ios.Printf("  Min Device:   %s\n", theme.StyledEnergy(data.MinEnergy*1000))
	ios.Printf("\n")

	// Sort by energy consumption (descending) for display
	sorted := sortDevicesByEnergy(data.Devices)

	ios.Printf("%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Rank", "Device", "Energy", "Avg Power", "Peak Power", "Share", "Status")
	for i, dev := range sorted {
		rank := fmt.Sprintf("#%d", i+1)

		statusStr := theme.StatusOK().Render("ok")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		} else if dev.Error != "" {
			statusStr = theme.StatusWarn().Render(output.Truncate(dev.Error, 15))
		}

		energyStr, avgStr, peakStr, shareStr := "-", "-", "-", "-"
		if dev.Online {
			energyStr = output.FormatEnergy(dev.Energy * 1000)
			avgStr = output.FormatPower(dev.AvgPower)
			peakStr = output.FormatPower(dev.PeakPower)
			if dev.Percentage > 0 {
				shareStr = fmt.Sprintf("%.1f%%", dev.Percentage)
			}
		}

		table.AddRow(rank, dev.Device, energyStr, avgStr, peakStr, shareStr, statusStr)
	}

	printTable(ios, table)

	ios.Printf("\n%s\n", theme.Bold().Render("Energy Distribution"))
	displayBarChart(ios, sorted, data.MaxEnergy)
}

func sortDevicesByEnergy(devices []model.DeviceEnergy) []model.DeviceEnergy {
	sorted := make([]model.DeviceEnergy, len(devices))
	copy(sorted, devices)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Energy > sorted[j].Energy
	})
	return sorted
}

func displayBarChart(ios *iostreams.IOStreams, devices []model.DeviceEnergy, maxEnergy float64) {
	if maxEnergy <= 0 {
		return
	}

	maxNameLen := 0
	for _, dev := range devices {
		if len(dev.Device) > maxNameLen {
			maxNameLen = len(dev.Device)
		}
	}

	for _, dev := range devices {
		if !dev.Online || dev.Energy <= 0 {
			continue
		}

		name := output.PadRight(dev.Device, maxNameLen)
		barLen := int((dev.Energy / maxEnergy) * 40)
		if barLen < 1 {
			barLen = 1
		}

		bar := repeatChar('█', barLen)
		ios.Printf("  %s │ %s %.2f kWh\n", name, theme.Highlight().Render(bar), dev.Energy)
	}
}

// DisplayPMStatusDetails shows detailed power meter status in human-readable format.
func DisplayPMStatusDetails(ios *iostreams.IOStreams, status *shelly.PMStatus, componentType string) {
	typeLabel := "Power Meter (PM)"
	if componentType == shelly.ComponentTypePM1 {
		typeLabel = "Power Meter (PM1)"
	}
	ios.Printf("%s #%d\n\n", typeLabel, status.ID)
	ios.Printf("Voltage: %.2f V\n", status.Voltage)
	ios.Printf("Current: %.2f A\n", status.Current)
	ios.Printf("Power:   %.2f W\n", status.APower)

	if status.Freq != nil {
		ios.Printf("Frequency: %.2f Hz\n", *status.Freq)
	}

	// Energy counters
	if status.AEnergy != nil {
		ios.Printf("\nAccumulated Energy:\n")
		ios.Printf("  Total: %.2f Wh\n", status.AEnergy.Total)
		if status.AEnergy.MinuteTs != nil && len(status.AEnergy.ByMinute) > 0 {
			ios.Printf("  Recent (by minute): %v\n", status.AEnergy.ByMinute[:min(5, len(status.AEnergy.ByMinute))])
		}
	}

	if status.RetAEnergy != nil {
		ios.Printf("\nReturn Energy:\n")
		ios.Printf("  Total: %.2f Wh\n", status.RetAEnergy.Total)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}
}
