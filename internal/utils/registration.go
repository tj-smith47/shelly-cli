package utils

import (
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// DeviceRegistration holds all information needed to register a device.
// This unifies the registration path for different device source types.
type DeviceRegistration struct {
	Name       string      // Device name (required)
	Address    string      // Device address (required)
	Generation int         // Shelly generation (0 for unknown/plugin devices)
	Type       string      // Model code/SKU
	Model      string      // Human-readable model name
	Platform   string      // Platform (shelly, tasmota, etc.)
	MAC        string      // MAC address (optional)
	Auth       *model.Auth // Authentication credentials (optional)
}

// ToDevice converts a DeviceRegistration to a model.Device.
func (r DeviceRegistration) ToDevice() model.Device {
	platform := r.Platform
	if platform == "" {
		platform = model.PlatformShelly
	}
	return model.Device{
		Name:       r.Name,
		Address:    r.Address,
		Generation: r.Generation,
		Type:       r.Type,
		Model:      r.Model,
		Platform:   platform,
		MAC:        model.NormalizeMAC(r.MAC),
	}
}

// DiscoveredDeviceToRegistration converts a library discovery result to a DeviceRegistration.
func DiscoveredDeviceToRegistration(d discovery.DiscoveredDevice) DeviceRegistration {
	name := d.ID
	if d.Name != "" {
		name = d.Name
	}

	return DeviceRegistration{
		Name:       name,
		Address:    d.Address.String(),
		Generation: int(d.Generation),
		Type:       d.Model,
		Model:      types.ModelDisplayName(d.Model),
		Platform:   model.PlatformShelly,
		MAC:        d.MACAddress,
	}
}

// PluginResultToRegistration converts a plugin detection result to a DeviceRegistration.
func PluginResultToRegistration(result *plugins.DeviceDetectionResult, address string) DeviceRegistration {
	name := result.DeviceName
	if name == "" {
		name = result.DeviceID
	}
	if name == "" {
		name = address
	}

	return DeviceRegistration{
		Name:       name,
		Address:    address,
		Generation: 0,
		Type:       result.Model,
		Model:      types.ModelDisplayName(result.Model),
		Platform:   result.Platform,
	}
}

// PluginDeviceToRegistration converts a PluginDevice to a DeviceRegistration.
func PluginDeviceToRegistration(d PluginDevice) DeviceRegistration {
	name := d.Name
	if name == "" {
		name = d.ID
	}
	if name == "" {
		name = d.Address
	}

	return DeviceRegistration{
		Name:       name,
		Address:    d.Address,
		Generation: 0,
		Type:       d.Model,
		Model:      types.ModelDisplayName(d.Model),
		Platform:   d.Platform,
	}
}

// RegisterDevice registers a single device from a DeviceRegistration.
// Returns true if the device was added, false if skipped (already exists).
func RegisterDevice(reg DeviceRegistration, skipExisting bool) (bool, error) {
	if reg.Name == "" {
		return false, fmt.Errorf("device name is required")
	}
	if reg.Address == "" {
		return false, fmt.Errorf("device address is required")
	}

	// Check if already registered
	if skipExisting {
		if _, exists := config.GetDevice(reg.Name); exists {
			return false, nil
		}
	}

	// Determine which registration function to use based on platform
	var err error
	if reg.Platform != "" && reg.Platform != model.PlatformShelly {
		err = config.RegisterDeviceWithPlatform(
			reg.Name,
			reg.Address,
			reg.Generation,
			reg.Type,
			reg.Model,
			reg.Platform,
			reg.Auth,
		)
	} else {
		err = config.RegisterDevice(
			reg.Name,
			reg.Address,
			reg.Generation,
			reg.Type,
			reg.Model,
			reg.Auth,
		)
	}

	if err != nil {
		return false, fmt.Errorf("failed to register device %q: %w", reg.Name, err)
	}

	// Update MAC if available
	if reg.MAC != "" {
		if macErr := config.UpdateDeviceInfo(reg.Name, config.DeviceUpdates{
			MAC: reg.MAC,
		}); macErr != nil {
			iostreams.DebugErr("update MAC for "+reg.Name, macErr)
		}
	}

	return true, nil
}

// RegisterDevicesBatch registers multiple devices from registrations.
// Returns the number of devices added and the first error encountered.
func RegisterDevicesBatch(regs []DeviceRegistration, skipExisting bool) (int, error) {
	added := 0
	for _, reg := range regs {
		wasAdded, err := RegisterDevice(reg, skipExisting)
		if err != nil {
			return added, err
		}
		if wasAdded {
			added++
		}
	}
	return added, nil
}

// RegisterDevicesFromSlice registers devices from a slice of convertible types.
// The convert function transforms each item to a DeviceRegistration.
// Returns the number of devices added.
func RegisterDevicesFromSlice[T any](items []T, convert func(T) DeviceRegistration, skipExisting bool) int {
	added := 0
	for _, item := range items {
		reg := convert(item)
		if reg.Address == "" {
			continue // Skip items without address
		}
		wasAdded, err := RegisterDevice(reg, skipExisting)
		if err != nil {
			continue // Skip on error
		}
		if wasAdded {
			added++
		}
	}
	return added
}
