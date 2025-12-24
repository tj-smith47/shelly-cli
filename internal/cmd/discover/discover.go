// Package discover provides device discovery commands.
package discover

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/httpscan"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/mdns"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultScanTimeout is the default timeout for HTTP subnet scanning.
const DefaultScanTimeout = 2 * time.Minute

// discoverOptions holds all discovery options.
type discoverOptions struct {
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
	opts := &discoverOptions{}

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
func run(ctx context.Context, f *cmdutil.Factory, opts *discoverOptions) error {
	ios := f.IOStreams()

	// If platform is specified (and not "shelly"), skip native Shelly discovery
	// and only run plugin detection for that platform
	if opts.platform != "" && opts.platform != model.PlatformShelly {
		return runPluginOnlyDiscovery(ctx, f, opts)
	}

	var shellyDevices []discovery.DiscoveredDevice
	var err error

	switch opts.method {
	case "http", "scan", "":
		shellyDevices, err = runHTTP(ctx, f, opts.timeout, opts.subnet)
	case "mdns":
		shellyDevices, err = runMDNS(ctx, f, opts.timeout)
	case "coiot":
		shellyDevices, err = runCoIoT(ctx, f, opts.timeout)
	case "ble":
		shellyDevices, err = runBLE(ctx, f, opts.timeout)
	default:
		return fmt.Errorf("unknown discovery method: %s (valid: http, mdns, ble, coiot)", opts.method)
	}

	if err != nil {
		return err
	}

	// Run plugin detection if not skipped
	var pluginDevices []shelly.PluginDiscoveredDevice
	if !opts.skipPlugins && opts.method == "http" {
		// Plugin detection only works with HTTP scan since we need to probe addresses
		pluginDevices = runPluginDetection(ctx, f, opts)
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
		displayPluginDiscoveredDevices(ios, pluginDevices)
	}

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, totalDevices)
	for _, d := range shellyDevices {
		addresses = append(addresses, d.Address.String())
	}
	for _, d := range pluginDevices {
		if d.Address != nil {
			addresses = append(addresses, d.Address.String())
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

		addedPlugin := registerPluginDevices(ios, pluginDevices, opts.skipExisting)
		totalAdded := addedShelly + addedPlugin
		ios.Added("device", totalAdded)
	}

	return nil
}

// runPluginOnlyDiscovery runs discovery for a specific platform only.
func runPluginOnlyDiscovery(ctx context.Context, f *cmdutil.Factory, opts *discoverOptions) error {
	ios := f.IOStreams()

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
	var pluginDevices []shelly.PluginDiscoveredDevice
	var foundCount int

	for i, addr := range addresses {
		if ctx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlatform(ctx, addr, nil, opts.platform)
		if detectErr == nil && result != nil {
			added := isPluginDeviceRegistered(result.Address)
			pluginDevices = append(pluginDevices, shelly.ToPluginDiscoveredDevice(result, added))
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

	displayPluginDiscoveredDevices(ios, pluginDevices)

	if opts.register {
		added := registerPluginDevices(ios, pluginDevices, opts.skipExisting)
		ios.Added("device", added)
	}

	return nil
}

// runPluginDetection runs plugin detection on addresses that Shelly detection may have missed.
func runPluginDetection(ctx context.Context, f *cmdutil.Factory, opts *discoverOptions) []shelly.PluginDiscoveredDevice {
	ios := f.IOStreams()

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
	// For now, we'll re-scan the subnet for plugin devices
	// In the future, we could track addresses that failed Shelly detection
	subnet := opts.subnet
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
	var pluginDevices []shelly.PluginDiscoveredDevice

	// Create a short timeout context for plugin detection
	pluginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, addr := range addresses {
		if pluginCtx.Err() != nil {
			break
		}

		result, detectErr := pluginDiscoverer.DetectWithPlugins(pluginCtx, addr, nil)
		if detectErr == nil && result != nil {
			added := isPluginDeviceRegistered(result.Address)
			pluginDevices = append(pluginDevices, shelly.ToPluginDiscoveredDevice(result, added))
		}
	}

	return pluginDevices
}

// displayPluginDiscoveredDevices prints a table of plugin-discovered devices.
func displayPluginDiscoveredDevices(ios *iostreams.IOStreams, devices []shelly.PluginDiscoveredDevice) {
	if len(devices) == 0 {
		return
	}

	ios.Println()
	ios.Title("Plugin-managed Devices")

	for _, d := range devices {
		name := d.Name
		if name == "" {
			name = d.ID
		}
		if name == "" && d.Address != nil {
			name = d.Address.String()
		}

		ios.Printf("  %s\n", name)
		if d.Address != nil {
			ios.Printf("    Address:  %s\n", d.Address)
		}
		ios.Printf("    Platform: %s\n", d.Platform)
		if d.Model != "" {
			ios.Printf("    Model:    %s\n", d.Model)
		}
		if d.Firmware != "" {
			ios.Printf("    Firmware: %s\n", d.Firmware)
		}
		if len(d.Components) > 0 {
			ios.Printf("    Components:\n")
			for _, c := range d.Components {
				if c.Name != "" {
					ios.Printf("      - %s:%d (%s)\n", c.Type, c.ID, c.Name)
				} else {
					ios.Printf("      - %s:%d\n", c.Type, c.ID)
				}
			}
		}
		ios.Println()
	}

	ios.Count(devices[0].Platform+" device", len(devices))
}

// registerPluginDevices registers plugin-detected devices.
func registerPluginDevices(ios *iostreams.IOStreams, devices []shelly.PluginDiscoveredDevice, skipExisting bool) int {
	added := 0
	for _, d := range devices {
		if d.Address == nil {
			continue
		}

		// Convert to DeviceDetectionResult for registration
		result := &plugins.DeviceDetectionResult{
			Detected:   true,
			Platform:   d.Platform,
			DeviceID:   d.ID,
			DeviceName: d.Name,
			Model:      d.Model,
			Firmware:   d.Firmware,
		}

		wasAdded, err := utils.RegisterPluginDiscoveredDevice(result, d.Address.String(), skipExisting)
		if err != nil {
			ios.DebugErr("registering plugin device", err)
			continue
		}
		if wasAdded {
			added++
		}
	}
	return added
}

// isPluginDeviceRegistered checks if a plugin device is already registered.
func isPluginDeviceRegistered(address string) bool {
	devices := utils.ListRegisteredDevices()
	for _, dev := range devices {
		if dev.Address == address {
			return true
		}
	}
	return false
}

// runHTTP performs HTTP subnet scanning discovery.
func runHTTP(ctx context.Context, f *cmdutil.Factory, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
	ios := f.IOStreams()

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

// runMDNS performs mDNS/Zeroconf discovery.
func runMDNS(ctx context.Context, f *cmdutil.Factory, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios := f.IOStreams()

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	var devices []discovery.DiscoveredDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via mDNS...", func(_ context.Context) error {
		var discoverErr error
		devices, discoverErr = mdnsDiscoverer.Discover(timeout)
		return discoverErr
	})

	return devices, err
}

// runCoIoT performs CoIoT/CoAP discovery.
func runCoIoT(ctx context.Context, f *cmdutil.Factory, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios := f.IOStreams()

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	var devices []discovery.DiscoveredDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via CoIoT...", func(_ context.Context) error {
		var discoverErr error
		devices, discoverErr = coiotDiscoverer.Discover(timeout)
		return discoverErr
	})

	return devices, err
}

// runBLE performs Bluetooth Low Energy discovery.
func runBLE(ctx context.Context, f *cmdutil.Factory, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	ios := f.IOStreams()

	if timeout == 0 {
		timeout = 15 * time.Second
	}

	bleDiscoverer, err := discovery.NewBLEDiscoverer()
	if err != nil {
		ios.Error("BLE discovery is not available on this system")
		ios.Hint("Ensure you have a Bluetooth adapter and it is enabled")
		ios.Hint("On Linux, you may need to run with elevated privileges")
		return nil, nil
	}
	defer func() {
		if stopErr := bleDiscoverer.Stop(); stopErr != nil {
			ios.DebugErr("stopping BLE discoverer", stopErr)
		}
	}()

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
