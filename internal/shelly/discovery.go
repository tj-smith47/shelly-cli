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
	Subnets    []string // For HTTP scanning (multiple subnets supported)
	AutoDetect bool     // Auto-detect subnets for HTTP scanning
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

// DiscoverMDNS performs mDNS/Zeroconf discovery and returns raw discovered devices.
// Caller is responsible for timeout handling and cleanup.
func DiscoverMDNS(timeout time.Duration) ([]discovery.DiscoveredDevice, func(), error) {
	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	cleanup := func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			iostreams.DebugErr("stopping mDNS discoverer", err)
		}
	}

	devices, err := mdnsDiscoverer.Discover(timeout)
	return devices, cleanup, err
}

func discoverMDNS(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	devices, cleanup, err := DiscoverMDNS(timeout)
	defer cleanup()

	if err != nil && ctx.Err() != nil {
		// Context cancelled, not an error
		return devices, nil
	}
	return devices, err
}

func discoverHTTP(ctx context.Context, opts DiscoveryOptions) ([]discovery.DiscoveredDevice, error) {
	subnets := opts.Subnets
	if len(subnets) == 0 && opts.AutoDetect {
		var err error
		subnets, err = utils.DetectSubnets()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnets: %w", err)
		}
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("subnet required for HTTP discovery")
	}

	var allAddresses []string
	for _, subnet := range subnets {
		if _, _, err := net.ParseCIDR(subnet); err != nil {
			return nil, fmt.Errorf("invalid subnet %q: %w", subnet, err)
		}
		allAddresses = append(allAddresses, discovery.GenerateSubnetAddresses(subnet)...)
	}
	if len(allAddresses) == 0 {
		return nil, fmt.Errorf("no addresses to scan in subnets")
	}

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	return discovery.ProbeAddresses(ctx, allAddresses), nil
}

// DiscoverCoIoT performs CoIoT/CoAP discovery and returns raw discovered devices.
// Caller is responsible for timeout handling and cleanup.
func DiscoverCoIoT(timeout time.Duration) ([]discovery.DiscoveredDevice, func(), error) {
	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	cleanup := func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			iostreams.DebugErr("stopping CoIoT discoverer", err)
		}
	}

	devices, err := coiotDiscoverer.Discover(timeout)
	return devices, cleanup, err
}

func discoverCoIoT(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	devices, cleanup, err := DiscoverCoIoT(timeout)
	defer cleanup()

	if err != nil && ctx.Err() != nil {
		return devices, nil
	}
	return devices, err
}

// DiscoverBLE performs BLE discovery and returns raw discovered devices.
// Returns nil, nil, error if BLE is not available on the system.
// Caller is responsible for cleanup.
func DiscoverBLE() (*discovery.BLEDiscoverer, func(), error) {
	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, nil, fmt.Errorf("BLE not available: %w", err)
	}
	cleanup := func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			iostreams.DebugErr("stopping BLE discoverer", stopErr)
		}
	}
	return bleDiscoverer, cleanup, nil
}

func discoverBLE(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	bleDiscoverer, cleanup, err := DiscoverBLE()
	if err != nil {
		return nil, err
	}
	defer cleanup()

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
		device.Added = IsDeviceRegistered(device.Address)

		result = append(result, device)
	}

	return result
}

// IsDeviceRegistered checks if a device is in the registry.
func IsDeviceRegistered(address string) bool {
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

// DiscoverByMAC performs a quick mDNS scan to find a device's current IP by MAC address.
// This is used for IP remapping when a device gets a new DHCP address.
// Returns the IP if found, or empty string if not discoverable within timeout.
// Default timeout is 2 seconds for quick recovery.
func (s *Service) DiscoverByMAC(ctx context.Context, mac string) (string, error) {
	// Normalize MAC for comparison
	normalizedMAC := normalizeMAC(mac)
	if normalizedMAC == "" {
		return "", fmt.Errorf("invalid MAC address: %q", mac)
	}

	// Quick mDNS scan (2 seconds)
	timeout := 2 * time.Second
	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			iostreams.DebugErrCat(iostreams.CategoryDiscovery, "stopping mDNS discoverer", err)
		}
	}()

	devices, err := mdnsDiscoverer.Discover(timeout)
	if err != nil {
		return "", fmt.Errorf("mDNS discovery failed: %w", err)
	}

	// Find device by MAC
	for _, d := range devices {
		if normalizeMAC(d.MACAddress) == normalizedMAC {
			return d.Address.String(), nil
		}
	}

	return "", nil // Not found
}

// normalizeMAC normalizes a MAC address for comparison (uppercase, no separators).
func normalizeMAC(mac string) string {
	if mac == "" {
		return ""
	}

	// Remove separators and uppercase
	result := make([]byte, 0, 12)
	for _, c := range mac {
		switch {
		case c >= '0' && c <= '9':
			result = append(result, byte(c))
		case c >= 'A' && c <= 'F':
			result = append(result, byte(c))
		case c >= 'a' && c <= 'f':
			result = append(result, byte(c-32)) // Convert to uppercase
		}
	}

	if len(result) != 12 {
		return ""
	}
	return string(result)
}

// RegisterDiscoveredDevice adds a discovered device to the registry.
func RegisterDiscoveredDevice(ctx context.Context, device DiscoveredDevice, svc *Service) error {
	// Check if already registered
	if IsDeviceRegistered(device.Address) {
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
