// Package mdns provides mDNS discovery command.
package mdns

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultTimeout is the default mDNS discovery timeout.
const DefaultTimeout = 10 * time.Second

// NewCommand creates the mDNS discovery command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout      time.Duration
		register     bool
		skipExisting bool
	)

	cmd := &cobra.Command{
		Use:     "mdns",
		Aliases: []string{"zeroconf", "bonjour"},
		Short:   "Discover devices using mDNS/Zeroconf",
		Long: `Discover Shelly devices using mDNS/Zeroconf.

mDNS (Multicast DNS) is the fastest discovery method. Devices broadcast
their presence on the local network using the _shelly._tcp.local service.
This works best for Gen2+ devices.

Note: mDNS requires multicast support on your network. If devices aren't
found, try 'shelly discover scan' which probes addresses directly.

Output is formatted as a table showing: ID, Address, Model, Generation,
Protocol, and Auth status.`,
		Example: `  # Basic mDNS discovery
  shelly discover mdns

  # With longer timeout for slow networks
  shelly discover mdns --timeout 30s

  # Auto-register discovered devices
  shelly discover mdns --register

  # Register but skip devices already in registry
  shelly discover mdns --register --skip-existing

  # Force re-register all discovered devices
  shelly discover mdns --register --skip-existing=false

  # Using aliases
  shelly discover zeroconf --timeout 20s
  shelly discover bonjour --register`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f, timeout, register, skipExisting)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	return cmd
}

func run(f *cmdutil.Factory, timeout time.Duration, register, skipExisting bool) error {
	ios := f.IOStreams()

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ios.StartProgress("Discovering devices via mDNS...")

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			ios.DebugErrCat(iostreams.CategoryDiscovery, "stopping mDNS discoverer", err)
		}
	}()

	devices, err := mdnsDiscoverer.Discover(timeout)
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
