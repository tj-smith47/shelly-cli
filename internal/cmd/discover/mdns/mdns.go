// Package mdns provides mDNS discovery command.
package mdns

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-go/discovery"

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
			return run(timeout, register, skipExisting)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	return cmd
}

func run(timeout time.Duration, register, skipExisting bool) error {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	spin := iostreams.NewSpinner("Discovering devices via mDNS...")
	spin.Start()

	mdnsDiscoverer := discovery.NewMDNSDiscoverer()
	defer func() {
		if err := mdnsDiscoverer.Stop(); err != nil {
			iostreams.DebugErr("stopping mDNS discoverer", err)
		}
	}()

	devices, err := mdnsDiscoverer.Discover(timeout)
	spin.Stop()

	if err != nil {
		return err
	}

	if len(devices) == 0 {
		iostreams.NoResults("devices", "Ensure devices are powered on and on the same network")
		return nil
	}

	helpers.DisplayDiscoveredDevices(devices)

	if register {
		added, err := helpers.RegisterDiscoveredDevices(devices, skipExisting)
		if err != nil {
			iostreams.Warning("Registration error: %v", err)
		}
		iostreams.Added("device", added)
	}

	return nil
}
