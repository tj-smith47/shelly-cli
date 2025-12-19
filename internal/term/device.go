package term

import (
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayDeviceStatus prints the device status information.
func DisplayDeviceStatus(ios *iostreams.IOStreams, status *shelly.DeviceStatus) {
	ios.Info("Device: %s", theme.Bold().Render(status.Info.ID))
	ios.Info("Model: %s (Gen%d)", status.Info.Model, status.Info.Generation)
	ios.Info("Firmware: %s", status.Info.Firmware)
	ios.Println()

	table := output.NewTable("Component", "Value")
	for key, value := range status.Status {
		table.AddRow(key, output.FormatDisplayValue(value))
	}
	printTable(ios, table)
}

// DisplayAllSnapshots displays a summary of all device snapshots.
// Requires a mutex lock to be held by the caller when accessing the snapshots map.
func DisplayAllSnapshots(ios *iostreams.IOStreams, snapshots map[string]*shelly.DeviceSnapshot) {
	// Clear screen
	ios.ClearScreen()

	ios.Title("Device Status Summary")
	ios.Printf("  Updated: %s\n\n", time.Now().Format("15:04:05"))

	totalPower := 0.0
	totalEnergy := 0.0
	onlineCount := 0
	offlineCount := 0

	// Display each device
	for name, snap := range snapshots {
		if snap.Error != nil {
			ios.Printf("%s %s: %s\n",
				theme.StatusError().Render("●"),
				name,
				theme.Dim().Render(snap.Error.Error()))
			offlineCount++
			continue
		}

		onlineCount++

		// Calculate device power using shared helper
		devicePower, deviceEnergy := output.CalculateSnapshotTotals(snap.Snapshot)
		totalPower += devicePower
		totalEnergy += deviceEnergy

		// Display device line
		statusIcon := theme.StatusOK().Render("●")
		deviceModel := "Unknown"
		if snap.Info != nil {
			deviceModel = snap.Info.Model
		}

		powerStr := output.FormatPowerColored(devicePower)
		energyStr := ""
		if deviceEnergy > 0 {
			energyStr = fmt.Sprintf("  %.2f Wh", deviceEnergy)
		}
		ios.Printf("%s %s (%s): %s%s\n",
			statusIcon, name, deviceModel, powerStr, energyStr)
	}

	// Display summary
	ios.Println()
	ios.Printf("═══════════════════════════════════════\n")
	ios.Printf("  Online:       %s / %d devices\n",
		theme.StatusOK().Render(fmt.Sprintf("%d", onlineCount)),
		onlineCount+offlineCount)
	ios.Printf("  Total Power:  %s\n", theme.StatusOK().Render(output.FormatPower(totalPower)))
	if totalEnergy > 0 {
		ios.Printf("  Total Energy: %.2f Wh\n", totalEnergy)
	}
}

// DisplayAuthStatus prints the authentication status.
func DisplayAuthStatus(ios *iostreams.IOStreams, status *shelly.AuthStatus) {
	ios.Title("Authentication Status")
	ios.Println()

	ios.Printf("  Status: %s\n", output.RenderBoolState(status.Enabled, "Enabled", "Disabled"))
}
