// Package term provides terminal display functions for the CLI.
package term

import (
	"time"

	"github.com/tj-smith47/shelly-go/provisioning"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DisplayBLEProvisionResult displays the result of a BLE provisioning operation.
func DisplayBLEProvisionResult(ios *iostreams.IOStreams, result *provisioning.BLEProvisionResult, ssid string) {
	if result == nil {
		ios.Error("No provisioning result")
		return
	}

	if result.Success {
		ios.Success("BLE provisioning completed successfully")
	} else {
		ios.Error("BLE provisioning failed")
	}

	if result.Device != nil {
		ios.Printf("  Device: %s\n", result.Device.Name)
		ios.Printf("  Address: %s\n", result.Device.Address)
		if result.Device.Model != "" {
			ios.Printf("  Model: %s\n", result.Device.Model)
		}
	}

	if ssid != "" {
		ios.Printf("  WiFi SSID: %s\n", ssid)
	}

	ios.Printf("  Duration: %s\n", result.Duration().Round(time.Millisecond))

	if result.Error != nil {
		ios.Printf("  Error: %s\n", result.Error.Error())
	}

	if result.Success && ssid != "" {
		ios.Println()
		ios.Info("The device will now attempt to connect to the WiFi network.")
		ios.Info("You may need to wait a moment and then discover the device on your network.")
	}
}

// DisplayBLEDevice displays information about a discovered BLE device.
func DisplayBLEDevice(ios *iostreams.IOStreams, device *provisioning.BLEDevice) {
	if device == nil {
		return
	}

	ios.Printf("Device: %s\n", device.Name)
	ios.Printf("  Address: %s\n", device.Address)
	if device.Model != "" {
		ios.Printf("  Model: %s\n", device.Model)
	}
	if device.RSSI != 0 {
		ios.Printf("  Signal: %d dBm\n", device.RSSI)
	}
	if device.Generation > 0 {
		ios.Printf("  Generation: Gen%d\n", device.Generation)
	}
}
