package term

import (
	"fmt"
	"time"

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

// DisplayBTHomeComponentStatus displays BTHome component status.
func DisplayBTHomeComponentStatus(ios *iostreams.IOStreams, status model.BTHomeComponentStatus) {
	ios.Println(theme.Bold().Render("BTHome Status:"))
	ios.Println()

	if status.Discovery != nil {
		ios.Println("  " + theme.Highlight().Render("Discovery:"))
		startTime := time.Unix(int64(status.Discovery.StartedAt), 0)
		ios.Printf("    Started: %s\n", startTime.Format(time.RFC3339))
		ios.Printf("    Duration: %ds\n", status.Discovery.Duration)
		ios.Println()
	} else {
		ios.Info("No active discovery scan.")
	}

	displayBTHomeErrors(ios, status.Errors)
}

// DisplayBTHomeDeviceStatus displays detailed BTHome device status.
func DisplayBTHomeDeviceStatus(ios *iostreams.IOStreams, status model.BTHomeDeviceStatus) {
	name := status.Name
	if name == "" {
		name = fmt.Sprintf("Device %d", status.ID)
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("BTHome Device: %s", name)))
	ios.Println()

	displayBTHomeDeviceBasicInfo(ios, status)
	displayBTHomeKnownObjects(ios, status.KnownObjects)
	displayBTHomeErrors(ios, status.Errors)
}

func displayBTHomeDeviceBasicInfo(ios *iostreams.IOStreams, status model.BTHomeDeviceStatus) {
	ios.Printf("  ID: %d\n", status.ID)
	if status.Addr != "" {
		ios.Printf("  Address: %s\n", status.Addr)
	}
	if status.RSSI != nil {
		ios.Printf("  RSSI: %d dBm\n", *status.RSSI)
	}
	if status.Battery != nil {
		ios.Printf("  Battery: %d%%\n", *status.Battery)
	}
	if status.PacketID != nil {
		ios.Printf("  Packet ID: %d\n", *status.PacketID)
	}
	if status.LastUpdateTS > 0 {
		lastUpdate := time.Unix(int64(status.LastUpdateTS), 0)
		ios.Printf("  Last Update: %s\n", lastUpdate.Format(time.RFC3339))
	}
}

func displayBTHomeKnownObjects(ios *iostreams.IOStreams, objects []model.BTHomeKnownObj) {
	if len(objects) == 0 {
		return
	}

	ios.Println()
	ios.Println("  " + theme.Highlight().Render("Known Objects:"))
	for _, obj := range objects {
		managed := ""
		if obj.Component != nil {
			managed = fmt.Sprintf(" (managed by %s)", *obj.Component)
		}
		ios.Printf("    Object ID: %d, Index: %d%s\n", obj.ObjID, obj.Idx, managed)
	}
}

func displayBTHomeErrors(ios *iostreams.IOStreams, errors []string) {
	if len(errors) == 0 {
		return
	}

	ios.Println()
	ios.Error("Errors:")
	for _, e := range errors {
		ios.Printf("  - %s\n", e)
	}
}
