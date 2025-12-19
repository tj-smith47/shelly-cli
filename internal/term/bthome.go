package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
