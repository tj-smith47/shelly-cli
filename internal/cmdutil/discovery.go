// Package cmdutil provides discovery orchestration utilities.
package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultScanTimeout is the default timeout for HTTP subnet scanning.
const DefaultScanTimeout = 2 * time.Minute

// ErrDiscoveryHandled signals that a discovery helper already reported the
// failure to the user (e.g. printed a "BLE not available" message). Callers
// should treat it as a clean stop and must NOT print it again, avoiding a
// contradictory generic "no devices found" summary.
var ErrDiscoveryHandled = errors.New("discovery already reported")

// DiscoveryOptions holds options for discovery orchestration functions.
type DiscoveryOptions struct {
	Factory      *Factory
	Platform     string
	Subnets      []string
	Register     bool
	SkipExisting bool
	AllNetworks  bool
	Timeout      time.Duration
}

// ResolveSubnets determines which subnets to scan based on explicit flags
// and auto-detection. When no subnets are specified explicitly:
//   - Detects all local subnets via utils.DetectSubnets()
//   - If allNetworks is true, non-TTY, or only one subnet found: returns all
//   - Otherwise presents an interactive multi-select prompt
func ResolveSubnets(ios *iostreams.IOStreams, explicit []string, allNetworks bool) ([]string, error) {
	if len(explicit) > 0 {
		for _, s := range explicit {
			if _, _, err := net.ParseCIDR(s); err != nil {
				return nil, fmt.Errorf("invalid subnet %q: %w", s, err)
			}
		}
		return explicit, nil
	}

	subnets, err := utils.DetectSubnets()
	if err != nil {
		return nil, fmt.Errorf("failed to detect subnets: %w", err)
	}

	// No prompt needed: single subnet, non-TTY, or --all-networks
	if allNetworks || len(subnets) == 1 || !ios.IsStdinTTY() {
		for _, s := range subnets {
			ios.Info("Detected subnet: %s", s)
		}
		return subnets, nil
	}

	// Interactive multi-select (all pre-selected by default)
	selected, selectErr := ios.MultiSelect("Select networks to scan:", subnets, subnets)
	if selectErr != nil {
		// Fallback to all detected subnets on prompt error (e.g., non-TTY piped input)
		return subnets, nil //nolint:nilerr // intentional fallback to all subnets
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no subnets selected")
	}
	return selected, nil
}

// DedupDiscoveredDevices removes duplicate devices from combined multi-method
// discovery results. It mirrors the SDK's identity precedence (ID, then MAC
// address, then IP address) so the same physical device found via two methods
// collapses to one entry. Address-less results (e.g. BLE/mDNS devices with a
// nil IP) that also lack an ID and MAC are kept individually rather than
// collapsing onto the shared net.IP(nil).String() == "<nil>" key, which would
// otherwise drop every address-less device but the first.
func DedupDiscoveredDevices(devices []discovery.DiscoveredDevice) []discovery.DiscoveredDevice {
	seen := make(map[string]bool, len(devices))
	result := make([]discovery.DiscoveredDevice, 0, len(devices))

	for _, d := range devices {
		key := d.ID
		if key == "" {
			key = d.MACAddress
		}
		if key == "" {
			// No stable identity: keep the device rather than collapsing all
			// address-less results onto the literal "<nil>" address key.
			if len(d.Address) == 0 {
				result = append(result, d)
				continue
			}
			key = d.Address.String()
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, d)
		}
	}

	return result
}

// generateAddresses generates probe addresses for all given subnets.
func generateAddresses(subnets []string) ([]string, error) {
	var allAddresses []string
	for _, subnet := range subnets {
		if _, _, err := net.ParseCIDR(subnet); err != nil {
			return nil, fmt.Errorf("invalid subnet %q: %w", subnet, err)
		}
		addrs := discovery.GenerateSubnetAddresses(subnet)
		allAddresses = append(allAddresses, addrs...)
	}
	if len(allAddresses) == 0 {
		return nil, fmt.Errorf("no addresses to scan in %s", strings.Join(subnets, ", "))
	}
	return allAddresses, nil
}

// CacheAndRegisterDevices persists the discovered device addresses to the
// completion cache and, when register is true, registers them with the local
// registry. It returns the number of devices newly registered (0 when register
// is false) so the caller can emit a single combined ios.Added summary that
// also accounts for plugin-discovered devices. It does not print the summary
// itself, keeping that decision with the caller.
func CacheAndRegisterDevices(ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice, register, skipExisting bool) int {
	addrs := make([]string, 0, len(devices))
	for _, d := range devices {
		addrs = append(addrs, d.Address.String())
	}
	if err := completion.SaveDiscoveryCache(addrs); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if !register {
		return 0
	}

	added, regErr := utils.RegisterDiscoveredDevices(devices, skipExisting)
	if regErr != nil {
		ios.Warning("Registration error: %v", regErr)
	}
	return added
}

// RunPluginOnlyDiscovery runs discovery for a specific platform only.
// Uses shelly.RunPluginPlatformDiscoveryWithProgress for the core logic.
func RunPluginOnlyDiscovery(ctx context.Context, opts *DiscoveryOptions) error {
	ios := opts.Factory.IOStreams()

	// Get plugin registry
	registry, err := plugins.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize plugin registry: %w", err)
	}

	// Verify plugin exists for platform
	plugin, err := registry.FindByPlatform(opts.Platform)
	if err != nil {
		return fmt.Errorf("failed to find plugin: %w", err)
	}
	if plugin == nil {
		return fmt.Errorf("no plugin found for platform %q (is shelly-%s installed?)", opts.Platform, opts.Platform)
	}

	// Resolve subnets
	subnets, resolveErr := ResolveSubnets(ios, opts.Subnets, opts.AllNetworks)
	if resolveErr != nil {
		return resolveErr
	}

	addresses, addrErr := generateAddresses(subnets)
	if addrErr != nil {
		return addrErr
	}

	subnetLabel := strings.Join(subnets, ", ")
	ios.Info("Scanning %d addresses for %s devices across %s...", len(addresses), opts.Platform, subnetLabel)

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	mw.AddLine("scan", fmt.Sprintf("0/%d addresses probed", len(addresses)))

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Scan every resolved subnet so plugin discovery covers the same address
	// space as native Shelly discovery when multiple --subnet values are given.
	pluginDevices := shelly.RunPluginPlatformDiscoveryWithProgress(ctx, registry, opts.Platform, subnets, shelly.IsDeviceRegistered, func(p shelly.DiscoveryProgress) bool {
		if p.Found && p.Device != nil {
			mw.UpdateLine("scan", iostreams.StatusRunning,
				fmt.Sprintf("%d/%d - found %s (%s)", p.Done, p.Total, p.Device.Name, p.Device.Model))
		} else {
			mw.UpdateLine("scan", iostreams.StatusRunning,
				fmt.Sprintf("%d/%d addresses probed", p.Done, p.Total))
		}
		return ctx.Err() == nil
	})

	mw.UpdateLine("scan", iostreams.StatusSuccess,
		fmt.Sprintf("%d/%d addresses probed, %d %s devices found",
			len(addresses), len(addresses), len(pluginDevices), opts.Platform))
	mw.Finalize()

	if len(pluginDevices) == 0 {
		ios.NoResults(opts.Platform+" devices",
			fmt.Sprintf("Ensure %s devices are powered on and accessible in %s", opts.Platform, subnetLabel))
		return nil
	}

	// Convert to term display type
	termDevices := term.ConvertPluginDevices(pluginDevices)
	term.DisplayPluginDiscoveredDevices(ios, termDevices)

	if opts.Register {
		added := term.RegisterPluginDevices(termDevices, opts.SkipExisting)
		ios.Added("device", added)
	}

	return nil
}

// RunPluginDetection runs plugin detection on addresses that Shelly detection may have missed.
// Uses shelly.RunPluginDetection for the core logic.
func RunPluginDetection(ctx context.Context, ios *iostreams.IOStreams, subnets []string) []term.PluginDiscoveredDevice {
	// Get plugin registry
	registry, err := plugins.NewRegistry()
	if err != nil {
		ios.DebugErr("initializing plugin registry", err)
		return nil
	}

	// Get detection-capable plugins
	capablePlugins, err := registry.ListDetectionCapable()
	if err != nil {
		ios.DebugErr("listing detection-capable plugins", err)
		return nil
	}

	if len(capablePlugins) == 0 {
		return nil // No detection-capable plugins installed
	}

	// Auto-detect if no subnets provided
	if len(subnets) == 0 {
		var detectErr error
		subnets, detectErr = utils.DetectSubnets()
		if detectErr != nil {
			return nil
		}
	}

	ios.Info("Checking for plugin-managed devices...")

	// Use shelly service layer for plugin detection with short timeout
	pluginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Scan every resolved subnet so plugin detection matches the address space
	// covered by native Shelly discovery for the same invocation.
	pluginDevices := shelly.RunPluginDetection(pluginCtx, registry, subnets, shelly.IsDeviceRegistered)

	return term.ConvertPluginDevices(pluginDevices)
}

// RunHTTPDiscovery performs HTTP subnet scanning discovery across one or more subnets.
func RunHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnets []string) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = DefaultScanTimeout
	}

	addresses, err := generateAddresses(subnets)
	if err != nil {
		return nil, err
	}

	subnetLabel := strings.Join(subnets, ", ")
	ios.Info("Scanning %d addresses in %s...", len(addresses), subnetLabel)

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	mw.AddLine("scan", fmt.Sprintf("0/%d addresses probed", len(addresses)))

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Use progress callback to update MultiWriter
	devices := discovery.ProbeAddressesWithProgress(ctx, addresses, func(p discovery.ProbeProgress) bool {
		status := iostreams.StatusRunning
		msg := fmt.Sprintf("%d/%d addresses probed", p.Done, p.Total)
		if p.Found && p.Device != nil {
			msg = fmt.Sprintf("%d/%d - found %s (%s)", p.Done, p.Total, p.Device.Name, p.Device.Model)
		}
		mw.UpdateLine("scan", status, msg)
		// Cancellation is deadline/ctx-driven: the SDK rechecks ctx.Err() in
		// every in-flight goroutine, so mirroring ctx.Err() here is enough to
		// stop the scan. The callback's bool return is purely that ctx mirror.
		return ctx.Err() == nil
	})

	// Mark scan complete
	mw.UpdateLine("scan", iostreams.StatusSuccess, fmt.Sprintf("%d/%d addresses probed, %d devices found",
		len(addresses), len(addresses), len(devices)))
	mw.Finalize()

	return devices, nil
}

// RunMDNSDiscovery performs mDNS/Zeroconf discovery.
func RunMDNSDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := RunWithSpinner(ctx, ios, "Discovering devices via mDNS...", func(ctx context.Context) error {
		var discoverErr error
		devices, discoverErr = shelly.DiscoverMDNS(ctx, timeout)
		return discoverErr
	})

	return devices, err
}

// RunCoIoTDiscovery performs CoIoT/CoAP discovery.
func RunCoIoTDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := RunWithSpinner(ctx, ios, "Discovering devices via CoIoT...", func(ctx context.Context) error {
		var discoverErr error
		devices, discoverErr = shelly.DiscoverCoIoT(ctx, timeout)
		return discoverErr
	})

	return devices, err
}

// RunBLEDiscovery performs Bluetooth Low Energy discovery.
func RunBLEDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := RunWithSpinner(ctx, ios, "Discovering devices via BLE...", func(ctx context.Context) error {
		var discoverErr error
		devices, discoverErr = shelly.DiscoverBLEContext(ctx, timeout)
		return discoverErr
	})
	if err != nil && shelly.IsBLEUnavailable(err) {
		// BLE-unavailable is a configuration condition, not a scan failure:
		// report it once with actionable hints, log the real cause for debug,
		// then signal a handled stop so the caller skips the contradictory
		// generic "no devices found" summary.
		ios.Error("BLE discovery is not available on this system")
		ios.Hint("Ensure you have a Bluetooth adapter and it is enabled")
		ios.Hint("On Linux, you may need to run with elevated privileges")
		ios.DebugErr("BLE discovery init", err)
		return nil, ErrDiscoveryHandled
	}

	return devices, err
}

// SelectWiFiNetwork scans device for nearby WiFi networks, presents the
// deduplicated list (strongest signal first) via an interactive picker, and
// returns the chosen SSID. It errors when the scan fails, no networks are
// found, or the selection cannot be resolved.
func SelectWiFiNetwork(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string) (string, error) {
	ios.Info("Scanning for networks...")

	results, err := svc.ScanWiFi(ctx, device)
	if err != nil {
		return "", fmt.Errorf("network scan failed: %w", err)
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no networks found")
	}

	networks := shelly.DedupeWiFiNetworks(results)

	options := make([]string, len(networks))
	for i, n := range networks {
		signal := output.FormatWiFiSignalStrength(n.RSSI)
		options[i] = fmt.Sprintf("%s (%s, ch %d)", n.SSID, signal, n.Channel)
	}

	selected, err := ios.Select("Select WiFi network:", options, 0)
	if err != nil {
		return "", fmt.Errorf("network selection failed: %w", err)
	}

	// Match the chosen option string back to its network to recover the SSID.
	for i, opt := range options {
		if opt == selected {
			return networks[i].SSID, nil
		}
	}

	return "", fmt.Errorf("selected network not found")
}
