// Package discover provides device discovery commands.
package discover

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/mdns"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/scan"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// NewCommand creates the discover command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout      time.Duration
		register     bool
		skipExisting bool
	)

	cmd := &cobra.Command{
		Use:     "discover",
		Aliases: []string{"disc", "find"},
		Short:   "Discover Shelly devices on the network",
		Long: `Discover Shelly devices using various protocols.

Available discovery methods:
  mdns   - mDNS/Zeroconf discovery (Gen2+ devices)
  ble    - Bluetooth Low Energy discovery (provisioning mode)
  coiot  - CoIoT/CoAP discovery (Gen1 devices)
  scan   - HTTP subnet scanning (all devices)`,
		Example: `  # Discover all devices using mDNS (default)
  shelly discover

  # Discover using mDNS only
  shelly disc mdns

  # Discover BLE devices in provisioning mode
  shelly discover ble

  # Scan a subnet
  shelly find scan 192.168.1.0/24`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDiscover(cmd.Context(), f, timeout, register, skipExisting)
		},
	}

	// Add flags for the parent discover command
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", shelly.DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	// Add subcommands
	cmd.AddCommand(mdns.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))
	cmd.AddCommand(coiot.NewCommand(f))
	cmd.AddCommand(scan.NewCommand(f))

	return cmd
}

// runDiscover runs mDNS discovery as the default discovery method.
func runDiscover(ctx context.Context, f *cmdutil.Factory, timeout time.Duration, register, skipExisting bool) error {
	ios := f.IOStreams()

	if timeout == 0 {
		timeout = shelly.DefaultTimeout
	}

	ios.StartProgress("Discovering devices via mDNS...")

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	// Create context with timeout for discovery
	discoverCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	devices, err := mdnsDiscoverer.DiscoverWithContext(discoverCtx)
	ios.StopProgress()

	if err != nil {
		return err
	}

	if len(devices) == 0 {
		ios.NoResults("devices", "Ensure devices are powered on and on the same network")
		return nil
	}

	cmdutil.DisplayDiscoveredDevices(ios, devices)

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, len(devices))
	for _, d := range devices {
		addresses = append(addresses, d.Address.String())
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if register {
		added, err := utils.RegisterDiscoveredDevices(devices, skipExisting)
		if err != nil {
			ios.Warning("Registration error: %v", err)
		}
		ios.Added("device", added)
	}

	return nil
}
