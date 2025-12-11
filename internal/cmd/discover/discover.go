// Package discover provides device discovery commands.
package discover

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/mdns"
	"github.com/tj-smith47/shelly-cli/internal/cmd/discover/scan"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DefaultTimeout is the default discovery timeout for the parent command.
const DefaultTimeout = 10 * time.Second

// NewCommand creates the discover command group.
func NewCommand() *cobra.Command {
	var (
		timeout      time.Duration
		register     bool
		skipExisting bool
	)

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover Shelly devices on the network",
		Long: `Discover Shelly devices using various protocols.

Available discovery methods:
  mdns   - mDNS/Zeroconf discovery (Gen2+ devices)
  ble    - Bluetooth Low Energy discovery (provisioning mode)
  coiot  - CoIoT/CoAP discovery (Gen1 devices)
  scan   - HTTP subnet scanning (all devices)

Examples:
  # Discover all devices using mDNS (default)
  shelly discover

  # Discover using mDNS only
  shelly discover mdns

  # Discover BLE devices in provisioning mode
  shelly discover ble

  # Scan a subnet
  shelly discover scan 192.168.1.0/24`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDiscover(timeout, register, skipExisting)
		},
	}

	// Add flags for the parent discover command
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Auto-register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	// Add subcommands
	cmd.AddCommand(mdns.NewCommand())
	cmd.AddCommand(ble.NewCommand())
	cmd.AddCommand(coiot.NewCommand())
	cmd.AddCommand(scan.NewCommand())

	return cmd
}

// runDiscover runs mDNS discovery as the default discovery method.
func runDiscover(timeout time.Duration, register, skipExisting bool) error {
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
