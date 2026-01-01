// Package cmdutil provides discovery orchestration utilities.
package cmdutil

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultScanTimeout is the default timeout for HTTP subnet scanning.
const DefaultScanTimeout = 2 * time.Minute

// DiscoveryOptions holds options for discovery orchestration functions.
type DiscoveryOptions struct {
	Factory      *Factory
	Platform     string
	Register     bool
	SkipExisting bool
	Subnet       string
	Timeout      time.Duration
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
	pluginDevices := shelly.RunPluginPlatformDiscoveryWithProgress(ctx, registry, opts.Platform, subnet, shelly.IsDeviceRegistered, func(p shelly.DiscoveryProgress) bool {
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
func RunPluginDetection(ctx context.Context, ios *iostreams.IOStreams, subnet string) []term.PluginDiscoveredDevice {
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

	pluginDevices := shelly.RunPluginDetection(pluginCtx, registry, subnet, shelly.IsDeviceRegistered)

	return term.ConvertPluginDevices(pluginDevices)
}

// RunHTTPDiscovery performs HTTP subnet scanning discovery.
func RunHTTPDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration, subnet string) ([]discovery.DiscoveredDevice, error) {
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

// RunMDNSDiscovery performs mDNS/Zeroconf discovery.
func RunMDNSDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var devices []discovery.DiscoveredDevice
	err := RunWithSpinner(ctx, ios, "Discovering devices via mDNS...", func(_ context.Context) error {
		var discoverErr error
		var cleanup func()
		devices, cleanup, discoverErr = shelly.DiscoverMDNS(timeout)
		defer cleanup()
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
	err := RunWithSpinner(ctx, ios, "Discovering devices via CoIoT...", func(_ context.Context) error {
		var discoverErr error
		var cleanup func()
		devices, cleanup, discoverErr = shelly.DiscoverCoIoT(timeout)
		defer cleanup()
		return discoverErr
	})

	return devices, err
}

// RunBLEDiscovery performs Bluetooth Low Energy discovery.
func RunBLEDiscovery(ctx context.Context, ios *iostreams.IOStreams, timeout time.Duration) ([]discovery.DiscoveredDevice, error) {
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
	err = RunWithSpinner(ctx, ios, "Discovering devices via BLE...", func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		var discoverErr error
		devices, discoverErr = bleDiscoverer.DiscoverWithContext(ctx)
		return discoverErr
	})

	return devices, err
}
