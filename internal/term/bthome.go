package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// BTHomeAddResult holds the result of adding a BTHome device.
type BTHomeAddResult struct {
	Key  string
	Name string
	Addr string
}

// DisplayBTHomeAddResult displays the result of adding a BTHome device.
func DisplayBTHomeAddResult(ios *iostreams.IOStreams, result BTHomeAddResult) {
	ios.Success("BTHome device added: %s", result.Key)
	if result.Name != "" {
		ios.Info("Name: %s", result.Name)
	}
	ios.Info("Address: %s", result.Addr)
}

// DisplayBTHomeDiscoveryStarted displays instructions after starting discovery.
func DisplayBTHomeDiscoveryStarted(ios *iostreams.IOStreams, device string, duration int) {
	ios.Println(theme.Bold().Render("BTHome Device Discovery Started"))
	ios.Println()
	ios.Info("Scanning for %d seconds...", duration)
	ios.Println()
	ios.Info("Discovered devices will emit 'device_discovered' events.")
	ios.Info("Monitor events with: shelly monitor events %s", device)
	ios.Println()
	ios.Info("When discovery completes, a 'discovery_done' event will be emitted.")
	ios.Info("Then use 'shelly bthome add %s --addr <mac>' to add discovered devices.", device)
}

// DisplayBTHomeDevices displays a list of BTHome devices.
func DisplayBTHomeDevices(ios *iostreams.IOStreams, devices []model.BTHomeDeviceInfo, gatewayDevice string) {
	if len(devices) == 0 {
		ios.Info("No BTHome devices found.")
		ios.Info("Use 'shelly bthome add %s' to discover new devices.", gatewayDevice)
		return
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("BTHome Devices (%d):", len(devices))))
	ios.Println()

	for _, dev := range devices {
		displayBTHomeDevice(ios, dev)
	}
}

func displayBTHomeDevice(ios *iostreams.IOStreams, dev model.BTHomeDeviceInfo) {
	name := dev.Name
	if name == "" {
		name = fmt.Sprintf("Device %d", dev.ID)
	}

	ios.Printf("  %s\n", theme.Highlight().Render(name))
	ios.Printf("    ID: %d\n", dev.ID)
	if dev.Addr != "" {
		ios.Printf("    Address: %s\n", dev.Addr)
	}
	if dev.RSSI != nil {
		ios.Printf("    RSSI: %d dBm\n", *dev.RSSI)
	}
	if dev.Battery != nil {
		ios.Printf("    Battery: %d%%\n", *dev.Battery)
	}
	ios.Println()
}
