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
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultScanTimeout is the default timeout for HTTP subnet scanning.
const DefaultScanTimeout = 2 * time.Minute

// NewCommand creates the discover command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout      time.Duration
		register     bool
		skipExisting bool
		subnet       string
		method       string
	)

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
  coiot  - CoIoT/CoAP discovery (Gen1 devices)`,
		Example: `  # Discover devices via HTTP scan (default, auto-detects subnet)
  shelly discover

  # Specify subnet for HTTP scan
  shelly discover --subnet 192.168.1.0/24

  # Use mDNS instead of HTTP scan
  shelly discover --method mdns

  # Use BLE discovery
  shelly discover --method ble

  # Auto-register discovered devices
  shelly discover --register`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, timeout, register, skipExisting, subnet, method)
		},
	}

	// Add flags for the parent discover command
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultScanTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().StringVar(&subnet, "subnet", "", "Subnet to scan (auto-detected if not specified)")
	cmd.Flags().StringVarP(&method, "method", "m", "http", "Discovery method: http, mdns, ble, coiot")

	// Add subcommands (kept for direct access)
	cmd.AddCommand(mdns.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))
	cmd.AddCommand(coiot.NewCommand(f))
	cmd.AddCommand(httpscan.NewCommand(f))

	return cmd
}

// run runs device discovery using the selected method.
func run(ctx context.Context, f *cmdutil.Factory, timeout time.Duration, register, skipExisting bool, subnet, method string) error {
	ios := f.IOStreams()

	var devices []discovery.DiscoveredDevice
	var err error

	switch method {
	case "http", "scan", "":
		devices, err = runHTTP(ctx, f, timeout, subnet)
	case "mdns":
		devices, err = runMDNS(ctx, f, timeout)
	case "coiot":
		devices, err = runCoIoT(ctx, f, timeout)
	case "ble":
		devices, err = runBLE(ctx, f, timeout)
	default:
		return fmt.Errorf("unknown discovery method: %s (valid: http, mdns, ble, coiot)", method)
	}

	if err != nil {
		return err
	}

	if len(devices) == 0 {
		ios.NoResults("devices", "Ensure devices are powered on and accessible on the network")
		return nil
	}

	term.DisplayDiscoveredDevices(ios, devices)

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, len(devices))
	for _, d := range devices {
		addresses = append(addresses, d.Address.String())
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if register {
		added, regErr := utils.RegisterDiscoveredDevices(devices, skipExisting)
		if regErr != nil {
			ios.Warning("Registration error: %v", regErr)
		}
		ios.Added("device", added)
	}

	return nil
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
