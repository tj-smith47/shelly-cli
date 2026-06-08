// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"
	"golang.org/x/sync/errgroup"

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

// DefaultHTTPScanTimeout is the timeout for HTTP subnet scanning. A 15s budget
// truncates large subnets (a /23 has 510 hosts), missing devices; 2 minutes
// matches the CLI's cmdutil.DefaultScanTimeout so the TUI discovers the same set.
const DefaultHTTPScanTimeout = 2 * time.Minute

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

// DiscoverMDNSContext returns an mDNS discoverer plus a cleanup func so the
// caller can drive discovery with its own context via DiscoverWithContext.
// Threading the caller ctx into the scan is what lets Ctrl+C abort an
// in-progress mDNS sweep promptly instead of blocking the full timeout.
func DiscoverMDNSContext() (discoverer *discovery.MDNSDiscoverer, cleanup func()) {
	d := discovery.NewMDNSDiscoverer()
	cleanup = func() {
		if err := d.Stop(); err != nil {
			iostreams.DebugErr("stopping mDNS discoverer", err)
		}
	}
	return d, cleanup
}

// DiscoverMDNS performs mDNS/Zeroconf discovery honoring ctx cancellation.
// The timeout bounds the scan; ctx cancellation (Ctrl+C) aborts it sooner.
func DiscoverMDNS(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	disc, cleanup := DiscoverMDNSContext()
	defer cleanup()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return disc.DiscoverWithContext(ctx)
}

func discoverMDNS(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	return DiscoverMDNS(ctx, timeout)
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

// DiscoverCoIoTContext returns a CoIoT discoverer plus a cleanup func so the
// caller can drive discovery with its own context via DiscoverWithContext.
// Threading the caller ctx into the scan is what lets Ctrl+C abort an
// in-progress CoIoT sweep promptly instead of blocking the full timeout.
func DiscoverCoIoTContext() (discoverer *discovery.CoIoTDiscoverer, cleanup func()) {
	d := discovery.NewCoIoTDiscoverer()
	cleanup = func() {
		if err := d.Stop(); err != nil {
			iostreams.DebugErr("stopping CoIoT discoverer", err)
		}
	}
	return d, cleanup
}

// DiscoverCoIoT performs CoIoT/CoAP discovery honoring ctx cancellation.
// The timeout bounds the scan; ctx cancellation (Ctrl+C) aborts it sooner.
func DiscoverCoIoT(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	disc, cleanup := DiscoverCoIoTContext()
	defer cleanup()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return disc.DiscoverWithContext(ctx)
}

func discoverCoIoT(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	return DiscoverCoIoT(ctx, timeout)
}

// ErrBLEUnavailable indicates the BLE adapter could not be initialized
// (no adapter, not permitted, or not supported on this platform). Callers
// use errors.Is to distinguish a missing-adapter condition from a scan
// failure so they can present a friendly "BLE not available" message.
var ErrBLEUnavailable = errors.New("BLE not available")

// IsBLEUnavailable reports whether err signals that BLE could not be
// initialized on this system, as opposed to a transient scan failure.
func IsBLEUnavailable(err error) bool {
	return errors.Is(err, ErrBLEUnavailable)
}

// DiscoverBLE performs BLE discovery and returns raw discovered devices.
// Returns nil, nil, error if BLE is not available on the system.
// Caller is responsible for cleanup.
func DiscoverBLE() (*discovery.BLEDiscoverer, func(), error) {
	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrBLEUnavailable, err)
	}
	cleanup := func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			iostreams.DebugErr("stopping BLE discoverer", stopErr)
		}
	}
	return bleDiscoverer, cleanup, nil
}

// DiscoverBLEContext performs BLE discovery, applying the timeout to ctx
// exactly once. Returns the BLE-unavailable error so callers can decide how
// to surface it. This is the single deadline-application point shared by all
// BLE callers so the timeout semantics never diverge between flows.
func DiscoverBLEContext(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	bleDiscoverer, cleanup, err := DiscoverBLE()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return bleDiscoverer.DiscoverWithContext(ctx)
}

func discoverBLE(ctx context.Context, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	return DiscoverBLEContext(ctx, timeout)
}

// enrichDiscoveredDevices converts raw discovered devices and checks registry status.
//
// The per-device generation re-probe runs concurrently so a large result set
// completes well within the caller's deadline; a sequential loop could exhaust
// the enrichment budget before reaching every device, leaving some at Generation 0.
func enrichDiscoveredDevices(ctx context.Context, rawDevices []discovery.DiscoveredDevice) []DiscoveredDevice {
	result := make([]DiscoveredDevice, len(rawDevices))

	g, gctx := errgroup.WithContext(ctx)
	for i, raw := range rawDevices {
		g.Go(func() error {
			device := DiscoveredDevice{
				Address:    raw.Address.String(),
				Name:       raw.Name,
				Model:      raw.Model,
				ID:         raw.ID,
				Generation: int(raw.Generation),
				Firmware:   raw.Firmware,
				AuthEn:     raw.AuthRequired,
			}

			// The scan already populates generation/firmware/auth for mDNS, CoIoT
			// and HTTP results, so only re-probe when generation is still unknown;
			// a failed fallback probe must never zero out a known generation.
			if device.Generation == 0 {
				probeGeneration(gctx, &device)
			}

			// Concurrent registry read is safe only because config is never
			// mutated during a discovery scan; revisit this if that changes.
			device.Added = IsDeviceRegistered(device.Address)

			result[i] = device
			return nil
		})
	}

	// Goroutines only report transport-level probe failures via DebugErr and
	// never return an error, so Wait surfaces nothing actionable here.
	if err := g.Wait(); err != nil {
		iostreams.DebugErr("enriching discovered devices", err)
	}

	return result
}

// probeGeneration re-detects a device's generation when discovery left it at 0.
// On success it fills Generation, AuthEn, and any empty Firmware/Model; a failed
// probe is logged and leaves the device untouched so a known field is never
// overwritten with an empty value.
func probeGeneration(ctx context.Context, device *DiscoveredDevice) {
	detection, err := client.DetectGeneration(ctx, device.Address, nil)
	if err != nil {
		iostreams.DebugErr("detecting generation for "+device.Address, err)
		return
	}

	device.Generation = int(detection.Generation)
	device.AuthEn = detection.AuthEn
	if device.Firmware == "" {
		device.Firmware = detection.Firmware
	}
	if device.Model == "" {
		device.Model = detection.Model
	}
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

// fillDeviceInfo backfills empty Name/Model/ID fields on device from a live
// DeviceInfo query. A failed query is non-fatal (the device keeps whatever the
// scan supplied), and an empty DeviceInfo field never overwrites a value the
// scan already populated.
func fillDeviceInfo(ctx context.Context, device *DiscoveredDevice, svc *Service) {
	info, err := svc.DeviceInfo(ctx, device.Address)
	if err != nil {
		iostreams.DebugErr("fetching device info for "+device.Address, err)
		return
	}

	if device.Name == "" {
		device.Name = info.ID
	}
	// Never replace a known value with an empty one: a Gen2 DeviceInfo can
	// succeed with an empty Model, which would otherwise clobber a good model
	// already supplied by the scan.
	if device.Model == "" && info.Model != "" {
		device.Model = info.Model
	}
	if device.ID == "" && info.ID != "" {
		device.ID = info.ID
	}
}

// RegisterDiscoveredDevice adds a discovered device to the registry.
func RegisterDiscoveredDevice(ctx context.Context, device DiscoveredDevice, svc *Service) error {
	// Check if already registered
	if IsDeviceRegistered(device.Address) {
		return nil
	}

	// Get device info if not already populated.
	if device.Name == "" || device.Model == "" {
		fillDeviceInfo(ctx, &device, svc)
	}

	name := device.Name
	if name == "" {
		name = device.Address
	}

	// device.Model is the raw model code from discovery; route through the
	// shared helper so Type=raw-code and Model=display-name, matching every
	// other registration path.
	return utils.RegisterDeviceFromModelCode(name, device.Address, device.Generation, device.Model, nil)
}
