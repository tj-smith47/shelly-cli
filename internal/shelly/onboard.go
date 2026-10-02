package shelly

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/provisioning"
	"github.com/tj-smith47/shelly-go/reprovision"

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
	// Open joins a network that has no password. Without it an empty Password
	// means the password is not known, and it is looked up in this host's stored
	// credentials for SSID.
	Open bool
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
	wifiDisc := discovery.NewWiFiDiscovererWithScanner(s.scanner())
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
			// A cancelled search still reports the devices seen before it ended.
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

	wifi, err := s.joinPassword(ctx, wifi)
	if err != nil {
		result.Error = err
		return result
	}

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
		bleWiFi.StaticIP = ipv4Static
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
	// One host lookup serves every device rather than one per device.
	wifiCfg, err := s.joinPassword(ctx, wifiCfg)
	if err != nil {
		for i, dev := range devices {
			results[i] = &OnboardResult{Device: dev, Method: string(OnboardSourceBLE), Error: err}
		}
		return results
	}
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

// HostWiFiCredentials recovers the SSID and passphrase of the network the host
// running the CLI is currently joined to. AP-hop provisioning (--ap-only) runs
// from a machine already on the target WiFi, so the host already holds both the
// SSID and the passphrase — no Shelly device surrenders its station key, and no
// previously-registered device is required. This mirrors the host-credential
// recovery that restore --to-ap performs, so the two AP-hop flows resolve
// credentials identically. Returns nil when the host is not on WiFi or the
// platform cannot recover the passphrase.
func (s *Service) HostWiFiCredentials(ctx context.Context) *OnboardWiFiConfig {
	ssid := s.hostCurrentSSID(ctx)
	if ssid == "" {
		return nil
	}
	pass, passErr := s.HostWiFiPassword(ctx, ssid)
	if passErr != nil || pass == "" {
		debug.TraceEvent("onboard: host passphrase for %q not recovered: %v", ssid, passErr)
		return nil
	}
	return &OnboardWiFiConfig{SSID: ssid, Password: pass}
}

// hostCurrentSSID returns the SSID of the WiFi network this host is on, or ""
// when it is on none or the platform cannot tell. It only reads the host's WiFi
// state.
func (s *Service) hostCurrentSSID(ctx context.Context) string {
	scanner := s.scanner()
	if scanner == nil {
		return ""
	}
	current, err := scanner.CurrentNetwork(ctx)
	if err != nil || current == nil || current.SSID == "" {
		debug.TraceEvent("host not on a WiFi network: %v", err)
		return ""
	}
	return current.SSID
}

// HostWiFiPassword recovers the passphrase this host has stored for ssid from the
// OS credential store, when the WiFi backend supports it. No Shelly device
// returns its station key, so this lets a device join a network the host knows
// without the passphrase being typed again.
func (s *Service) HostWiFiPassword(ctx context.Context, ssid string) (string, error) {
	if ssid == "" {
		return "", fmt.Errorf("no WiFi network named")
	}
	scanner := s.scanner()
	if scanner == nil {
		return "", fmt.Errorf("WiFi not supported on this platform")
	}
	provider, ok := scanner.(discovery.HostNetworkPasswordProvider)
	if !ok {
		return "", fmt.Errorf("host passphrase recovery not supported on this platform")
	}
	return provider.HostNetworkPassword(ctx, ssid)
}

// hostPassphrase returns the passphrase this host has stored for ssid, or
// PassphraseError naming ssid when there is none. Why the lookup failed goes to
// the debug trace, since the passphrase error is what the user acts on.
func (s *Service) hostPassphrase(ctx context.Context, ssid string) (string, error) {
	pass, err := s.HostWiFiPassword(ctx, ssid)
	if err != nil || pass == "" {
		debug.TraceEvent("host passphrase for %q not recovered: %v", ssid, err)
		return "", PassphraseError(ssid)
	}
	return pass, nil
}

// scanner returns the WiFi backend set with WithWiFiScanner, or the platform's.
func (s *Service) scanner() discovery.WiFiScanner {
	if s.wifiScanner != nil {
		return s.wifiScanner
	}
	// A test binary never reaches the host's WiFi, whichever way its service
	// was built.
	if testing.Testing() {
		return OfflineWiFiScanner{}
	}
	return discovery.NewWiFiDiscoverer().Scanner
}

// errNoWiFi is returned by every OfflineWiFiScanner call.
var errNoWiFi = errors.New("WiFi is not available")

// OfflineWiFiScanner is a WiFi backend with no network: it reports no current
// network, knows no stored passphrase and refuses every scan and connect, so a
// service using it never queries or moves the host's WiFi. Demo mode and tests
// use it.
type OfflineWiFiScanner struct{}

// Scan refuses to scan.
func (OfflineWiFiScanner) Scan(context.Context) ([]discovery.WiFiNetwork, error) {
	return nil, errNoWiFi
}

// Connect refuses to connect.
func (OfflineWiFiScanner) Connect(context.Context, string, string) error { return errNoWiFi }

// Disconnect refuses to disconnect.
func (OfflineWiFiScanner) Disconnect(context.Context) error { return errNoWiFi }

// CurrentNetwork reports that the host is on no WiFi network.
func (OfflineWiFiScanner) CurrentNetwork(context.Context) (*discovery.WiFiNetwork, error) {
	return nil, errNoWiFi
}

// joinPassword returns wifi with its passphrase filled in from this host's
// stored credentials when it has none and the network is not open. A network
// with no passphrase from either is refused with PassphraseError, so a secured
// network is never written as an open one.
func (s *Service) joinPassword(ctx context.Context, wifi *OnboardWiFiConfig) (*OnboardWiFiConfig, error) {
	if wifi.Open || wifi.Password != "" {
		return wifi, nil
	}
	pass, err := s.hostPassphrase(ctx, wifi.SSID)
	if err != nil {
		return nil, err
	}
	resolved := *wifi
	resolved.Password = pass
	return &resolved, nil
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

		source.WiFi = provisionWiFi(bkp)

	case fromTemplate != "":
		tpl, ok := config.GetDeviceTemplate(fromTemplate)
		if !ok {
			return nil, fmt.Errorf("template %q not found", fromTemplate)
		}
		source.Template = &tpl
	}

	return source, nil
}

// provisionWiFi returns the network a source device's backup records, for a
// new device to join, or nil when the backup records none. An open station
// stays open; a secured one comes with its key only when the backup holds it.
// The static address is left out because it belongs to the source device.
func provisionWiFi(bkp *backup.DeviceBackup) *OnboardWiFiConfig {
	n := reprovision.NetworkFromBackup(bkp.Backup)
	if n.SSID == "" {
		return nil
	}
	return &OnboardWiFiConfig{SSID: n.SSID, Password: n.Password, Open: n.Open}
}

// ApplyProvisionSource applies a previously loaded provision source to a newly
// provisioned device at the given address. Skips network config since WiFi
// was already configured during provisioning.
func (s *Service) ApplyProvisionSource(ctx context.Context, deviceAddr string, source *ProvisionSource) error {
	switch {
	case source.Backup != nil:
		result, err := s.RestoreBackup(ctx, deviceAddr, source.Backup, backup.RestoreOptions{
			SkipNetwork: true,
		})
		if err != nil {
			return err
		}
		// A restore can reject sections while reporting no top-level error;
		// provisioning must not report success when the device refused config.
		if restoreErr := result.Err(); restoreErr != nil {
			return fmt.Errorf("apply backup during provisioning: %w", restoreErr)
		}
		return nil

	case source.Template != nil:
		_, _, err := s.applyTemplate(ctx, deviceAddr, source.Template.Config, false, stationsOmit)
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
