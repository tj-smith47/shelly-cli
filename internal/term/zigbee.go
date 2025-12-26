// Package term provides composed terminal presentation for the CLI.
package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayZigbeeDevices displays a list of Zigbee-capable devices.
func DisplayZigbeeDevices(ios *iostreams.IOStreams, devices []model.ZigbeeDevice) {
	if len(devices) == 0 {
		ios.Info("No Zigbee-capable devices found.")
		ios.Info("Zigbee is supported on Gen4 devices.")
		return
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Zigbee-Capable Devices (%d):", len(devices))))
	ios.Println()

	for _, dev := range devices {
		displayZigbeeDevice(ios, dev)
	}
}

func displayZigbeeDevice(ios *iostreams.IOStreams, dev model.ZigbeeDevice) {
	ios.Printf("  %s\n", theme.Highlight().Render(dev.Name))
	ios.Printf("    Address: %s\n", dev.Address)
	if dev.Model != "" {
		ios.Printf("    Model: %s\n", dev.Model)
	}

	ios.Printf("    Zigbee: %s\n", output.RenderEnabledState(dev.Enabled))

	if dev.NetworkState != "" {
		ios.Printf("    State: %s\n", output.RenderNetworkState(dev.NetworkState))
	}
	if dev.EUI64 != "" {
		ios.Printf("    EUI64: %s\n", dev.EUI64)
	}
	ios.Println()
}

// OutputZigbeeDevicesJSON outputs Zigbee devices as JSON.
func OutputZigbeeDevicesJSON(ios *iostreams.IOStreams, devices []model.ZigbeeDevice) error {
	jsonBytes, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(jsonBytes))
	return nil
}

// DisplayZigbeeStatus displays the Zigbee status for a device.
func DisplayZigbeeStatus(ios *iostreams.IOStreams, status model.ZigbeeStatus) {
	ios.Println(theme.Bold().Render("Zigbee Status:"))
	ios.Println()

	ios.Printf("  Enabled: %s\n", output.RenderEnabledState(status.Enabled))
	displayZigbeeNetworkState(ios, status.NetworkState)

	if status.EUI64 != "" {
		ios.Printf("  EUI64: %s\n", status.EUI64)
	}

	if status.NetworkState == "joined" {
		displayZigbeeNetworkInfo(ios, status)
	}
}

func displayZigbeeNetworkState(ios *iostreams.IOStreams, state string) {
	if state == "" {
		return
	}

	stateStyle := theme.Dim()
	switch state {
	case "joined":
		stateStyle = theme.StatusOK()
	case "steering":
		stateStyle = theme.StatusWarn()
	}
	ios.Printf("  Network State: %s\n", stateStyle.Render(state))
}

func displayZigbeeNetworkInfo(ios *iostreams.IOStreams, status model.ZigbeeStatus) {
	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Network Info:"))
	ios.Printf("    PAN ID: 0x%04X\n", status.PANID)
	ios.Printf("    Channel: %d\n", status.Channel)
	if status.CoordinatorEUI64 != "" {
		ios.Printf("    Coordinator: %s\n", status.CoordinatorEUI64)
	}
}

// OutputZigbeeStatusJSON outputs Zigbee status as JSON.
func OutputZigbeeStatusJSON(ios *iostreams.IOStreams, status model.ZigbeeStatus) error {
	jsonBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}
	ios.Println(string(jsonBytes))
	return nil
}
