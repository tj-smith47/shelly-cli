// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// DeviceSpec represents a device parsed from --device flag.
type DeviceSpec struct {
	Name     string
	Address  string
	Username string
	Password string
}

// JSONDevice represents a device in JSON format.
type JSONDevice struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ParseDeviceSpec parses a device specification in format: name=ip[:user:pass]
// Examples:
//   - kitchen=192.168.1.100
//   - secure=192.168.1.102:admin:secret
func ParseDeviceSpec(spec string) (*DeviceSpec, error) {
	// Split name and address
	parts := strings.SplitN(spec, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid device spec %q: expected format name=ip[:user:pass]", spec)
	}

	name := strings.TrimSpace(parts[0])
	if name == "" {
		return nil, fmt.Errorf("invalid device spec %q: name cannot be empty", spec)
	}

	addrPart := strings.TrimSpace(parts[1])
	if addrPart == "" {
		return nil, fmt.Errorf("invalid device spec %q: address cannot be empty", spec)
	}

	device := &DeviceSpec{Name: name}

	// Check for auth: ip:user:pass
	addrParts := strings.SplitN(addrPart, ":", 3)
	device.Address = addrParts[0]

	if len(addrParts) == 3 {
		device.Username = addrParts[1]
		device.Password = addrParts[2]
	} else if len(addrParts) == 2 {
		// Could be ip:port or ip:user (ambiguous without password)
		// Treat as ip:port unless it looks like a username
		// Since Shelly auth always requires both user and pass, treat as ip:port
		device.Address = addrParts[0] + ":" + addrParts[1]
	}

	return device, nil
}

// ParseDevicesJSON parses device(s) from JSON.
// Accepts:
//   - File path (if it exists as a file)
//   - JSON array: [{"name":"kitchen","address":"192.168.1.100"}]
//   - Single JSON object: {"name":"kitchen","address":"192.168.1.100"}
func ParseDevicesJSON(input string) ([]JSONDevice, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	var jsonData []byte

	// Check if input is a file path
	if _, err := os.Stat(input); err == nil {
		data, err := os.ReadFile(input) //nolint:gosec // User-provided file path is intentional
		if err != nil {
			return nil, fmt.Errorf("failed to read devices file %q: %w", input, err)
		}
		jsonData = data
	} else {
		jsonData = []byte(input)
	}

	// Trim whitespace for detection
	trimmed := strings.TrimSpace(string(jsonData))
	if trimmed == "" {
		return nil, nil
	}

	// Try parsing as array first
	if strings.HasPrefix(trimmed, "[") {
		var devices []JSONDevice
		if err := json.Unmarshal(jsonData, &devices); err != nil {
			return nil, fmt.Errorf("invalid JSON array: %w", err)
		}
		return devices, nil
	}

	// Try parsing as single object
	if strings.HasPrefix(trimmed, "{") {
		var device JSONDevice
		if err := json.Unmarshal(jsonData, &device); err != nil {
			return nil, fmt.Errorf("invalid JSON object: %w", err)
		}
		return []JSONDevice{device}, nil
	}

	return nil, fmt.Errorf("invalid JSON: must be an array or object")
}

// RegisterDevicesFromFlags registers devices from --device and --devices-json flags.
// Returns the number of devices registered and any error.
func RegisterDevicesFromFlags(deviceSpecs, devicesJSON []string) (int, error) {
	registered := 0
	var errs []string

	// Process --device flags
	for _, spec := range deviceSpecs {
		device, err := ParseDeviceSpec(spec)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		if err := registerParsedDevice(device.Name, device.Address, device.Username, device.Password); err != nil {
			errs = append(errs, fmt.Sprintf("failed to register %q: %v", device.Name, err))
			continue
		}
		registered++
	}

	// Process --devices-json flags
	for _, jsonInput := range devicesJSON {
		devices, err := ParseDevicesJSON(jsonInput)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		for _, device := range devices {
			if device.Name == "" {
				errs = append(errs, "JSON device missing required 'name' field")
				continue
			}
			if device.Address == "" {
				errs = append(errs, fmt.Sprintf("JSON device %q missing required 'address' field", device.Name))
				continue
			}

			if err := registerParsedDevice(device.Name, device.Address, device.Username, device.Password); err != nil {
				errs = append(errs, fmt.Sprintf("failed to register %q: %v", device.Name, err))
				continue
			}
			registered++
		}
	}

	if len(errs) > 0 {
		return registered, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return registered, nil
}

// registerParsedDevice registers a single device with optional auth.
// Generation is auto-discovered on first use, so we default to 0 (unknown).
func registerParsedDevice(name, address, username, password string) error {
	var auth *model.Auth
	if username != "" && password != "" {
		auth = &model.Auth{
			Username: username,
			Password: password,
		}
	}

	// Register with generation=0 (unknown) - will be auto-detected on first use
	// Device type and model will also be detected on first use
	return config.RegisterDevice(name, address, 0, "", "", auth)
}
