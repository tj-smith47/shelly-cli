// Package coiot provides CoIoT discovery command.
package coiot

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultTimeout is the default discovery timeout.
const DefaultTimeout = 10 * time.Second

// Discoverer defines the interface for CoIoT discovery.
type Discoverer interface {
	Discover(timeout time.Duration) ([]discovery.DiscoveredDevice, error)
	Stop() error
}

// newDiscoverer is the factory function for creating discoverers.
// It can be replaced in tests to inject mock discoverers.
var newDiscoverer = func() Discoverer {
	return discovery.NewCoIoTDiscoverer()
}

// Options holds command options.
type Options struct {
	Factory      *cmdutil.Factory
	Timeout      time.Duration
	Register     bool
	SkipExisting bool
	Gen1Only     bool
	Verbose      bool
}

// NewCommand creates the discover coiot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "coiot",
		Aliases: []string{"coap"},
		Short:   "Discover devices via CoIoT",
		Long: `Discover Shelly devices on the network using CoIoT multicast.

CoIoT is a protocol used by Gen1 Shelly devices to announce their presence
on the local network. This command listens for CoAP announcements on the
multicast group 224.0.1.187:5683.

Gen1-specific information displayed:
  - Device type and firmware
  - Number of relays and meters
  - CoIoT status values`,
		Example: `  # Basic CoIoT discovery
  shelly discover coiot

  # With longer timeout
  shelly discover coiot --timeout 30s

  # Show only Gen1 devices (filter out Gen2+)
  shelly discover coiot --gen1-only

  # Show detailed Gen1 info
  shelly discover coiot --verbose

  # Auto-register discovered devices
  shelly discover coiot --register`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", DefaultTimeout, "Discovery timeout")
	cmd.Flags().BoolVar(&opts.Register, "register", false, "Automatically register discovered devices")
	cmd.Flags().BoolVar(&opts.SkipExisting, "skip-existing", true, "Skip devices already registered")
	cmd.Flags().BoolVar(&opts.Gen1Only, "gen1-only", false, "Show only Gen1 devices")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Show detailed Gen1-specific information")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	coiotDiscoverer := newDiscoverer()
	defer func() {
		if err := coiotDiscoverer.Stop(); err != nil {
			ios.DebugErr("stopping CoIoT discoverer", err)
		}
	}()

	var devices []discovery.DiscoveredDevice
	err := cmdutil.RunWithSpinner(ctx, ios, "Discovering devices via CoIoT...", func(ctx context.Context) error {
		var discoverErr error
		devices, discoverErr = coiotDiscoverer.Discover(timeout)
		return discoverErr
	})
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		ios.NoResults("devices", "CoIoT works best with Gen1 devices. Try 'shelly discover mdns' for Gen2+")
		return nil
	}

	// Filter and enhance with Gen1 info if requested
	if opts.Gen1Only || opts.Verbose {
		devices = shelly.FilterGen1Devices(ctx, devices, opts.Gen1Only)
	}

	if len(devices) == 0 {
		ios.NoResults("Gen1 devices", "No Gen1 devices found. Try without --gen1-only flag")
		return nil
	}

	if opts.Verbose {
		term.DisplayGen1Details(ctx, ios, devices)
	} else {
		term.DisplayDiscoveredDevices(ios, devices)
	}

	// Save discovered addresses to completion cache
	addresses := make([]string, 0, len(devices))
	for _, d := range devices {
		addresses = append(addresses, d.Address.String())
	}
	if err := completion.SaveDiscoveryCache(addresses); err != nil {
		ios.DebugErr("saving discovery cache", err)
	}

	if opts.Register {
		added, err := utils.RegisterDiscoveredDevices(devices, opts.SkipExisting)
		if err != nil {
			ios.Warning("Registration error: %v", err)
		}
		ios.Added("device", added)
	}

	return nil
}
