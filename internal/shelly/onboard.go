package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"
	gen1 "github.com/tj-smith47/shelly-go/gen1"
	"github.com/tj-smith47/shelly-go/provisioning"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/shelly/provision"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/utils"
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

// DefaultOnboardScanTimeout is the default discovery budget for onboarding when
// no timeout is supplied. Onboard discovery runs WiFi-AP retry sweeps (multiple
// 3s iterations) concurrently with a BLE sweep, so it needs the same 2-minute
// budget the rest of discovery uses (cmdutil.DefaultScanTimeout); 30s truncated
// larger subnet scans. Defined here, the lowest layer, to avoid an import cycle
// (cmdutil already imports shelly).
const DefaultOnboardScanTimeout = 2 * time.Minute

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
	// Static IP configuration (all four required together; empty StaticIP = DHCP).
	StaticIP string // device's static address on the target network
	Gateway  string
	Netmask  string
	DNS      string
}

// IsStatic reports whether a static IP was requested (StaticIP set).
func (c *OnboardWiFiConfig) IsStatic() bool {
	return c != nil && c.StaticIP != ""
}

// OnboardOptions configures the onboard operation.
type OnboardOptions struct {
	WiFi       *OnboardWiFiConfig
	Timezone   string
	DeviceName string
	Timeout    time.Duration
	BLEOnly    bool
	APOnly     bool
	NoCloud    bool
}

// OnboardResult holds the outcome of onboarding a single device.
type OnboardResult struct {
	Device     *OnboardDevice
	NewAddress string
	Error      error
	// Note carries a non-fatal warning, e.g. the device was provisioned but
	// could not be located on the network afterward. A non-empty Note with a
	// nil Error and an empty NewAddress means the device is in a partial state:
	// source config was not applied and it is not browsable yet.
	Note       string
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
		timeout = DefaultOnboardScanTimeout
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
	if !opts.APOnly {
		wg.Go(func() {
			report(string(OnboardSourceBLE), 0, false, nil)
			found, bleErr := s.discoverBLEForOnboard(ctx)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
			report(string(OnboardSourceBLE), len(found), true, bleErr)
		})
	}

	// WiFi AP discovery (Gen1 unprovisioned)
	if !opts.BLEOnly {
		wg.Go(func() {
			report("WiFi AP", 0, false, nil)
			found, err := s.discoverWiFiAPForOnboard(ctx)
			mu.Lock()
			devices = append(devices, found...)
			mu.Unlock()
			report("WiFi AP", len(found), true, err)
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
// The timeout is already applied to ctx by DiscoverForOnboard, so no second
// WithTimeout is needed here. A BLE-unavailable condition is returned as an
// error so the caller can surface it via OnboardProgress.Err, matching the
// user-visible feedback the discover and wizard flows give for the same case.
func (s *Service) discoverBLEForOnboard(ctx context.Context) ([]OnboardDevice, error) {
	bleDiscoverer, cleanup, err := DiscoverBLE()
	if err != nil {
		debug.TraceEvent("onboard BLE discovery not available: %v", err)
		return nil, err
	}
	defer cleanup()

	rawDevices, err := bleDiscoverer.DiscoverWithContext(ctx)
	if err != nil {
		debug.TraceEvent("onboard BLE discovery error: %v", err)
		return nil, err
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
			// ID (the BLE address, the map key) is always non-empty and uniquely
			// identifies the device. The LocalName fallback must guard against
			// empty names: "" == "" would otherwise match the first nameless
			// record and copy its RSSI onto the wrong device.
			if detailed.ID == raw.ID || (raw.Name != "" && detailed.LocalName == raw.Name) {
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

	return result, nil
}

// discoverWiFiAPForOnboard scans for Shelly WiFi AP SSIDs.
// WiFi scans are inherently unreliable — a single sweep may miss APs on
// different channels or with weak signal. This function retries the scan
// every 3 seconds until the context deadline, accumulating unique results.
func (s *Service) discoverWiFiAPForOnboard(ctx context.Context) ([]OnboardDevice, error) {
	wifiDisc := discovery.NewWiFiDiscoverer()
	seen := make(map[string]OnboardDevice) // keyed by SSID

	const scanInterval = 3 * time.Second
	var lastErr error
	firstAttempt := true

	for {
		rawDevices, err := wifiDisc.DiscoverWithContext(ctx)
		if err != nil {
			debug.TraceEvent("onboard WiFi AP scan attempt error: %v", err)
			lastErr = err
			// A platform/tooling failure (no WiFi scanner, missing
			// nmcli/wpa_cli/iwconfig) surfaces as a *discovery.WiFiError and
			// recurs on every sweep; retrying only burns the discovery budget.
			var wifiErr *discovery.WiFiError
			if firstAttempt && errors.As(err, &wifiErr) {
				return nil, err
			}
		}
		firstAttempt = false

		s.collectWiFiAPDevices(wifiDisc, rawDevices, seen)

		// Wait before next scan, or exit if context is done.
		select {
		case <-ctx.Done():
			// Return whatever we found so far.
			if len(seen) == 0 && lastErr != nil {
				return nil, lastErr
			}
			result := make([]OnboardDevice, 0, len(seen))
			for _, dev := range seen {
				result = append(result, dev)
			}
			return result, nil
		case <-time.After(scanInterval):
			// Continue scanning.
		}
	}
}

// collectWiFiAPDevices merges one WiFi scan's results into seen (keyed by SSID,
// falling back to MAC). SSID/RSSI come from the discoverer's detail map keyed by
// SSID and joined on raw.Name (which equals network.SSID for AP-discovered
// devices); index pairing would attach SSID/RSSI to the wrong device because
// GetDiscoveredDevices ranges a map whose order is unrelated to rawDevices and
// accumulates across sweeps.
func (s *Service) collectWiFiAPDevices(
	wifiDisc *discovery.WiFiDiscoverer,
	rawDevices []discovery.DiscoveredDevice,
	seen map[string]OnboardDevice,
) {
	wifiDetails := wifiDisc.GetDiscoveredDevices()
	detailBySSID := make(map[string]discovery.WiFiDiscoveredDevice, len(wifiDetails))
	for _, d := range wifiDetails {
		detailBySSID[d.SSID] = d
	}

	for _, raw := range rawDevices {
		dev := OnboardDevice{
			Name:        raw.Name,
			Model:       raw.Model,
			Address:     discovery.DefaultAPIP,
			MACAddress:  raw.MACAddress,
			Source:      OnboardSourceWiFiAP,
			Generation:  int(raw.Generation),
			Provisioned: false,
		}

		if d, ok := detailBySSID[raw.Name]; ok {
			dev.SSID = d.SSID
			dev.RSSI = d.Signal
		}
		if dev.SSID == "" {
			dev.SSID = raw.Name // raw.Name is the SSID; keeps the dedup key correct
		}

		if dev.Name == "" && dev.SSID != "" {
			dev.Name = dev.SSID
		}

		key := dev.SSID
		if key == "" {
			key = dev.MACAddress
		}
		if key != "" {
			seen[key] = dev
		}
	}
}

// shellyDeviceIDSuffix extracts the trailing hex device-ID from a Shelly
// LocalName or AP SSID (e.g. "ShellyPlus1PM-AABBCC" → "aabbcc"). This suffix is
// shared across a device's BLE and WiFi-AP advertisements, so it is the stable
// key for collapsing BLE and AP rows for the same physical device — their MACs
// differ (BT radio vs WiFi BSSID) and never collide. Only Shelly-formatted
// names (with the "shelly" prefix) are considered, so arbitrary names cannot
// produce a spurious suffix that splits a legitimate MAC-keyed duplicate.
// Returns "" when the input is not a Shelly name or carries no device ID.
func shellyDeviceIDSuffix(name string) string {
	if !strings.HasPrefix(strings.ToLower(name), "shelly") {
		return ""
	}
	_, deviceID := discovery.ParseShellySSID(name)
	return strings.ToLower(deviceID)
}

// deduplicateOnboardDevices merges duplicate devices found by multiple
// discovery methods. Prefers BLE source over others since it carries
// more provisioning information.
func deduplicateOnboardDevices(devices []OnboardDevice) []OnboardDevice {
	seen := make(map[string]int) // MAC → index in result
	result := make([]OnboardDevice, 0, len(devices))

	for _, dev := range devices {
		// A device advertising both BLE and a WiFi AP has DISTINCT MACs (BT radio
		// vs WiFi BSSID), so a MAC key never collapses the two sources. The shared
		// Shelly device-ID hex suffix is the only stable cross-source key, so try
		// it first (from BLE LocalName, then AP SSID) before falling back to MAC.
		key := shellyDeviceIDSuffix(dev.Name)
		if key == "" {
			key = shellyDeviceIDSuffix(dev.SSID)
		}
		if key == "" {
			key = normalizeMAC(dev.MACAddress)
		}
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
	result := &OnboardResult{Device: device, Method: string(OnboardSourceBLE)}

	// Initialize BLE transmitter
	transmitter, err := provisioning.NewTinyGoBLETransmitter()
	if err != nil {
		result.Error = fmt.Errorf("BLE init failed: %w", err)
		return result
	}

	// Create provisioner
	bleProvisioner := provisioning.NewBLEProvisioner()
	bleProvisioner.Transmitter = transmitter

	// Register device. ParseBLEDeviceName expects the Shelly LocalName
	// (e.g. "ShellyPlus1-AABBCC"), not the BLE hardware address — the address
	// never starts with "Shelly" so IsShellyDevice would reject it and the
	// model would always come back empty.
	model, _ := provisioning.ParseBLEDeviceName(device.Name)
	if model == "" {
		model = device.Model
	}
	bleProvisioner.AddDiscoveredDevice(&provisioning.BLEDevice{
		Name:     device.Name,
		Address:  device.BLEAddress,
		Model:    model,
		IsShelly: provisioning.IsShellyDevice(device.Name),
	})

	// Build config
	bleWiFi := &provisioning.WiFiConfig{
		SSID:     wifi.SSID,
		Password: wifi.Password,
	}
	if wifi.IsStatic() {
		bleWiFi.StaticIP = "static"
		bleWiFi.IP = wifi.StaticIP
		bleWiFi.Netmask = wifi.Netmask
		bleWiFi.Gateway = wifi.Gateway
		bleWiFi.Nameserver = wifi.DNS
	}
	bleConfig := &provisioning.BLEProvisionConfig{
		WiFi:       bleWiFi,
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
		// Defensive: today a failed provision surfaces via the err != nil branch
		// above, but if the SDK ever returns Success=false with a nil err, carry
		// the result-level detail instead of a fixed, non-actionable string.
		if bleResult.Error != nil {
			result.Error = fmt.Errorf("BLE provisioning failed: %w", bleResult.Error)
		} else {
			result.Error = fmt.Errorf("BLE provisioning reported failure with no detail")
		}
		return result
	}

	// Wait for device to appear on network
	newIP, err := s.WaitForDeviceOnNetwork(ctx, device.Name, device.MACAddress, 30*time.Second)
	if err != nil {
		// Provisioning succeeded but couldn't find device on network. Carry the
		// cause so the UI can warn instead of falsely reporting clean success.
		debug.TraceEvent("onboard BLE post-provision detection failed for %s: %v", device.Name, err)
		result.Note = fmt.Sprintf("provisioned but not found on network: %v", err)
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
	configErr := s.configureWiFiAtAP(ctx, wifi, device.Generation)

	// Always try to reconnect to home network, even if config failed.
	// Determine the reconnect target: if the original network is different from
	// the provisioning target, try it first with saved creds (NetworkManager).
	// Otherwise use the target WiFi credentials — nl80211 has no credential store.
	reconnectSSID, reconnectPass := reconnectCredentials(originalNet, wifi)
	if reconnErr := scanner.Connect(ctx, reconnectSSID, reconnectPass); reconnErr != nil {
		debug.TraceEvent("onboard reconnect to %s failed: %v", reconnectSSID, reconnErr)
		// If we tried the original network, fall back to the provisioning target.
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
		// Provisioning succeeded but couldn't find device on network. Carry the
		// cause so the UI can warn instead of falsely reporting clean success.
		debug.TraceEvent("onboard AP post-provision detection failed for %s: %v", device.Name, err)
		result.Note = fmt.Sprintf("provisioned but not found on network: %v", err)
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
func (s *Service) configureWiFiAtAP(ctx context.Context, wifi *OnboardWiFiConfig, generation int) error {
	address := discovery.DefaultAPIP

	if generation == 1 {
		// Gen1: use WithGen1Connection → Device().SetWiFiStation[Static]
		return s.WithGen1Connection(ctx, address, func(conn *client.Gen1Client) error {
			if wifi.IsStatic() {
				return conn.Device().SetWiFiStationStatic(ctx, wifi.SSID, wifi.Password,
					wifi.StaticIP, wifi.Gateway, wifi.Netmask, wifi.DNS)
			}
			return conn.Device().SetWiFiStation(ctx, true, wifi.SSID, wifi.Password)
		})
	}

	// Gen2+: RPC-based WiFi.SetConfig (static or DHCP).
	if wifi.IsStatic() {
		return s.ConfigureWiFiStatic(ctx, address, wifi.SSID, wifi.Password,
			wifi.StaticIP, wifi.Netmask, wifi.Gateway, wifi.DNS)
	}
	return s.ConfigureWiFi(ctx, address, wifi.SSID, wifi.Password)
}

// reconnectCredentials determines the SSID and password for reconnecting to
// the home network after AP provisioning. If the original network differs from
// the provisioning target, returns the original SSID with an empty password
// (NetworkManager may have saved credentials). Otherwise returns the target
// WiFi credentials directly since nl80211 has no credential store.
func reconnectCredentials(originalNet *discovery.WiFiNetwork, wifi *OnboardWiFiConfig) (ssid, password string) {
	if originalNet != nil && originalNet.SSID != "" && originalNet.SSID != wifi.SSID {
		return originalNet.SSID, ""
	}
	return wifi.SSID, wifi.Password
}

// WaitForDeviceOnNetwork waits for a provisioned device to appear on the
// network by repeatedly scanning via mDNS by MAC address. Returns the device's
// new IP address or an error if not found within timeout.
func (s *Service) WaitForDeviceOnNetwork(
	ctx context.Context,
	name string,
	mac string,
	timeout time.Duration,
) (string, error) {
	deadline := time.Now().Add(timeout)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		// Try mDNS discovery by MAC
		if mac != "" {
			ip, err := s.DiscoverByMAC(ctx, mac)
			if err == nil && ip != "" {
				return ip, nil
			}
		}

		// Cancellable wait between polls so Ctrl+C is observed promptly rather
		// than after the full interval elapses.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}

	return "", fmt.Errorf("device %q not found on network within %s", name, timeout)
}

// RegisterOnboardedDevice adds a successfully onboarded device to the config registry.
// The device name is preserved as-is from discovery (typically the SSID or BLE name).
func RegisterOnboardedDevice(device *OnboardDevice, newAddress string) error {
	if newAddress == "" {
		return fmt.Errorf("no address for device %q", device.Name)
	}

	// Already registered?
	if IsDeviceRegistered(newAddress) {
		return nil
	}

	name := device.Name
	if name == "" {
		name = newAddress
	}

	// device.Model is the raw model code; route through the shared helper so
	// Type=raw-code and Model=display-name, matching every other registration
	// path (previously Model was stored empty here).
	return utils.RegisterDeviceFromModelCode(name, newAddress, device.Generation, device.Model, nil)
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

// FindByAP returns the single device whose AP SSID matches apSSID — an exact
// (case-insensitive) match, or a case-insensitive substring (so a short
// device-ID fragment from a scan diff resolves the full SSID). The bool is
// false when no device matches. Used for non-interactive single-device onboard.
func FindByAP(devices []OnboardDevice, apSSID string) (OnboardDevice, bool) {
	target := strings.ToLower(apSSID)
	if target == "" {
		// An empty target would substring-match every SSID; treat as no match.
		return OnboardDevice{}, false
	}
	for i := range devices {
		ssid := strings.ToLower(devices[i].SSID)
		if ssid != "" && (ssid == target || strings.Contains(ssid, target)) {
			return devices[i], true
		}
	}
	return OnboardDevice{}, false
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

// GetWiFiCredentials attempts to retrieve WiFi credentials from existing
// registered devices. The network is identified by SSID, not by device
// generation: every registered device reports the station SSID it is joined to
// (Gen1 via /settings, Gen2+ via WiFi.GetConfig), but only Gen1 devices return
// the key — and only when it is not masked. The results are grouped by SSID so
// that a device whose key is masked can still be onboarded using the key
// recovered from a different device on the same network. The most widely-used
// SSID with a recoverable key wins. Returns nil if no key could be recovered.
func (s *Service) GetWiFiCredentials(ctx context.Context) *OnboardWiFiConfig {
	devices := config.ListDevices()

	readings := make([]wifiReading, 0, len(devices))
	for name, dev := range devices {
		ssid, key := s.readStationWiFi(ctx, name, dev.Generation)
		readings = append(readings, wifiReading{ssid: ssid, key: key})
	}

	creds := selectWiFiNetwork(readings)
	if creds == nil {
		debug.TraceEvent("onboard: no WiFi password recoverable from %d registered device(s)", len(devices))
	}
	return creds
}

// wifiReading is one device's reported station SSID and (Gen1, unmasked) key.
type wifiReading struct {
	ssid string
	key  string
}

// selectWiFiNetwork groups readings by SSID and returns the credentials for the
// most widely-used network that has a recoverable password, breaking ties by
// SSID. A device whose key is masked still contributes its SSID, so the password
// can come from a different device on the same network. Returns nil when no
// password was recovered for any SSID.
func selectWiFiNetwork(readings []wifiReading) *OnboardWiFiConfig {
	type network struct {
		devices  int
		password string
	}
	networks := map[string]*network{}
	for _, r := range readings {
		if r.ssid == "" {
			continue
		}
		n := networks[r.ssid]
		if n == nil {
			n = &network{}
			networks[r.ssid] = n
		}
		n.devices++
		if r.key != "" && n.password == "" {
			n.password = r.key
		}
	}

	best := ""
	for ssid, n := range networks {
		if n.password == "" {
			continue
		}
		switch {
		case best == "", n.devices > networks[best].devices:
			best = ssid
		case n.devices == networks[best].devices && ssid < best:
			best = ssid
		}
	}
	if best == "" {
		return nil
	}
	return &OnboardWiFiConfig{SSID: best, Password: networks[best].password}
}

// readStationWiFi returns the station SSID a registered device is joined to,
// plus its key when the device exposes it (Gen1 only, and not when masked).
// Unreachable devices contribute nothing.
func (s *Service) readStationWiFi(ctx context.Context, name string, generation int) (ssid, key string) {
	var err error
	if generation == 1 {
		ssid, key, err = s.readGen1StationWiFi(ctx, name)
	} else {
		ssid, err = s.readGen2StationSSID(ctx, name)
	}
	if err != nil {
		debug.TraceEvent("onboard: WiFi read from %s failed: %v", name, err)
	}
	return ssid, key
}

// readGen1StationWiFi reads a Gen1 device's station SSID and key from /settings.
func (s *Service) readGen1StationWiFi(ctx context.Context, name string) (ssid, key string, err error) {
	err = s.WithGen1Connection(ctx, name, func(conn *client.Gen1Client) error {
		settings, settingsErr := conn.GetSettings(ctx)
		if settingsErr != nil {
			return settingsErr
		}
		if settings.WiFiSta != nil {
			ssid = settings.WiFiSta.SSID
			key = settings.WiFiSta.Key
		}
		return nil
	})
	return ssid, key, err
}

// readGen2StationSSID reads a Gen2+ device's station SSID (the key is write-only).
func (s *Service) readGen2StationSSID(ctx context.Context, name string) (ssid string, err error) {
	err = s.WithConnection(ctx, name, func(conn *client.Client) error {
		raw, callErr := conn.Call(ctx, "WiFi.GetConfig", nil)
		if callErr != nil {
			return callErr
		}
		ssid = provision.ExtractWiFiSSID(raw)
		return nil
	})
	return ssid, err
}

// ProvisionSource holds configuration to apply to newly provisioned devices.
// Populated from either a live device backup or a saved template.
type ProvisionSource struct {
	Backup   *backup.DeviceBackup   // From --from-device (Gen1+Gen2)
	Template *config.DeviceTemplate // From --from-template (Gen2+ only)
	WiFi     *OnboardWiFiConfig     // Extracted WiFi creds (if available)
}

// LoadProvisionSource loads device configuration from a source device or template.
// For --from-device: creates a backup and extracts WiFi credentials.
// For --from-template: loads the saved template from config.
func (s *Service) LoadProvisionSource(ctx context.Context, fromDevice, fromTemplate string) (*ProvisionSource, error) {
	source := &ProvisionSource{}

	switch {
	case fromDevice != "":
		bkp, err := s.CreateBackup(ctx, fromDevice, backup.Options{})
		if err != nil {
			return nil, fmt.Errorf("failed to backup source device %q: %w", fromDevice, err)
		}
		source.Backup = bkp

		// Extract WiFi credentials from backup (Gen1 devices include the password)
		source.WiFi = extractWiFiFromBackup(bkp)

	case fromTemplate != "":
		tpl, ok := config.GetDeviceTemplate(fromTemplate)
		if !ok {
			return nil, fmt.Errorf("template %q not found", fromTemplate)
		}
		source.Template = &tpl
	}

	return source, nil
}

// extractWiFiFromBackup extracts WiFi credentials from a device backup.
// Prefers bkp.WiFi (always populated during backup) over parsing bkp.Config.
// Gen1 backups include the WiFi password ("key"); Gen2+ include only the SSID.
func extractWiFiFromBackup(bkp *backup.DeviceBackup) *OnboardWiFiConfig {
	// Primary: use the dedicated WiFi blob (populated by marshalGen1WiFi / Gen2 export).
	if cfg := extractWiFiFromBlob(bkp.WiFi); cfg != nil {
		return cfg
	}

	// Fallback for Gen1: parse the full settings from Config.
	if bkp.DeviceInfo != nil && bkp.DeviceInfo.Generation == 1 && bkp.Config != nil {
		var settings gen1.Settings
		if err := json.Unmarshal(bkp.Config, &settings); err != nil {
			debug.TraceEvent("extractWiFiFromBackup: failed to parse Gen1 settings: %v", err)
			return nil
		}
		if settings.WiFiSta != nil && settings.WiFiSta.SSID != "" && settings.WiFiSta.Key != "" {
			return &OnboardWiFiConfig{
				SSID:     settings.WiFiSta.SSID,
				Password: settings.WiFiSta.Key,
			}
		}
	}

	return nil
}

// extractWiFiFromBlob parses the backup WiFi blob for station SSID and password.
// Gen1 blobs include "key" (password); Gen2+ blobs include only "ssid".
func extractWiFiFromBlob(data json.RawMessage) *OnboardWiFiConfig {
	if data == nil {
		return nil
	}
	var wifiData map[string]any
	if err := json.Unmarshal(data, &wifiData); err != nil {
		return nil
	}
	sta, ok := wifiData[fieldSTA].(map[string]any)
	if !ok {
		return nil
	}
	ssid, ok := sta["ssid"].(string)
	if !ok || ssid == "" {
		return nil
	}
	cfg := &OnboardWiFiConfig{SSID: ssid}
	if key, ok := sta["key"].(string); ok && key != "" {
		cfg.Password = key
	}
	return cfg
}

// ApplyProvisionSource applies a previously loaded provision source to a newly
// provisioned device at the given address. Skips network config since WiFi
// was already configured during provisioning.
func (s *Service) ApplyProvisionSource(ctx context.Context, deviceAddr string, source *ProvisionSource) error {
	switch {
	case source.Backup != nil:
		_, err := s.RestoreBackup(ctx, deviceAddr, source.Backup, backup.RestoreOptions{
			SkipNetwork: true,
		})
		return err

	case source.Template != nil:
		_, err := s.ApplyTemplate(ctx, deviceAddr, source.Template.Config, false)
		return err
	}

	return nil
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
