// Package mdns provides mDNS discovery command.
package mdns

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
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
		Use:   "mdns",
		Short: "Discover devices using mDNS/Zeroconf",
		Long: `Discover Shelly devices using mDNS/Zeroconf.

This method works best for Gen2+ devices that advertise the
_shelly._tcp.local service.

Examples:
  # Basic mDNS discovery
  shelly discover mdns

  # With longer timeout
  shelly discover mdns --timeout 30s

  # Auto-register discovered devices
  shelly discover mdns --register`,
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
			ios.DebugErr("stopping mDNS discoverer", err)
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

	helpers.DisplayDiscoveredDevices(devices)

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, len(devices))
	for _, d := range devices {
		addresses = append(addresses, d.Address.String())
	}
	if err := cmdutil.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if register {
		added, err := helpers.RegisterDiscoveredDevices(devices, skipExisting)
		if err != nil {
			ios.Warning("Registration error: %v", err)
		}
		ios.Added("device", added)
	}

	return nil
}
