// Package httpscan provides HTTP subnet scanning discovery command.
package httpscan

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultTimeout is the default scan timeout.
const DefaultTimeout = 2 * time.Minute

// Options holds the command options.
type Options struct {
	Factory      *cmdutil.Factory
	Subnets      []string
	Register     bool
	SkipExisting bool
	AllNetworks  bool
	Timeout      time.Duration
}

// NewCommand creates the discover http command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "http [subnet...]",
		Aliases: []string{"scan", "search", "probe"},
		Short:   "Discover devices via HTTP subnet scanning",
		Long: `Discover Shelly devices by probing HTTP endpoints on a subnet.

If no subnet is provided, attempts to detect the local network(s).
When multiple subnets are detected, an interactive prompt lets you
choose which to scan. Use --all-networks to scan all without prompting.

This method is slower than mDNS or CoIoT but works when multicast
is blocked or devices are on different VLANs.

The scan probes each IP address in the subnet range for Shelly device
HTTP endpoints. Progress is shown in real-time. Discovered devices
can be automatically registered with --register.

Use --skip-existing (enabled by default) to avoid re-registering
devices that are already in your registry.

Output is formatted as a table showing: ID, Address, Model, Generation,
Protocol, and Auth status.`,
		Example: `  # Scan default network (auto-detect)
  shelly discover http

  # Scan specific subnet
  shelly discover http 192.168.1.0/24

  # Scan multiple subnets
  shelly discover http 192.168.1.0/24 10.0.0.0/24

  # Scan all detected subnets without prompting
  shelly discover http --all-networks

  # Use --network flag (repeatable)
  shelly discover http --network 192.168.1.0/24 --network 10.0.0.0/24

  # Scan a /16 network (large, use longer timeout)
  shelly discover http 10.0.0.0/16 --timeout 30m

  # Auto-register discovered devices
  shelly discover http --register

  # Using 'scan' alias
  shelly discover scan --timeout 5m

  # Force re-register all discovered devices
  shelly discover http --register --skip-existing=false

  # Combine flags for initial network setup
  shelly discover http 192.168.1.0/24 --register --timeout 10m`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Subnets = append(opts.Subnets, args...)
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", DefaultTimeout, "Scan timeout")
	cmd.Flags().BoolVar(&opts.Register, "register", false, "Automatically register discovered devices")
	cmd.Flags().BoolVar(&opts.SkipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().StringArrayVar(&opts.Subnets, "network", nil, "Subnet(s) to scan (repeatable, auto-detected if not specified)")
	cmd.Flags().BoolVar(&opts.AllNetworks, "all-networks", false, "Scan all detected subnets without prompting")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	subnets, err := cmdutil.ResolveSubnets(ios, opts.Subnets, opts.AllNetworks)
	if err != nil {
		return err
	}

	// Generate addresses across all subnets
	var allAddresses []string
	for _, subnet := range subnets {
		allAddresses = append(allAddresses, discovery.GenerateSubnetAddresses(subnet)...)
	}
	if len(allAddresses) == 0 {
		return fmt.Errorf("no addresses to scan in %s", strings.Join(subnets, ", "))
	}

	subnetLabel := strings.Join(subnets, ", ")
	ios.Success("Scanning %d addresses in %s", len(allAddresses), subnetLabel)

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Use animated spinner for scan progress
	ios.StartProgress(fmt.Sprintf("0/%d addresses probed", len(allAddresses)))

	devices := discovery.ProbeAddressesWithProgress(ctx, allAddresses, func(p discovery.ProbeProgress) bool {
		msg := fmt.Sprintf("%d/%d addresses probed", p.Done, p.Total)
		if p.Found && p.Device != nil {
			msg = fmt.Sprintf("%d/%d - found %s (%s)", p.Done, p.Total, p.Device.Name, p.Device.Model)
		}
		ios.UpdateProgress(msg)
		return ctx.Err() == nil // Continue unless context canceled
	})

	// Stop spinner with success summary
	ios.StopProgressWithSuccess(fmt.Sprintf("%d/%d addresses probed, %d devices found",
		len(allAddresses), len(allAddresses), len(devices)))

	if len(devices) == 0 {
		ios.NoResults("devices", "Ensure devices are powered on and accessible on the network")
		return nil
	}

	term.DisplayDiscoveredDevices(ios, devices)

	// Save discovered addresses to completion cache
	deviceAddrs := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceAddrs = append(deviceAddrs, d.Address.String())
	}
	if err := completion.SaveDiscoveryCache(deviceAddrs); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if opts.Register {
		added, regErr := utils.RegisterDiscoveredDevices(devices, opts.SkipExisting)
		if regErr != nil {
			ios.Warning("Registration error: %v", regErr)
		}
		ios.Added("device", added)
	}

	return nil
}
