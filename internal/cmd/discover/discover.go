// Package discover provides device discovery commands.
package discover

import (
	"context"
	"errors"
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

	// "", "http", and "scan" all dispatch to the HTTP subnet scan; compute the
	// predicate once so subnet resolution, the scan switch, and plugin detection
	// can never disagree about which methods run an HTTP scan.
	isHTTP := opts.Method == "" || opts.Method == methodHTTP || opts.Method == "scan"

	// Resolve subnets for HTTP scanning
	var subnets []string
	if isHTTP {
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
		// A handled discovery error (e.g. BLE unavailable) was already reported
		// with actionable hints; stop cleanly before the generic NoResults
		// summary so the user isn't told both "BLE not available" and
		// "no devices found".
		if errors.Is(err, cmdutil.ErrDiscoveryHandled) {
			return nil
		}
		return err
	}

	// Run plugin detection if not skipped
	var pluginDevices []term.PluginDiscoveredDevice
	if !opts.SkipPlugins && isHTTP && subnets != nil {
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

	// Cache + register the native Shelly devices via the shared helper so this
	// command and `discover http` stay in lockstep on cache/register wording.
	addedShelly := cmdutil.CacheAndRegisterDevices(ios, shellyDevices, opts.Register, opts.SkipExisting)

	// Plugin-discovered devices carry string addresses and a separate registry
	// path, so their cache + registration stays here rather than in the helper.
	pluginAddrs := make([]string, 0, len(pluginDevices))
	for _, d := range pluginDevices {
		if d.Address != "" {
			pluginAddrs = append(pluginAddrs, d.Address)
		}
	}
	if len(pluginAddrs) > 0 {
		if err := completion.SaveDiscoveryCache(pluginAddrs); err != nil {
			ios.DebugErr("saving discovery cache", err)
		}
	}

	if opts.Register {
		addedPlugin := term.RegisterPluginDevices(pluginDevices, opts.SkipExisting)
		ios.Added("device", addedShelly+addedPlugin)
	}

	return nil
}
