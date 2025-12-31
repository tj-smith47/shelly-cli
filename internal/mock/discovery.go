package mock

import (
	"net"
)

// DiscoveryResult represents a discovered device for demo mode.
// This mirrors discovery.DiscoveredDevice from shelly-go.
type DiscoveryResult struct {
	ID         string
	Name       string
	Model      string
	Address    net.IP
	MACAddress string
	Generation int
}

// GetDiscoveredDevices returns discovered devices from the current demo fixtures.
// Returns nil if not in demo mode or if fixtures aren't loaded.
func GetDiscoveredDevices() []DiscoveryResult {
	demo := GetCurrentDemo()
	if demo == nil || demo.Fixtures == nil {
		return nil
	}

	devices := make([]DiscoveryResult, len(demo.Fixtures.Discovery))
	for i, d := range demo.Fixtures.Discovery {
		devices[i] = DiscoveryResult{
			ID:         d.Name, // Use name as ID if not specified
			Name:       d.Name,
			Model:      d.Model,
			Address:    net.ParseIP(d.Address),
			MACAddress: d.MAC,
			Generation: d.Generation,
		}
	}
	return devices
}

// HasDiscoveryFixtures returns true if demo mode has discovery fixtures defined.
func HasDiscoveryFixtures() bool {
	demo := GetCurrentDemo()
	return demo != nil && demo.Fixtures != nil && len(demo.Fixtures.Discovery) > 0
}

// HasFleetFixtures returns true if demo mode has fleet fixtures defined.
func HasFleetFixtures() bool {
	demo := GetCurrentDemo()
	return demo != nil && demo.Fixtures != nil && len(demo.Fixtures.Fleet.Devices) > 0
}

// ToLibraryDiscoveredDevices converts mock discovery results to the shelly-go discovery type.
// This allows the mock results to be displayed using the standard term.DisplayDiscoveredDevices.
func ToLibraryDiscoveredDevices(results []DiscoveryResult) []LibraryDiscoveredDevice {
	devices := make([]LibraryDiscoveredDevice, len(results))
	for i, d := range results {
		devices[i] = LibraryDiscoveredDevice(d)
	}
	return devices
}

// LibraryDiscoveredDevice is a copy of discovery.DiscoveredDevice fields needed for display.
// This avoids importing the discovery package in mock.
type LibraryDiscoveredDevice struct {
	ID         string
	Name       string
	Model      string
	Address    net.IP
	MACAddress string
	Generation int
}
