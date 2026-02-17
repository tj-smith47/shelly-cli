// Package discover provides device discovery commands.
package discover

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/httpscan"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/mdns"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

const methodHTTP = "http"

// Options holds all discovery options.
type Options struct {
	Factory      *cmdutil.Factory
	Method       string
	Platform     string
	Subnets      []string
	Register     bool
	SkipExisting bool
	SkipPlugins  bool
	AllNetworks  bool
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

  # Scan multiple subnets
  shelly discover --subnet 192.168.1.0/24 --subnet 10.0.0.0/24

  # Scan all detected subnets without prompting
  shelly discover --all-networks

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
	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", cmdutil.DefaultScanTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&opts.Register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&opts.SkipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().StringArrayVar(&opts.Subnets, "subnet", nil, "Subnet(s) to scan (repeatable, auto-detected if not specified)")
	cmd.Flags().BoolVar(&opts.AllNetworks, "all-networks", false, "Scan all detected subnets without prompting")
	cmd.Flags().StringVarP(&opts.Method, "method", "m", methodHTTP, "Discovery method: http, mdns, ble, coiot")
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
		return mock.RunDemoDiscovery(ios, opts.Register, opts.SkipExisting)
	}

	// If platform is specified (and not "shelly"), skip native Shelly discovery
	// and only run plugin detection for that platform
	if opts.Platform != "" && opts.Platform != model.PlatformShelly {
		return cmdutil.RunPluginOnlyDiscovery(ctx, &cmdutil.DiscoveryOptions{
			Factory:      opts.Factory,
			Platform:     opts.Platform,
			Register:     opts.Register,
			SkipExisting: opts.SkipExisting,
			Subnets:      opts.Subnets,
			AllNetworks:  opts.AllNetworks,
			Timeout:      opts.Timeout,
		})
	}

	// Resolve subnets for HTTP scanning
	var subnets []string
	if opts.Method == "" || opts.Method == methodHTTP || opts.Method == "scan" {
		var resolveErr error
		subnets, resolveErr = cmdutil.ResolveSubnets(ios, opts.Subnets, opts.AllNetworks)
		if resolveErr != nil {
			return resolveErr
		}
	}

	var shellyDevices []discovery.DiscoveredDevice
	var err error

	switch opts.Method {
	case methodHTTP, "scan", "":
		shellyDevices, err = cmdutil.RunHTTPDiscovery(ctx, ios, opts.Timeout, subnets)
	case "mdns":
		shellyDevices, err = cmdutil.RunMDNSDiscovery(ctx, ios, opts.Timeout)
	case "coiot":
		shellyDevices, err = cmdutil.RunCoIoTDiscovery(ctx, ios, opts.Timeout)
	case "ble":
		shellyDevices, err = cmdutil.RunBLEDiscovery(ctx, ios, opts.Timeout)
	default:
		return fmt.Errorf("unknown discovery method: %s (valid: http, mdns, ble, coiot)", opts.Method)
	}

	if err != nil {
		return err
	}

	// Run plugin detection if not skipped
	var pluginDevices []term.PluginDiscoveredDevice
	if !opts.SkipPlugins && opts.Method == methodHTTP {
		pluginDevices = cmdutil.RunPluginDetection(ctx, ios, subnets)
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

		addedPlugin := term.RegisterPluginDevices(pluginDevices, opts.SkipExisting)
		totalAdded := addedShelly + addedPlugin
		ios.Added("device", totalAdded)
	}

	return nil
}
