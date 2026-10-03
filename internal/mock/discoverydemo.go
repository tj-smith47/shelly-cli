// Package mock provides demo discovery functionality.
package mock

import (
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
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

	status := cmdutil.StatusStreams(ios)
	if len(shellyDevices) > 0 {
		status.Success("Discovered %d device(s) (demo mode)", len(shellyDevices))
		status.Println()
	}
	if err := cmdutil.PrintDiscovered(ios, shellyDevices, term.DisplayDiscoveredDevices,
		"devices", "No discovery fixtures defined in demo mode"); err != nil {
		return err
	}
	if len(shellyDevices) == 0 {
		return nil
	}

	added := cmdutil.CacheAndRegisterDevices(status, shellyDevices, register, skipExisting)
	if register {
		status.Added("device", added)
	}

	return nil
}
