package shelly

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/provisioning"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// OnboardSource indicates how an unprovisioned device was discovered.
type OnboardSource string

// Discovery source constants.
const (
	OnboardSourceBLE    OnboardSource = "BLE"
	OnboardSourceWiFiAP OnboardSource = "WiFi AP"
	OnboardSourceMDNS   OnboardSource = "mDNS"
	OnboardSourceCoIoT  OnboardSource = "CoIoT"
	OnboardSourceHTTP   OnboardSource = "HTTP"
)

// OnboardDevice represents a device discovered during the onboard scan.
// Unifies BLE, WiFi AP, and network discovery results into a single type.
type OnboardDevice struct {
	Name        string
	Model       string
	Address     string // IP for networked; BLE addr for BLE; 192.168.33.1 for AP
	MACAddress  string
	SSID        string        // WiFi AP SSID (only for WiFi AP source)
	BLEAddress  string        // BLE address (only for BLE source)
	Source      OnboardSource // How the device was found
	Generation  int
	RSSI        int
	Registered  bool // Already in config registry
	Provisioned bool // Already on network (has IP)
}

// OnboardWiFiConfig holds WiFi credentials for onboarding.
type OnboardWiFiConfig struct {
	SSID     string
	Password string
}

// OnboardOptions configures the onboard operation.
type OnboardOptions struct {
	WiFi         *OnboardWiFiConfig
	Subnet       string
	Timezone     string
	DeviceName   string
	Timeout      time.Duration
	BLEOnly      bool
	APOnly       bool
	NetworkOnly  bool
	RegisterOnly bool
	NoCloud      bool
}

// OnboardResult holds the outcome of onboarding a single device.
type OnboardResult struct {
	Device     *OnboardDevice
	NewAddress string
	Error      error
	Registered bool
	Method     string // "BLE", "WiFi AP", "register-only"
}

// OnboardProgress reports discovery progress for a single scan method.
type OnboardProgress struct {
	Method string
	Found  int
	Done   bool
	Err    error
}

// DiscoverForOnboard runs concurrent multi-protocol discovery to find
// unprovisioned and unregistered Shelly devices. Each discovery method
// runs in its own goroutine. Results are merged and deduplicated by MAC.
// The progress callback is invoked per-method for UI updates.
func (s *Service) DiscoverForOnboard(
	ctx context.Context,
	opts *OnboardOptions,
	progress func(OnboardProgress),
) ([]OnboardDevice, error) {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		mu      sync.Mutex
		devices []OnboardDevice
		wg      sync.WaitGroup
	)

	report := func(method string, found int, done bool, err error) {
		if progress != nil {
			progress(OnboardProgress{Method: method, Found: found, Done: done, Err: err})
		}
	}

	// BLE discovery (Gen2+ in provisioning mode)
	if !opts.NetworkOnly && !opts.APOnly {
		wg.Go(func() {
			report("BLE", 0, false, nil)
			found := s.discoverBLEForOnboard(ctx, timeout)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
			report("BLE", len(found), true, nil)
		})
	}

	// WiFi AP discovery (Gen1 unprovisioned)
	if !opts.BLEOnly && !opts.NetworkOnly {
		wg.Go(func() {
			report("WiFi AP", 0, false, nil)
			found, err := s.discoverWiFiAPForOnboard(ctx)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
			report("WiFi AP", len(found), true, err)
		})
	}

	// Network discovery (mDNS + CoIoT + HTTP for already-networked devices)
	if !opts.BLEOnly && !opts.APOnly {
		wg.Go(func() {
			report("Network", 0, false, nil)
			found := s.discoverNetworkForOnboard(ctx, opts, timeout)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
			report("Network", len(found), true, nil)
		})
	}

	wg.Wait()

	// Deduplicate by MAC address, preferring BLE source
	deduped := deduplicateOnboardDevices(devices)

	// Mark registered devices
	for i := range deduped {
		if deduped[i].Address != "" && deduped[i].Address != discovery.DefaultAPIP {
			deduped[i].Registered = IsDeviceRegistered(deduped[i].Address)
		}
	}

	return deduped, nil
}

// discoverBLEForOnboard runs BLE discovery and converts results to OnboardDevice.
func (s *Service) discoverBLEForOnboard(ctx context.Context, timeout time.Duration) []OnboardDevice {
	bleDiscoverer, cleanup, err := DiscoverBLE()
	if err != nil {
		debug.TraceEvent("onboard BLE discovery not available: %v", err)
		return nil
	}
	defer cleanup()

	bleCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rawDevices, err := bleDiscoverer.DiscoverWithContext(bleCtx)
	if err != nil {
		debug.TraceEvent("onboard BLE discovery error: %v", err)
		return nil
	}

	bleDetailedDevices := bleDiscoverer.GetDiscoveredDevices()

	var result []OnboardDevice
	for _, raw := range rawDevices {
		dev := OnboardDevice{
			Name:        raw.Name,
			Model:       raw.Model,
			MACAddress:  raw.MACAddress,
			BLEAddress:  raw.Address.String(),
			Source:      OnboardSourceBLE,
			Generation:  int(raw.Generation),
			Provisioned: false, // BLE devices are unprovisioned
		}

		// Get RSSI from detailed BLE info
		for _, detailed := range bleDetailedDevices {
			if detailed.ID == raw.ID || detailed.LocalName == raw.Name {
				dev.RSSI = detailed.RSSI
				if detailed.LocalName != "" {
					dev.Name = detailed.LocalName
				}
				break
			}
		}

		if dev.Name == "" {
			dev.Name = raw.ID
		}

		result = append(result, dev)
	}

	return result
}

// discoverWiFiAPForOnboard scans for Shelly WiFi AP SSIDs.
func (s *Service) discoverWiFiAPForOnboard(ctx context.Context) ([]OnboardDevice, error) {
	wifiDisc := discovery.NewWiFiDiscoverer()
	rawDevices, err := wifiDisc.DiscoverWithContext(ctx)
	if err != nil {
		debug.TraceEvent("onboard WiFi AP discovery error: %v", err)
		return nil, err
	}

	wifiDetails := wifiDisc.GetDiscoveredDevices()

	var result []OnboardDevice
	for i, raw := range rawDevices {
		dev := OnboardDevice{
			Name:        raw.Name,
			Model:       raw.Model,
			Address:     discovery.DefaultAPIP,
			MACAddress:  raw.MACAddress,
			Source:      OnboardSourceWiFiAP,
			Generation:  int(raw.Generation),
			Provisioned: false,
		}

		// Get SSID and signal from WiFi details
		if i < len(wifiDetails) {
			dev.SSID = wifiDetails[i].SSID
			dev.RSSI = wifiDetails[i].Signal
		}

		if dev.Name == "" && dev.SSID != "" {
			dev.Name = dev.SSID
		}

		result = append(result, dev)
	}

	return result, nil
}

// discoverNetworkForOnboard runs mDNS + CoIoT concurrently to find
// devices already on the network but not registered.
func (s *Service) discoverNetworkForOnboard(ctx context.Context, opts *OnboardOptions, timeout time.Duration) []OnboardDevice {
	var (
		mu      sync.Mutex
		devices []discovery.DiscoveredDevice
		wg      sync.WaitGroup
	)

	// mDNS (Gen2+)
	wg.Go(func() {
		mdnsTimeout := timeout
		if mdnsTimeout > 10*time.Second {
			mdnsTimeout = 10 * time.Second
		}
		found, cleanup, err := DiscoverMDNS(mdnsTimeout)
		defer cleanup()
		if err != nil {
			debug.TraceEvent("onboard mDNS discovery error: %v", err)
			return
		}
		mu.Lock()
		devices = append(devices, found...)
		mu.Unlock()
	})

	// CoIoT (Gen1)
	wg.Go(func() {
		coiotTimeout := timeout
		if coiotTimeout > 10*time.Second {
			coiotTimeout = 10 * time.Second
		}
		found, cleanup, err := DiscoverCoIoT(coiotTimeout)
		defer cleanup()
		if err != nil {
			debug.TraceEvent("onboard CoIoT discovery error: %v", err)
			return
		}
		mu.Lock()
		devices = append(devices, found...)
		mu.Unlock()
	})

	// HTTP subnet scan (if subnet available)
	if opts.Subnet != "" {
		wg.Go(func() {
			_, _, err := net.ParseCIDR(opts.Subnet)
			if err != nil {
				debug.TraceEvent("onboard HTTP scan invalid subnet: %v", err)
				return
			}
			addresses := discovery.GenerateSubnetAddresses(opts.Subnet)
			if len(addresses) == 0 {
				return
			}
			httpCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			found := discovery.ProbeAddresses(httpCtx, addresses)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
		})
	}

	wg.Wait()

	// Convert and enrich
	enriched := enrichDiscoveredDevices(ctx, devices)
	result := make([]OnboardDevice, 0, len(enriched))
	for _, d := range enriched {
		source := OnboardSourceMDNS
		// Try to determine the original protocol
		if d.Generation == 1 {
			source = OnboardSourceCoIoT
		}

		result = append(result, OnboardDevice{
			Name:        d.Name,
			Model:       d.Model,
			Address:     d.Address,
			MACAddress:  d.ID, // enriched ID often contains MAC-derived info
			Source:      source,
			Generation:  d.Generation,
			Registered:  d.Added,
			Provisioned: true, // Network-discovered devices are already provisioned
		})
	}

	return result
}

// deduplicateOnboardDevices merges duplicate devices found by multiple
// discovery methods. Prefers BLE source over others since it carries
// more provisioning information.
func deduplicateOnboardDevices(devices []OnboardDevice) []OnboardDevice {
	seen := make(map[string]int) // MAC → index in result
	result := make([]OnboardDevice, 0, len(devices))

	for _, dev := range devices {
		key := normalizeMAC(dev.MACAddress)
		if key == "" {
			key = strings.ToLower(dev.Name)
		}
		if key == "" {
			result = append(result, dev)
			continue
		}
		if idx, exists := seen[key]; exists {
			// Prefer BLE source — it carries the most provisioning info
			if dev.Source == OnboardSourceBLE && result[idx].Source != OnboardSourceBLE {
				result[idx] = dev
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, dev)
	}

	return result
}

// OnboardViaBLE provisions a Gen2+ device via BLE and waits for it
// to appear on the network. Returns the result with the device's new IP.
func (s *Service) OnboardViaBLE(
	ctx context.Context,
	device *OnboardDevice,
	wifi *OnboardWiFiConfig,
	opts *OnboardOptions,
) *OnboardResult {
	result := &OnboardResult{Device: device, Method: "BLE"}

	// Initialize BLE transmitter
	transmitter, err := provisioning.NewTinyGoBLETransmitter()
	if err != nil {
		result.Error = fmt.Errorf("BLE init failed: %w", err)
		return result
	}

	// Create provisioner
	bleProvisioner := provisioning.NewBLEProvisioner()
	bleProvisioner.Transmitter = transmitter

	// Register device
	model, _ := provisioning.ParseBLEDeviceName(device.BLEAddress)
	bleProvisioner.AddDiscoveredDevice(&provisioning.BLEDevice{
		Name:     device.Name,
		Address:  device.BLEAddress,
		Model:    model,
		IsShelly: provisioning.IsShellyDevice(device.Name),
	})

	// Build config
	bleConfig := &provisioning.BLEProvisionConfig{
		WiFi: &provisioning.WiFiConfig{
			SSID:     wifi.SSID,
			Password: wifi.Password,
		},
		DeviceName: opts.DeviceName,
		Timezone:   opts.Timezone,
	}
	if opts.NoCloud {
		disable := false
		bleConfig.EnableCloud = &disable
	}

	// Provision
	bleResult, err := bleProvisioner.ProvisionViaBLE(ctx, device.BLEAddress, bleConfig)
	if err != nil {
		result.Error = fmt.Errorf("BLE provisioning failed: %w", err)
		return result
	}
	if !bleResult.Success {
		result.Error = fmt.Errorf("BLE provisioning completed with errors")
		return result
	}

	// Wait for device to appear on network
	newIP, err := s.WaitForDeviceOnNetwork(ctx, device.Name, device.MACAddress, 30*time.Second)
	if err != nil {
		// Provisioning succeeded but couldn't find device on network
		debug.TraceEvent("onboard BLE post-provision detection failed for %s: %v", device.Name, err)
	}
	result.NewAddress = newIP

	// Register
	if newIP != "" {
		if regErr := RegisterOnboardedDevice(device, newIP); regErr != nil {
			debug.TraceEvent("onboard register %s: %v", device.Name, regErr)
		} else {
			result.Registered = true
		}
	}

	return result
}

// OnboardViaAP provisions a Gen1 device by connecting to its WiFi AP,
// configuring WiFi credentials, reconnecting to the home network, and
// waiting for the device to appear.
func (s *Service) OnboardViaAP(
	ctx context.Context,
	device *OnboardDevice,
	wifi *OnboardWiFiConfig,
	opts *OnboardOptions,
) *OnboardResult {
	result := &OnboardResult{Device: device, Method: "WiFi AP"}

	wifiDisc := discovery.NewWiFiDiscoverer()
	scanner := wifiDisc.Scanner
	if scanner == nil {
		result.Error = fmt.Errorf("WiFi scanning not supported on this platform")
		return result
	}

	// Remember current network for reconnection (may fail if not connected)
	originalNet, netErr := scanner.CurrentNetwork(ctx)
	if netErr != nil {
		debug.TraceEvent("onboard: could not detect current network: %v", netErr)
	}

	// Connect to Shelly AP (open network, no password)
	if err := scanner.Connect(ctx, device.SSID, ""); err != nil {
		result.Error = fmt.Errorf("failed to connect to Shelly AP %q: %w", device.SSID, err)
		return result
	}

	// Wait for DHCP assignment
	time.Sleep(2 * time.Second)

	// Configure WiFi on the device at 192.168.33.1
	// Use the provision service for Gen2+ at AP, or direct HTTP for Gen1
	configErr := s.configureWiFiAtAP(ctx, wifi.SSID, wifi.Password, device.Generation)

	// Always try to reconnect to home network, even if config failed.
	// Prefer the original network; fall back to the provisioning target.
	reconnectSSID := wifi.SSID
	reconnectPass := wifi.Password
	if originalNet != nil && originalNet.SSID != "" {
		reconnectSSID = originalNet.SSID
		reconnectPass = "" // original network uses saved credentials
	}
	if reconnErr := scanner.Connect(ctx, reconnectSSID, reconnectPass); reconnErr != nil {
		debug.TraceEvent("onboard reconnect to %s failed: %v", reconnectSSID, reconnErr)
		// If we tried the original network, fall back to the target WiFi
		if reconnectSSID != wifi.SSID {
			if err := scanner.Connect(ctx, wifi.SSID, wifi.Password); err != nil {
				debug.TraceEvent("onboard fallback connect to %s also failed: %v", wifi.SSID, err)
			}
		}
	}

	if configErr != nil {
		result.Error = fmt.Errorf("failed to configure WiFi on device: %w", configErr)
		return result
	}

	// Wait for device to appear on network
	newIP, err := s.WaitForDeviceOnNetwork(ctx, device.Name, device.MACAddress, 30*time.Second)
	if err != nil {
		debug.TraceEvent("onboard AP post-provision detection failed for %s: %v", device.Name, err)
	}
	result.NewAddress = newIP

	// Register
	if newIP != "" {
		if regErr := RegisterOnboardedDevice(device, newIP); regErr != nil {
			debug.TraceEvent("onboard register %s: %v", device.Name, regErr)
		} else {
			result.Registered = true
		}
	}

	return result
}

// configureWiFiAtAP sends WiFi credentials to a device at 192.168.33.1.
// Uses Gen1 HTTP settings API for Gen1 devices, RPC for Gen2+.
func (s *Service) configureWiFiAtAP(ctx context.Context, ssid, password string, generation int) error {
	address := discovery.DefaultAPIP

	if generation == 1 {
		// Gen1: use WithGen1Connection → Device().SetWiFiStation
		return s.WithGen1Connection(ctx, address, func(conn *client.Gen1Client) error {
			return conn.Device().SetWiFiStation(ctx, true, ssid, password)
		})
	}

	// Gen2+: use Service.ConfigureWiFi (RPC-based)
	return s.ConfigureWiFi(ctx, address, ssid, password)
}

// WaitForDeviceOnNetwork waits for a provisioned device to appear on the
// network by repeatedly scanning via mDNS (using MAC) with HTTP fallback.
// Returns the device's new IP address or an error if not found within timeout.
func (s *Service) WaitForDeviceOnNetwork(
	ctx context.Context,
	name string,
	mac string,
	timeout time.Duration,
) (string, error) {
	deadline := time.Now().Add(timeout)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Try mDNS discovery by MAC
		if mac != "" {
			ip, err := s.DiscoverByMAC(ctx, mac)
			if err == nil && ip != "" {
				return ip, nil
			}
		}

		time.Sleep(interval)
	}

	return "", fmt.Errorf("device %q not found on network within %s", name, timeout)
}

// RegisterOnboardedDevice adds a successfully onboarded device to the config registry.
func RegisterOnboardedDevice(device *OnboardDevice, newAddress string) error {
	if newAddress == "" {
		return fmt.Errorf("no address for device %q", device.Name)
	}

	// Already registered?
	if IsDeviceRegistered(newAddress) {
		return nil
	}

	// Derive a clean name for registration
	name := sanitizeDeviceName(device.Name)
	if name == "" {
		name = newAddress
	}

	return config.RegisterDevice(name, newAddress, device.Generation, device.Model, "", nil)
}

// sanitizeDeviceName creates a config-safe device name from a discovery name.
// Lowercases and replaces spaces with hyphens.
func sanitizeDeviceName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	// Remove characters that aren't alphanumeric, hyphen, or underscore
	var clean strings.Builder
	for _, c := range name {
		switch {
		case c >= 'a' && c <= 'z':
			clean.WriteRune(c)
		case c >= '0' && c <= '9':
			clean.WriteRune(c)
		case c == '-' || c == '_':
			clean.WriteRune(c)
		}
	}
	return clean.String()
}

// FilterUnregistered returns only devices that are not already in the config registry.
func FilterUnregistered(devices []OnboardDevice) []OnboardDevice {
	result := make([]OnboardDevice, 0, len(devices))
	for _, d := range devices {
		if !d.Registered {
			result = append(result, d)
		}
	}
	return result
}

// SplitBySource separates devices by their discovery source for routing
// to the appropriate provisioning method.
func SplitBySource(devices []OnboardDevice) (ble, ap, network []*OnboardDevice) {
	for i := range devices {
		switch devices[i].Source {
		case OnboardSourceBLE:
			ble = append(ble, &devices[i])
		case OnboardSourceWiFiAP:
			ap = append(ap, &devices[i])
		default:
			network = append(network, &devices[i])
		}
	}
	return ble, ap, network
}

// OnboardBLEParallel provisions multiple BLE devices concurrently.
func (s *Service) OnboardBLEParallel(
	ctx context.Context,
	devices []*OnboardDevice,
	wifiCfg *OnboardWiFiConfig,
	opts *OnboardOptions,
) []*OnboardResult {
	results := make([]*OnboardResult, len(devices))
	var wg sync.WaitGroup

	for i, dev := range devices {
		wg.Go(func() {
			results[i] = s.OnboardViaBLE(ctx, dev, wifiCfg, opts)
		})
	}

	wg.Wait()
	return results
}

// RegisterNetworkDevices registers already-networked devices in config.
func RegisterNetworkDevices(devices []*OnboardDevice) []*OnboardResult {
	results := make([]*OnboardResult, 0, len(devices))
	for _, dev := range devices {
		r := &OnboardResult{Device: dev, Method: "register-only"}
		if regErr := RegisterOnboardedDevice(dev, dev.Address); regErr != nil {
			r.Error = regErr
		} else {
			r.NewAddress = dev.Address
			r.Registered = true
		}
		results = append(results, r)
	}
	return results
}
