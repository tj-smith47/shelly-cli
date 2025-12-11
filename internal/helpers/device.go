// Package helpers provides common functionality shared across CLI commands.
package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/gen2"
	"github.com/tj-smith47/shelly-go/rpc"
	"github.com/tj-smith47/shelly-go/transport"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// DeviceConnection represents an active connection to a Shelly device.
type DeviceConnection struct {
	Device     *gen2.Device
	Client     *rpc.Client
	Transport  transport.Transport
	Config     config.Device
	Generation int
}

// Close closes the device connection and releases resources.
func (dc *DeviceConnection) Close() error {
	if dc.Device != nil {
		return dc.Device.Close()
	}
	if dc.Client != nil {
		return dc.Client.Close()
	}
	return nil
}

// CloseQuietly closes the device connection, logging any errors in verbose mode.
// Use this in defer statements where connection close errors are not critical.
func (dc *DeviceConnection) CloseQuietly() {
	if err := dc.Close(); err != nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "debug: failed to close connection: %v\n", err)
		}
	}
}

// ConnectToDevice establishes a connection to a device by name or address.
// It resolves the device from the registry if a name is provided,
// or treats the identifier as a direct address.
func ConnectToDevice(ctx context.Context, identifier string) (*DeviceConnection, error) {
	// Resolve device from registry or use as address
	device, err := config.ResolveDevice(identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device %q: %w", identifier, err)
	}

	return ConnectToDeviceConfig(ctx, device)
}

// ConnectToDeviceConfig establishes a connection using a device config.
func ConnectToDeviceConfig(ctx context.Context, device config.Device) (*DeviceConnection, error) {
	// Build URL from address
	url := device.Address
	if url != "" && url[0] != 'h' {
		url = "http://" + url
	}

	// Build transport options
	var opts []transport.Option

	// Configure authentication if provided
	if device.Auth != nil && device.Auth.Password != "" {
		opts = append(opts, transport.WithAuth(device.Auth.Username, device.Auth.Password))
	}

	// Create HTTP transport with options
	httpTransport := transport.NewHTTP(url, opts...)

	// Create RPC client
	client := rpc.NewClient(httpTransport)

	// Create Gen2+ device wrapper
	gen2Device := gen2.NewDevice(client)

	// Try to get device info to verify connection
	info, err := gen2Device.GetDeviceInfo(ctx)
	if err != nil {
		// Close on error
		if closeErr := gen2Device.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to connect to device: %w (close error: %w)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to connect to device %q: %w", device.Address, err)
	}

	return &DeviceConnection{
		Device:     gen2Device,
		Client:     client,
		Transport:  httpTransport,
		Config:     device,
		Generation: info.Gen,
	}, nil
}

// DiscoveredDeviceToConfig converts a discovered device to a config.Device.
func DiscoveredDeviceToConfig(d discovery.DiscoveredDevice) config.Device {
	name := d.ID
	if d.Name != "" {
		name = d.Name
	}

	return config.Device{
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

// DefaultTimeout returns the default timeout for device operations.
func DefaultTimeout() time.Duration {
	return 10 * time.Second
}

// DefaultDiscoveryTimeout returns the default timeout for discovery operations.
func DefaultDiscoveryTimeout() time.Duration {
	return 10 * time.Second
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
