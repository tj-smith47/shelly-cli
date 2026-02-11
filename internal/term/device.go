package term

import (
	"fmt"
	"sort"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayDeviceStatus prints the device status information.
func DisplayDeviceStatus(ios *iostreams.IOStreams, status *shelly.DeviceStatus) {
	ios.Info("Device: %s", theme.Bold().Render(status.Info.ID))
	ios.Info("Model: %s (Gen%d)", status.Info.Model, status.Info.Generation)
	ios.Info("Firmware: %s", status.Info.Firmware)
	ios.Println()

	builder := table.NewBuilder("Component", "Status")
	for key, value := range status.Status {
		if m, ok := value.(map[string]any); ok {
			builder.AddRow(key, output.FormatComponentStatus(key, m))
		} else {
			builder.AddRow(key, output.FormatDisplayValue(value))
		}
	}
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device status table", err)
	}
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

// DisplayPluginDeviceStatus prints status for a plugin-managed device.
func DisplayPluginDeviceStatus(ios *iostreams.IOStreams, device model.Device, status *plugins.DeviceStatusResult) {
	// Header with device info
	ios.Info("Device: %s", theme.Bold().Render(device.DisplayName()))
	ios.Info("Platform: %s", device.GetPlatform())
	if device.Model != "" {
		ios.Info("Model: %s", device.Model)
	}
	ios.Info("Address: %s", device.Address)
	ios.Println()

	// Online status
	onlineStr := output.RenderOnline(status.Online, output.CaseTitle)
	ios.Printf("  Status: %s\n", onlineStr)
	ios.Println()

	// Components
	if len(status.Components) > 0 {
		ios.Printf("%s\n", theme.Bold().Render("Components"))

		builder := table.NewBuilder("Component", "Status")

		// Sort component keys for consistent output
		keys := make([]string, 0, len(status.Components))
		for k := range status.Components {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := status.Components[key]
			if m, ok := value.(map[string]any); ok {
				builder.AddRow(key, output.FormatComponentStatus(key, m))
			} else {
				builder.AddRow(key, output.FormatDisplayValue(value))
			}
		}
		tbl := builder.WithModeStyle(ios).Build()
		if err := tbl.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print plugin components table", err)
		}
		ios.Println()
	}

	// Sensors
	if len(status.Sensors) > 0 {
		ios.Printf("%s\n", theme.Bold().Render("Sensors"))

		builder := table.NewBuilder("Sensor", "Value")

		// Sort sensor keys for consistent output
		keys := make([]string, 0, len(status.Sensors))
		for k := range status.Sensors {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := status.Sensors[key]
			if m, ok := value.(map[string]any); ok {
				builder.AddRow(key, output.FormatComponentStatus(key, m))
			} else {
				builder.AddRow(key, output.FormatDisplayValue(value))
			}
		}
		tbl := builder.WithModeStyle(ios).Build()
		if err := tbl.PrintTo(ios.Out); err != nil {
			ios.DebugErr("print plugin sensors table", err)
		}
		ios.Println()
	}

	// Energy metrics
	if status.Energy != nil {
		ios.Printf("%s\n", theme.Bold().Render("Energy"))
		displayPluginEnergy(ios, status.Energy)
	}
}

// displayPluginEnergy prints energy metrics from a plugin status.
func displayPluginEnergy(ios *iostreams.IOStreams, energy *plugins.EnergyStatus) {
	if energy.Power > 0 {
		ios.Printf("  Power:   %s\n", output.FormatPowerColored(energy.Power))
	}
	if energy.Voltage > 0 {
		ios.Printf("  Voltage: %.1f V\n", energy.Voltage)
	}
	if energy.Current > 0 {
		ios.Printf("  Current: %.3f A\n", energy.Current)
	}
	if energy.PowerFactor > 0 {
		ios.Printf("  PF:      %.2f\n", energy.PowerFactor)
	}
	if energy.Total > 0 {
		ios.Printf("  Total:   %.3f kWh\n", energy.Total)
	}
}

// DisplayDeviceInfo renders device info as a formatted table.
func DisplayDeviceInfo(ios *iostreams.IOStreams, info *shelly.DeviceInfo) {
	builder := table.NewBuilder("Property", "Value")

	builder.AddRow("ID", info.ID)
	builder.AddRow("MAC", model.NormalizeMAC(info.MAC))
	builder.AddRow("Model", info.Model)
	builder.AddRow("Generation", output.RenderGeneration(info.Generation))
	builder.AddRow("Firmware", info.Firmware)
	builder.AddRow("Application", info.App)
	builder.AddRow("Auth Enabled", output.RenderAuthRequired(info.AuthEn))
	if info.Address != "" {
		builder.AddRow("Address", info.Address)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print device info table", err)
	}
}
