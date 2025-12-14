// Package scan provides subnet scanning discovery command.
package scan

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DefaultTimeout is the default scan timeout.
const DefaultTimeout = 2 * time.Minute

// NewCommand creates the discover scan command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var register bool
	var skipExisting bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:     "scan [subnet]",
		Aliases: []string{"search", "probe"},
		Short:   "Scan subnet for devices",
		Long: `Scan a network subnet for Shelly devices by probing HTTP endpoints.

If no subnet is provided, attempts to detect the local network.
This method is slower than mDNS or CoIoT but works when multicast
is blocked or devices are on different VLANs.`,
		Example: `  # Scan default network
  shelly discover scan

  # Scan specific subnet
  shelly discover scan 192.168.1.0/24

  # Auto-register discovered devices
  shelly discover scan --register

  # Scan with custom timeout
  shelly discover scan --timeout 5m`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subnet := ""
			if len(args) > 0 {
				subnet = args[0]
			}
			return run(cmd.Context(), f, subnet, timeout, register, skipExisting)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", DefaultTimeout, "Scan timeout")
	cmd.Flags().BoolVar(&register, "register", false, "Automatically register discovered devices")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", true, "Skip devices already registered")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, subnet string, timeout time.Duration, register, skipExisting bool) error {
	if subnet == "" {
		var err error
		subnet, err = detectSubnet()
		if err != nil {
			return fmt.Errorf("failed to detect subnet: %w", err)
		}
		iostreams.Info("Detected subnet: %s", subnet)
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	// Generate addresses to probe
	addresses := discovery.GenerateSubnetAddresses(subnet)
	if len(addresses) == 0 {
		return fmt.Errorf("no addresses to scan in subnet %s", subnet)
	}

	iostreams.Info("Scanning %d addresses in %s...", len(addresses), ipNet)

	ios := f.IOStreams()

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

	// Add progress line
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

	if len(devices) == 0 {
		iostreams.NoResults("devices", "Ensure devices are powered on and accessible on the network")
		return nil
	}

	helpers.DisplayDiscoveredDevices(devices)

	// Save discovered addresses to completion cache
	deviceAddrs := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceAddrs = append(deviceAddrs, d.Address.String())
	}
	if err := cmdutil.SaveDiscoveryCache(deviceAddrs); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if register {
		added, err := helpers.RegisterDiscoveredDevices(devices, skipExisting)
		if err != nil {
			iostreams.Warning("Registration error: %v", err)
		}
		iostreams.Added("device", added)
	}

	return nil
}

func detectSubnet() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				// Return the network address with mask
				network := ipNet.IP.Mask(ipNet.Mask)
				ones, _ := ipNet.Mask.Size()
				return fmt.Sprintf("%s/%d", network, ones), nil
			}
		}
	}

	return "", fmt.Errorf("no suitable network interface found")
}
