// Package coiot provides CoIoT discovery command.
package coiot

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DefaultTimeout is the default discovery timeout.
const DefaultTimeout = 10 * time.Second

// NewCommand creates the discover coiot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var timeout time.Duration
	var register bool
	var skipExisting bool

	cmd := &cobra.Command{
		Use:   "coiot",
		Short: "Discover devices via CoIoT",
		Long: `Discover Shelly devices on the network using CoIoT multicast.

CoIoT is a protocol used by Gen1 Shelly devices to announce their presence
on the local network. This command listens for CoAP announcements on the
multicast group 224.0.1.187:5683.

Examples:
  # Basic CoIoT discovery
  shelly discover coiot

  # With longer timeout
  shelly discover coiot --timeout 30s

  # Auto-register discovered devices
  shelly discover coiot --register`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(timeout, register, skipExisting)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Automatically register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	return cmd
}

func run(timeout time.Duration, register, skipExisting bool) error {
	ios := iostreams.System()

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	ios.StartProgress("Discovering devices via CoIoT...")

	coiotDiscoverer := discovery.NewCoIoTDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	devices, err := coiotDiscoverer.Discover(timeout)
	ios.StopProgress()

	if err != nil {
		return err
	}

	if len(devices) == 0 {
		ios.NoResults("devices", "CoIoT works best with Gen1 devices. Try 'shelly discover mdns' for Gen2+")
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
