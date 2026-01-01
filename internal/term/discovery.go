package term

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DisplayDiscoveredDevices prints a table of discovered devices.
// Handles empty case with ios.NoResults and prints count at the end.
func DisplayDiscoveredDevices(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	builder := output.FormatDiscoveredDevices(devices)
	if builder == nil {
		ios.NoResults("devices")
		return
	}
	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print discovered devices table", err)
	}
	ios.Println("")
	ios.Count("device", len(devices))
}

// DisplayBLEDevices prints a table of BLE discovered devices.
func DisplayBLEDevices(ios *iostreams.IOStreams, devices []discovery.BLEDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	builder := table.NewBuilder("Name", "Address", "Model", "RSSI", "Connectable", "BTHome")

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

		builder.AddRow(
			name,
			d.Address.String(),
			d.Model,
			rssiStr,
			connStr,
			btHomeStr,
		)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print BLE devices table", err)
	}
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

// DisplayPluginDiscoveredDevices prints a list of plugin-discovered devices.
func DisplayPluginDiscoveredDevices(ios *iostreams.IOStreams, devices []PluginDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	ios.Println()
	ios.Title("Plugin-managed Devices")

	for _, d := range devices {
		name := d.Name
		if name == "" {
			name = d.ID
		}
		if name == "" && d.Address != "" {
			name = d.Address
		}

		ios.Printf("  %s\n", name)
		if d.Address != "" {
			ios.Printf("    Address:  %s\n", d.Address)
		}
		ios.Printf("    Platform: %s\n", d.Platform)
		if d.Model != "" {
			ios.Printf("    Model:    %s\n", d.Model)
		}
		if d.Firmware != "" {
			ios.Printf("    Firmware: %s\n", d.Firmware)
		}
		if len(d.Components) > 0 {
			ios.Printf("    Components:\n")
			for _, c := range d.Components {
				if c.Name != "" {
					ios.Printf("      - %s:%d (%s)\n", c.Type, c.ID, c.Name)
				} else {
					ios.Printf("      - %s:%d\n", c.Type, c.ID)
				}
			}
		}
		ios.Println()
	}

	ios.Count(devices[0].Platform+" device", len(devices))
}

// PluginDiscoveredDevice represents a device discovered by a plugin.
// This is a display-oriented type used by term functions.
type PluginDiscoveredDevice struct {
	ID         string
	Name       string
	Model      string
	Address    string
	Platform   string
	Firmware   string
	Components []PluginComponentInfo
}

// PluginComponentInfo represents component info from plugin detection.
type PluginComponentInfo struct {
	Type string
	ID   int
	Name string
}

// ConvertPluginDevice converts a shelly.PluginDiscoveredDevice to a display-oriented type.
func ConvertPluginDevice(d PluginDeviceAdapter) PluginDiscoveredDevice {
	components := make([]PluginComponentInfo, len(d.Components))
	for i, c := range d.Components {
		components[i] = PluginComponentInfo(c)
	}
	return PluginDiscoveredDevice{
		ID:         d.ID,
		Name:       d.Name,
		Model:      d.Model,
		Address:    d.Address,
		Platform:   d.Platform,
		Firmware:   d.Firmware,
		Components: components,
	}
}

// PluginDeviceAdapter is the interface for converting plugin devices to display format.
type PluginDeviceAdapter struct {
	ID         string
	Name       string
	Model      string
	Address    string
	Platform   string
	Firmware   string
	Components []ComponentAdapter
}

// ComponentAdapter holds component info for conversion.
type ComponentAdapter struct {
	Type string
	ID   int
	Name string
}

// ConvertPluginDevices converts shelly plugin devices to display-oriented types.
func ConvertPluginDevices(devices []shelly.PluginDiscoveredDevice) []PluginDiscoveredDevice {
	result := make([]PluginDiscoveredDevice, len(devices))
	for i, d := range devices {
		components := make([]PluginComponentInfo, len(d.Components))
		for j, c := range d.Components {
			components[j] = PluginComponentInfo{
				Type: c.Type,
				ID:   c.ID,
				Name: c.Name,
			}
		}
		result[i] = PluginDiscoveredDevice{
			ID:         d.ID,
			Name:       d.Name,
			Model:      d.Model,
			Address:    d.Address.String(),
			Platform:   d.Platform,
			Firmware:   d.Firmware,
			Components: components,
		}
	}
	return result
}

// RegisterPluginDevices converts display-oriented plugin devices to utils format and registers them.
// Returns the number of devices successfully registered.
func RegisterPluginDevices(devices []PluginDiscoveredDevice, skipExisting bool) int {
	pluginDevices := make([]utils.PluginDevice, len(devices))
	for i, d := range devices {
		pluginDevices[i] = utils.PluginDevice{
			Address:  d.Address,
			Platform: d.Platform,
			ID:       d.ID,
			Name:     d.Name,
			Model:    d.Model,
			Firmware: d.Firmware,
		}
	}
	return utils.RegisterPluginDiscoveredDevices(pluginDevices, skipExisting)
}
