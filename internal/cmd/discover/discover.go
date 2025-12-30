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
	timeout      time.Duration
	register     bool
	skipExisting bool
	subnet       string
	method       string
	skipPlugins  bool
	platform     string
}

// NewCommand creates the discover command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, opts)
		},
	}

	// Add flags for the parent discover command
	cmd.Flags().DurationVarP(&opts.timeout, "timeout", "t", DefaultScanTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&opts.register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&opts.skipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().StringVar(&opts.subnet, "subnet", "", "Subnet to scan (auto-detected if not specified)")
	cmd.Flags().StringVarP(&opts.method, "method", "m", "http", "Discovery method: http, mdns, ble, coiot")
	cmd.Flags().BoolVar(&opts.skipPlugins, "skip-plugins", false, "Skip plugin detection (Shelly-only discovery)")
	cmd.Flags().StringVarP(&opts.platform, "platform", "p", "", "Only discover devices of this platform (e.g., tasmota)")

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
func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

	// Check for demo mode
	if mock.IsDemoMode() && mock.HasDiscoveryFixtures() {
		return runDemoMode(f, opts)
	}

	// If platform is specified (and not "shelly"), skip native Shelly discovery
	// and only run plugin detection for that platform
	if opts.platform != "" && opts.platform != model.PlatformShelly {
		return runPluginOnlyDiscovery(ctx, f, ios, opts)
	}

	var shellyDevices []discovery.DiscoveredDevice
	var err error

	switch opts.method {
	case "http", "scan", "":
		shellyDevices, err = runHTTPDiscovery(ctx, ios, opts.timeout, opts.subnet)
	case "mdns":
		shellyDevices, err = runMDNSDiscovery(ctx, ios, opts.timeout)
	case "coiot":
		shellyDevices, err = runCoIoTDiscovery(ctx, ios, opts.timeout)
	case "ble":
		shellyDevices, err = runBLEDiscovery(ctx, ios, opts.timeout)
	default:
		return fmt.Errorf("unknown discovery method: %s (valid: http, mdns, ble, coiot)", opts.method)
	}

	if err != nil {
		return err
	}

	// Run plugin detection if not skipped
	var pluginDevices []term.PluginDiscoveredDevice
	if !opts.skipPlugins && opts.method == "http" {
		// Plugin detection only works with HTTP scan since we need to probe addresses
		pluginDevices = runPluginDetection(ctx, ios, opts.subnet)
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
	if opts.register {
		addedShelly, regErr := utils.RegisterDiscoveredDevices(shellyDevices, opts.skipExisting)
		if regErr != nil {
			ios.Warning("Registration error: %v", regErr)
		}

		addedPlugin := registerPluginDevices(pluginDevices, opts.skipExisting)
		totalAdded := addedShelly + addedPlugin
		ios.Added("device", totalAdded)
	}

	return nil
}

// runPluginOnlyDiscovery runs discovery for a specific platform only.
func runPluginOnlyDiscovery(ctx context.Context, _ *cmdutil.Factory, ios *iostreams.IOStreams, opts *Options) error {
	// Get plugin registry
	registry, err := plugins.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize plugin registry: %w", err)
	}

	// Find plugin for the specified platform
	plugin, err := registry.FindByPlatform(opts.platform)
	if err != nil {
		return fmt.Errorf("failed to find plugin: %w", err)
	}
	if plugin == nil {
		return fmt.Errorf("no plugin found for platform %q (is shelly-%s installed?)", opts.platform, opts.platform)
	}

	// Get addresses to scan
	subnet := opts.subnet
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

	ios.Info("Scanning %d addresses for %s devices...", len(addresses), opts.platform)

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	mw.AddLine("scan", fmt.Sprintf("0/%d addresses probed", len(addresses)))

	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	// Create plugin discoverer and scan
	pluginDiscoverer := shelly.NewPluginDiscoverer(registry)
	var pluginDevices []term.PluginDiscoveredDevice
	var foundCount int

	for i, addr := range addresses {
		if ctx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlatform(ctx, addr, nil, opts.platform)
		if detectErr == nil && result != nil {
			pluginDevices = append(pluginDevices, toTermPluginDevice(result))
			foundCount++
			mw.UpdateLine("scan", iostreams.StatusRunning,
				fmt.Sprintf("%d/%d - found %s (%s)", i+1, len(addresses), result.Detection.DeviceName, result.Detection.Model))
		} else {
			mw.UpdateLine("scan", iostreams.StatusRunning,
				fmt.Sprintf("%d/%d addresses probed", i+1, len(addresses)))
		}
	}

	mw.UpdateLine("scan", iostreams.StatusSuccess,
		fmt.Sprintf("%d/%d addresses probed, %d %s devices found",
			len(addresses), len(addresses), foundCount, opts.platform))
	mw.Finalize()

	if len(pluginDevices) == 0 {
		ios.NoResults(opts.platform+" devices",
			fmt.Sprintf("Ensure %s devices are powered on and accessible in %s", opts.platform, ipNet))
		return nil
	}

	term.DisplayPluginDiscoveredDevices(ios, pluginDevices)

	if opts.register {
		added := registerPluginDevices(pluginDevices, opts.skipExisting)
		ios.Added("device", added)
	}

	return nil
}

// runPluginDetection runs plugin detection on addresses that Shelly detection may have missed.
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

	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return nil
	}

	ios.Info("Checking for plugin-managed devices...")

	pluginDiscoverer := shelly.NewPluginDiscoverer(registry)
	var pluginDevices []term.PluginDiscoveredDevice

	// Create a short timeout context for plugin detection
	pluginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, addr := range addresses {
		if pluginCtx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlugins(pluginCtx, addr, nil)
		if detectErr == nil && result != nil {
			pluginDevices = append(pluginDevices, toTermPluginDevice(result))
		}
	}

	return pluginDevices
}

// toTermPluginDevice converts a shelly plugin detection result to term display type.
func toTermPluginDevice(result *shelly.PluginDetectionResult) term.PluginDiscoveredDevice {
	components := make([]term.PluginComponentInfo, len(result.Detection.Components))
	for i, c := range result.Detection.Components {
		components[i] = term.PluginComponentInfo{
			Type: c.Type,
			ID:   c.ID,
			Name: c.Name,
		}
	}
	return term.PluginDiscoveredDevice{
		ID:         result.Detection.DeviceID,
		Name:       result.Detection.DeviceName,
		Model:      result.Detection.Model,
		Address:    result.Address,
		Platform:   result.Detection.Platform,
		Firmware:   result.Detection.Firmware,
		Components: components,
	}
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
func runDemoMode(f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()

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
	if opts.register {
		added, err := utils.RegisterDiscoveredDevices(shellyDevices, opts.skipExisting)
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
