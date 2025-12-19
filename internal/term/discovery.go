package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
