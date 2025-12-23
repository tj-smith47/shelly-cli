// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DiscoveredDevice represents a device found during discovery.
type DiscoveredDevice struct {
	ID         string
	Name       string
	Model      string
	Address    string
	Generation int
	Firmware   string
	AuthEn     bool
	Added      bool // Whether already in device registry
}

// DiscoveryOptions configures the discovery operation.
type DiscoveryOptions struct {
	Method     DiscoveryMethod
	Timeout    time.Duration
	Subnet     string // For HTTP scanning
	AutoDetect bool   // Auto-detect subnet for HTTP scanning
}

// DiscoveryMethod specifies the discovery protocol.
type DiscoveryMethod int

// Discovery method constants.
const (
	DiscoveryMDNS DiscoveryMethod = iota
	DiscoveryHTTP
	DiscoveryCoIoT
	DiscoveryBLE
)

// DefaultDiscoveryTimeout is the default discovery timeout.
const DefaultDiscoveryTimeout = 10 * time.Second

// DiscoverDevices discovers devices using the specified method.
func (s *Service) DiscoverDevices(ctx context.Context, opts DiscoveryOptions) ([]DiscoveredDevice, error) {
	if opts.Timeout == 0 {
		opts.Timeout = DefaultDiscoveryTimeout
	}

	var rawDevices []discovery.DiscoveredDevice
	var err error

	switch opts.Method {
	case DiscoveryMDNS:
		rawDevices, err = discoverMDNS(ctx, opts.Timeout)
	case DiscoveryHTTP:
		rawDevices, err = discoverHTTP(ctx, opts)
	case DiscoveryCoIoT:
		rawDevices, err = discoverCoIoT(ctx, opts.Timeout)
	case DiscoveryBLE:
		rawDevices, err = discoverBLE(ctx, opts.Timeout)
	default:
		return nil, fmt.Errorf("unknown discovery method: %d", opts.Method)
	}

	if err != nil {
		return nil, err
	}

	// Convert to our type and enrich with registry status
	return enrichDiscoveredDevices(ctx, rawDevices), nil
}

func discoverMDNS(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			iostreams.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices, err := mdnsDiscoverer.Discover(timeout)
	if err != nil && ctx.Err() != nil {
		// Context cancelled, not an error
		return devices, nil
	}
	return devices, err
}

func discoverHTTP(ctx context.Context, opts DiscoveryOptions) ([]discovery.DiscoveredDevice, error) {
	subnet := opts.Subnet
	if subnet == "" && opts.AutoDetect {
		var err error
		subnet, err = utils.DetectSubnet()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnet: %w", err)
		}
	}
	if subnet == "" {
		return nil, fmt.Errorf("subnet required for HTTP discovery")
	}

	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses to scan in subnet %s", subnet)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	return discovery.ProbeAddresses(ctx, addresses), nil
}

func discoverCoIoT(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			iostreams.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices, err := coiotDiscoverer.Discover(timeout)
	if err != nil && ctx.Err() != nil {
		return devices, nil
	}
	return devices, err
}

func discoverBLE(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, fmt.Errorf("BLE not available: %w", err)
	}
	defer func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			iostreams.DebugErr("stopping BLE discoverer", stopErr)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return bleDiscoverer.DiscoverWithContext(ctx)
}

// enrichDiscoveredDevices converts raw discovered devices and checks registry status.
func enrichDiscoveredDevices(ctx context.Context, rawDevices []discovery.DiscoveredDevice) []DiscoveredDevice {
	result := make([]DiscoveredDevice, 0, len(rawDevices))

	for _, raw := range rawDevices {
		device := DiscoveredDevice{
			Address: raw.Address.String(),
			Name:    raw.Name,
			Model:   raw.Model,
			ID:      raw.ID,
		}

		// Try to get additional info via detection
		detection, err := client.DetectGeneration(ctx, device.Address, nil)
		if err == nil {
			device.Generation = int(detection.Generation)
			device.Firmware = detection.Firmware
			device.AuthEn = detection.AuthEn
			if device.Model == "" {
				device.Model = detection.Model
			}
		}

		// Check if in registry
		device.Added = isDeviceRegistered(device.Address)

		result = append(result, device)
	}

	return result
}

// isDeviceRegistered checks if a device is in the registry.
func isDeviceRegistered(address string) bool {
	cfg := config.Get()
	if cfg == nil {
		return false
	}

	for _, dev := range cfg.Devices {
		if dev.Address == address {
			return true
		}
	}
	return false
}

// RegisterDiscoveredDevice adds a discovered device to the registry.
func RegisterDiscoveredDevice(ctx context.Context, device DiscoveredDevice, svc *Service) error {
	// Check if already registered
	if isDeviceRegistered(device.Address) {
		return nil
	}

	// Get device info if not already populated
	if device.Name == "" || device.Model == "" {
		info, err := svc.DeviceInfo(ctx, device.Address)
		if err == nil {
			if device.Name == "" {
				device.Name = info.ID
			}
			device.Model = info.Model
			device.ID = info.ID
		}
	}

	name := device.Name
	if name == "" {
		name = device.Address
	}

	return config.RegisterDevice(name, device.Address, device.Generation, "", device.Model, nil)
}
