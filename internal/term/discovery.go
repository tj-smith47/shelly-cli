package term

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayDiscoveredDevices prints a table of discovered devices.
// Handles empty case with ios.NoResults and prints count at the end.
func DisplayDiscoveredDevices(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	table := output.FormatDiscoveredDevices(devices)
	if table == nil {
		ios.NoResults("devices")
		return
	}
	printTable(ios, table)
	ios.Count("device", len(devices))
}

// DisplayBLEDevices prints a table of BLE discovered devices.
func DisplayBLEDevices(ios *iostreams.IOStreams, devices []discovery.BLEDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	table := output.NewTable("Name", "Address", "Model", "RSSI", "Connectable", "BTHome")

	for _, d := range devices {
		name := d.LocalName
		if name == "" {
			name = d.ID
		}

		// Theme RSSI value (stronger is better)
		rssiStr := fmt.Sprintf("%d dBm", d.RSSI)
		switch {
		case d.RSSI > -50:
			rssiStr = theme.StatusOK().Render(rssiStr)
		case d.RSSI > -70:
			rssiStr = theme.StatusWarn().Render(rssiStr)
		default:
			rssiStr = theme.StatusError().Render(rssiStr)
		}

		// Connectable status
		connStr := output.RenderYesNo(d.Connectable, output.CaseTitle, theme.FalseError)

		// BTHome data indicator
		btHomeStr := "-"
		if d.BTHomeData != nil {
			btHomeStr = theme.StatusInfo().Render("Yes")
		}

		table.AddRow(
			name,
			d.Address.String(),
			d.Model,
			rssiStr,
			connStr,
			btHomeStr,
		)
	}

	printTable(ios, table)
	ios.Println("")
	ios.Count("BLE device", len(devices))
}

// DisplayGen1Details shows detailed Gen1-specific information.
func DisplayGen1Details(ctx context.Context, ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Found %d device(s):", len(devices))))
	ios.Println()

	for _, d := range devices {
		DisplaySingleGen1Device(ctx, ios, d)
	}
}

// DisplaySingleGen1Device displays details for a single Gen1 device.
func DisplaySingleGen1Device(ctx context.Context, ios *iostreams.IOStreams, d discovery.DiscoveredDevice) {
	ios.Printf("  %s\n", theme.Highlight().Render(d.Name))
	ios.Printf("    Address: %s\n", d.Address)
	ios.Printf("    Model:   %s\n", d.Model)

	// Try to get Gen1-specific details
	result, err := client.DetectGeneration(ctx, d.Address.String(), nil)
	if err != nil {
		ios.Printf("    Gen:     %s\n", theme.Dim().Render("unknown"))
		ios.Println()
		return
	}

	if !result.IsGen1() {
		ios.Printf("    Gen:     %s\n", theme.Dim().Render(fmt.Sprintf("Gen%d", result.Generation)))
		ios.Println()
		return
	}

	displayGen1Info(ctx, ios, d, result)
	ios.Println()
}

// displayGen1Info displays Gen1-specific device information.
func displayGen1Info(ctx context.Context, ios *iostreams.IOStreams, d discovery.DiscoveredDevice, result *client.DetectionResult) {
	ios.Printf("    Gen:     %s\n", theme.StatusOK().Render("Gen1"))
	ios.Printf("    Type:    %s\n", result.DeviceType)
	ios.Printf("    FW:      %s\n", result.Firmware)
	if result.AuthEn {
		ios.Printf("    Auth:    %s\n", theme.StatusWarn().Render("enabled"))
	}

	// Try to get full Gen1 status for more details
	gen1Device := model.Device{Address: d.Address.String()}
	gen1Client, err := client.ConnectGen1(ctx, gen1Device)
	if err != nil {
		return
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	if _, err := gen1Client.GetStatus(ctx); err == nil {
		ios.Printf("    Status:  %s\n", theme.StatusOK().Render("available"))
	}
}
