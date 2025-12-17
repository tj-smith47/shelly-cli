// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DiscoveredDeviceToConfig converts a discovered device to a model.Device.
func DiscoveredDeviceToConfig(d discovery.DiscoveredDevice) model.Device {
	name := d.ID
	if d.Name != "" {
		name = d.Name
	}

	return model.Device{
		Name:       name,
		Address:    d.Address.String(),
		Generation: int(d.Generation),
		Type:       d.Model,
		Model:      d.Model,
	}
}

// RegisterDiscoveredDevices adds discovered devices to the registry.
// Returns the number of devices added and any error.
func RegisterDiscoveredDevices(devices []discovery.DiscoveredDevice, skipExisting bool) (int, error) {
	added := 0

	for _, d := range devices {
		name := d.ID
		if d.Name != "" {
			name = d.Name
		}

		// Skip if already registered
		if skipExisting {
			if _, exists := config.GetDevice(name); exists {
				continue
			}
		}

		err := config.RegisterDevice(
			name,
			d.Address.String(),
			int(d.Generation),
			d.Model,
			d.Model,
			nil,
		)
		if err != nil {
			return added, fmt.Errorf("failed to register device %q: %w", name, err)
		}
		added++
	}

	return added, nil
}

// UnmarshalJSON converts an RPC response (interface{}) to a typed struct.
// It re-serializes the interface{} to JSON then unmarshals to the target type.
func UnmarshalJSON(data, v any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
