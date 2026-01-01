// Package mock provides demo discovery functionality.
package mock

import (
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// RunDemoDiscovery returns mock discovery results from fixtures.
// This is used by the discover command when demo mode is active.
func RunDemoDiscovery(ios *iostreams.IOStreams, register, skipExisting bool) error {
	// Get mock discovered devices
	mockDevices := GetDiscoveredDevices()

	// Convert to discovery.DiscoveredDevice for display compatibility
	shellyDevices := make([]discovery.DiscoveredDevice, len(mockDevices))
	for i, d := range mockDevices {
		shellyDevices[i] = discovery.DiscoveredDevice{
			ID:         d.ID,
			Name:       d.Name,
			Model:      d.Model,
			Address:    d.Address,
			MACAddress: d.MACAddress,
			Generation: types.Generation(d.Generation),
			Protocol:   discovery.ProtocolManual, // Demo devices use "manual" protocol
		}
	}

	if len(shellyDevices) == 0 {
		ios.NoResults("devices", "No discovery fixtures defined in demo mode")
		return nil
	}

	ios.Success("Discovered %d device(s) (demo mode)", len(shellyDevices))
	ios.Println()
	term.DisplayDiscoveredDevices(ios, shellyDevices)

	// Save to completion cache
	addresses := make([]string, len(shellyDevices))
	for i, d := range shellyDevices {
		addresses[i] = d.Address.String()
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	// Register devices if requested
	if register {
		added, err := utils.RegisterDiscoveredDevices(shellyDevices, skipExisting)
		if err != nil {
			ios.Warning("Registration error: %v", err)
		}
		ios.Added("device", added)
	}

	return nil
}
