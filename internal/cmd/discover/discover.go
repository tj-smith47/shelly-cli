// Package discover provides device discovery commands.
package discover

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/httpscan"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/mdns"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultScanTimeout is the default timeout for HTTP subnet scanning.
const DefaultScanTimeout = 2 * time.Minute

// Options holds all discovery options.
type Options struct {
	Factory      *cmdutil.Factory
	Method       string
	Platform     string
	Register     bool
	SkipExisting bool
	SkipPlugins  bool
	Subnet       string
	Timeout      time.Duration
}

// NewCommand creates the discover command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "discover",
		Aliases: []string{"disc", "find"},
		Short:   "Discover Shelly devices on the network",
		Long: `Discover Shelly devices using various protocols.

By default, uses HTTP subnet scanning which works reliably even when
multicast is blocked. Automatically detects the local subnet.

Available discovery methods (--method):
  http   - HTTP subnet scanning (default, works everywhere)
  mdns   - mDNS/Zeroconf discovery (Gen2+ devices)
  ble    - Bluetooth Low Energy discovery (provisioning mode)
  coiot  - CoIoT/CoAP discovery (Gen1 devices)

Plugin-managed devices (e.g., Tasmota, ESPHome) can also be discovered
if the corresponding plugin is installed. Use --skip-plugins to disable
plugin detection, or --platform to filter by specific platform.`,
		Example: `  # Discover devices via HTTP scan (default, auto-detects subnet)
  shelly discover

  # Specify subnet for HTTP scan
  shelly discover --subnet 192.168.1.0/24

  # Use mDNS instead of HTTP scan
  shelly discover --method mdns

  # Use BLE discovery
  shelly discover --method ble

  # Auto-register discovered devices
  shelly discover --register

  # Skip plugin detection (Shelly-only)
  shelly discover --skip-plugins

  # Discover only Tasmota devices
  shelly discover --platform tasmota`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	// Add flags for the parent discover command
	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", DefaultScanTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&opts.Register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&opts.SkipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().StringVar(&opts.Subnet, "subnet", "", "Subnet to scan (auto-detected if not specified)")
	cmd.Flags().StringVarP(&opts.Method, "method", "m", "http", "Discovery method: http, mdns, ble, coiot")
	cmd.Flags().BoolVar(&opts.SkipPlugins, "skip-plugins", false, "Skip plugin detection (Shelly-only discovery)")
	cmd.Flags().StringVarP(&opts.Platform, "platform", "p", "", "Only discover devices of this platform (e.g., tasmota)")

	// Add subcommands (kept for direct access)
	cmd.AddCommand(mdns.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))
	cmd.AddCommand(coiot.NewCommand(f))
	cmd.AddCommand(httpscan.NewCommand(f))

	return cmd
}

// run runs device discovery using the selected method.
//
//nolint:gocyclo // Complexity from handling multiple discovery methods and plugin integration
func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check for demo mode
	if mock.IsDemoMode() && mock.HasDiscoveryFixtures() {
		return runDemoMode(opts)
	}

	// If platform is specified (and not "shelly"), skip native Shelly discovery
	// and only run plugin detection for that platform
	if opts.Platform != "" && opts.Platform != model.PlatformShelly {
		return runPluginOnlyDiscovery(ctx, opts)
	}

	var shellyDevices []discovery.DiscoveredDevice
	var err error

	switch opts.Method {
	case "http", "scan", "":
		shellyDevices, err = runHTTPDiscovery(ctx, ios, opts.Timeout, opts.Subnet)
	case "mdns":
		shellyDevices, err = runMDNSDiscovery(ctx, ios, opts.Timeout)
	case "coiot":
		shellyDevices, err = runCoIoTDiscovery(ctx, ios, opts.Timeout)
	case "ble":
		shellyDevices, err = runBLEDiscovery(ctx, ios, opts.Timeout)
	default:
		return fmt.Errorf("unknown discovery method: %s (valid: http, mdns, ble, coiot)", opts.Method)
	}

	if err != nil {
		return err
	}

	// Run plugin detection if not skipped
	var pluginDevices []term.PluginDiscoveredDevice
	if !opts.SkipPlugins && opts.Method == "http" {
		pluginDevices = runPluginDetection(ctx, ios, opts.Subnet)
	}

	// Display results
	totalDevices := len(shellyDevices) + len(pluginDevices)
	if totalDevices == 0 {
		ios.NoResults("devices", "Ensure devices are powered on and accessible on the network")
		return nil
	}

	// Display Shelly devices
	if len(shellyDevices) > 0 {
		term.DisplayDiscoveredDevices(ios, shellyDevices)
	}

	// Display plugin-discovered devices
	if len(pluginDevices) > 0 {
		term.DisplayPluginDiscoveredDevices(ios, pluginDevices)
	}

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, totalDevices)
	for _, d := range shellyDevices {
		addresses = append(addresses, d.Address.String())
	}
	for _, d := range pluginDevices {
		if d.Address != "" {
			addresses = append(addresses, d.Address)
		}
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	// Register devices if requested
	if opts.Register {
		addedShelly, regErr := utils.RegisterDiscoveredDevices(shellyDevices, opts.SkipExisting)
		if regErr != nil {
			ios.Warning("Registration error: %v", regErr)
		}

		addedPlugin := registerPluginDevices(pluginDevices, opts.SkipExisting)
		totalAdded := addedShelly + addedPlugin
		ios.Added("device", totalAdded)
	}

	return nil
}

// runPluginOnlyDiscovery runs discovery for a specific platform only.
// Uses shelly.RunPluginPlatformDiscoveryWithProgress for the core logic.
func runPluginOnlyDiscovery(ctx context.Context, opts *Options) error {
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

	// Get or detect subnet
	subnet := opts.Subnet
	if subnet == "" {
		var detectErr error
		subnet, detectErr = utils.DetectSubnet()
		if detectErr != nil {
			return fmt.Errorf("failed to detect subnet: %w", detectErr)
		}
		ios.Info("Detected subnet: %s", subnet)
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return fmt.Errorf("no addresses to scan in subnet %s", subnet)
	}

	ios.Info("Scanning %d addresses for %s devices...", len(addresses), opts.Platform)

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	mw.AddLine("scan", fmt.Sprintf("0/%d addresses probed", len(addresses)))

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Use shelly service layer for plugin discovery with progress callback
	pluginDevices := shelly.RunPluginPlatformDiscoveryWithProgress(ctx, registry, opts.Platform, subnet, isDeviceRegistered, func(p shelly.DiscoveryProgress) bool {
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
			fmt.Sprintf("Ensure %s devices are powered on and accessible in %s", opts.Platform, ipNet))
		return nil
	}

	// Convert to term display type
	termDevices := convertPluginDevices(pluginDevices)
	term.DisplayPluginDiscoveredDevices(ios, termDevices)

	if opts.Register {
		added := registerPluginDevices(termDevices, opts.SkipExisting)
		ios.Added("device", added)
	}

	return nil
}

// runPluginDetection runs plugin detection on addresses that Shelly detection may have missed.
// Uses shelly.RunPluginDetection for the core logic.
func runPluginDetection(ctx context.Context, ios *iostreams.IOStreams, subnet string) []term.PluginDiscoveredDevice {
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

	// Get addresses to probe for plugin detection
	if subnet == "" {
		var detectErr error
		subnet, detectErr = utils.DetectSubnet()
		if detectErr != nil {
			return nil
		}
	}

	ios.Info("Checking for plugin-managed devices...")

	// Use shelly service layer for plugin detection with short timeout
	pluginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pluginDevices := shelly.RunPluginDetection(pluginCtx, registry, subnet, isDeviceRegistered)

	return convertPluginDevices(pluginDevices)
}

// convertPluginDevices converts shelly plugin devices to term display types.
func convertPluginDevices(devices []shelly.PluginDiscoveredDevice) []term.PluginDiscoveredDevice {
	result := make([]term.PluginDiscoveredDevice, len(devices))
	for i, d := range devices {
		components := make([]term.PluginComponentInfo, len(d.Components))
		for j, c := range d.Components {
			components[j] = term.PluginComponentInfo{
				Type: c.Type,
				ID:   c.ID,
				Name: c.Name,
			}
		}
		result[i] = term.PluginDiscoveredDevice{
			ID:         d.ID,
			Name:       d.Name,
			Model:      d.Model,
			Address:    d.Address.String(),
			Platform:   d.Platform,
			Firmware:   d.Firmware,
			Components: components,
		}
	}
	return result
}

// isDeviceRegistered checks if an address is already in the device registry.
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

// registerPluginDevices registers plugin-detected devices.
func registerPluginDevices(devices []term.PluginDiscoveredDevice, skipExisting bool) int {
	pluginDevices := make([]utils.PluginDevice, len(devices))
	for i, d := range devices {
		pluginDevices[i] = utils.PluginDevice{
			Address:  d.Address,
			Platform: d.Platform,
			ID:       d.ID,
			Name:     d.Name,
			Model:    d.Model,
			Firmware: d.Firmware,
		}
	}
	return utils.RegisterPluginDiscoveredDevices(pluginDevices, skipExisting)
}

// runDemoMode returns mock discovery results from fixtures.
func runDemoMode(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Get mock discovered devices
	mockDevices := mock.GetDiscoveredDevices()

	// Convert to discovery.DiscoveredDevice for display compatibility
	shellyDevices := make([]discovery.DiscoveredDevice, len(mockDevices))
	for i, d := range mockDevices {
		shellyDevices[i] = discovery.DiscoveredDevice{
			ID:         d.ID,
			Name:       d.Name,
			Model:      d.Model,
			Address:    d.Address,
			MACAddress: d.MACAddress,
			Generation: types.Generation(d.Generation),
			Protocol:   discovery.ProtocolManual, // Demo devices use "manual" protocol
		}
	}

	if len(shellyDevices) == 0 {
		ios.NoResults("devices", "No discovery fixtures defined in demo mode")
		return nil
	}

	ios.Success("Discovered %d device(s) (demo mode)", len(shellyDevices))
	ios.Println()
	term.DisplayDiscoveredDevices(ios, shellyDevices)

	// Save to completion cache
	addresses := make([]string, len(shellyDevices))
	for i, d := range shellyDevices {
		addresses[i] = d.Address.String()
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	// Register devices if requested
	if opts.Register {
		added, err := utils.RegisterDiscoveredDevices(shellyDevices, opts.SkipExisting)
		if err != nil {
			ios.Warning("Registration error: %v", err)
		}
		ios.Added("device", added)
	}

	return nil
}

// runHTTPDiscovery performs HTTP subnet scanning discovery.
func runHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = DefaultScanTimeout
	}

	if subnet == "" {
		var err error
		subnet, err = utils.DetectSubnet()
		if err != nil {
			return nil, fmt.Errorf("failed to detect subnet: %w", err)
		}
		ios.Info("Detected subnet: %s", subnet)
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	// Generate addresses to probe
	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses to scan in subnet %s", subnet)
	}

	ios.Info("Scanning %d addresses in %s...", len(addresses), ipNet)

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
		return ctx.Err() == nil // Continue unless context canceled
	})

	// Mark scan complete
	mw.UpdateLine("scan", iostreams.StatusSuccess, fmt.Sprintf("%d/%d addresses probed, %d devices found",
		len(addresses), len(addresses), len(devices)))
	mw.Finalize()

	return devices, nil
}

// runMDNSDiscovery performs mDNS/Zeroconf discovery.
func runMDNSDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via mDNS...", func(_ context.Context) error {
		var discoverErr error
		var cleanup func()
		devices, cleanup, discoverErr = shelly.DiscoverMDNS(timeout)
		defer cleanup()
		return discoverErr
	})

	return devices, err
}

// runCoIoTDiscovery performs CoIoT/CoAP discovery.
func runCoIoTDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via CoIoT...", func(_ context.Context) error {
		var discoverErr error
		var cleanup func()
		devices, cleanup, discoverErr = shelly.DiscoverCoIoT(timeout)
		defer cleanup()
		return discoverErr
	})

	return devices, err
}

// runBLEDiscovery performs Bluetooth Low Energy discovery.
func runBLEDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	bleDiscoverer, cleanup, err := shelly.DiscoverBLE()
	if err != nil {
		ios.Error("BLE discovery is not available on this system")
		ios.Hint("Ensure you have a Bluetooth adapter and it is enabled")
		ios.Hint("On Linux, you may need to run with elevated privileges")
		return nil, nil
	}
	defer cleanup()

	var devices []discovery.DiscoveredDevice
	err = cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via BLE...", func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		var discoverErr error
		devices, discoverErr = bleDiscoverer.DiscoverWithContext(ctx)
		return discoverErr
	})

	return devices, err
}
