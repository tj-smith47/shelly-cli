// Package coiot provides CoIoT discovery command.
package coiot

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// DefaultTimeout is the default discovery timeout.
const DefaultTimeout = 10 * time.Second

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

	// Filter and enhance with Gen1 info if requested
	if opts.Gen1Only || opts.Verbose {
		devices = filterAndEnhanceGen1(ctx, devices, opts.Gen1Only)
	}

	if len(devices) == 0 {
		ios.NoResults("Gen1 devices", "No Gen1 devices found. Try without --gen1-only flag")
		return nil
	}

	if opts.Verbose {
		displayGen1Details(ctx, ios, devices)
	} else {
		cmdutil.DisplayDiscoveredDevices(ios, devices)
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

// filterAndEnhanceGen1 filters for Gen1 devices and optionally enhances with extra info.
func filterAndEnhanceGen1(ctx context.Context, devices []discovery.DiscoveredDevice, filterOnly bool) []discovery.DiscoveredDevice {
	var filtered []discovery.DiscoveredDevice

	for _, d := range devices {
		// Detect device generation
		result, err := client.DetectGeneration(ctx, d.Address.String(), nil)
		if err != nil {
			// Can't detect, include if not filtering
			if !filterOnly {
				filtered = append(filtered, d)
			}
			continue
		}

		if result.IsGen1() {
			filtered = append(filtered, d)
		} else if !filterOnly {
			// Include non-Gen1 if not filtering
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// displayGen1Details shows detailed Gen1-specific information.
func displayGen1Details(ctx context.Context, ios *iostreams.IOStreams, devices []discovery.DiscoveredDevice) {
	ios.Println(theme.Bold().Render(fmt.Sprintf("Found %d device(s):", len(devices))))
	ios.Println()

	for _, d := range devices {
		displaySingleGen1Device(ctx, ios, d)
	}
}

// displaySingleGen1Device displays details for a single device.
func displaySingleGen1Device(ctx context.Context, ios *iostreams.IOStreams, d discovery.DiscoveredDevice) {
	ios.Printf("  %s\n", theme.Highlight().Render(d.Name))
	ios.Printf("    Address: %s\n", d.Address)
	ios.Printf("    Model:   %s\n", d.Model)

	// Try to get Gen1-specific details
	result, err := client.DetectGeneration(ctx, d.Address.String(), nil)
	if err != nil {
		ios.Printf("    Gen:     %s\n", theme.Dim().Render("unknown"))
		ios.Println()
		return
	}

	if !result.IsGen1() {
		ios.Printf("    Gen:     %s\n", theme.Dim().Render(fmt.Sprintf("Gen%d", result.Generation)))
		ios.Println()
		return
	}

	displayGen1Info(ctx, ios, d, result)
	ios.Println()
}

// displayGen1Info displays Gen1-specific device information.
func displayGen1Info(ctx context.Context, ios *iostreams.IOStreams, d discovery.DiscoveredDevice, result *client.DetectionResult) {
	ios.Printf("    Gen:     %s\n", theme.StatusOK().Render("Gen1"))
	ios.Printf("    Type:    %s\n", result.DeviceType)
	ios.Printf("    FW:      %s\n", result.Firmware)
	if result.AuthEn {
		ios.Printf("    Auth:    %s\n", theme.StatusWarn().Render("enabled"))
	}

	// Try to get full Gen1 status for more details
	gen1Device := model.Device{Address: d.Address.String()}
	gen1Client, err := client.ConnectGen1(ctx, gen1Device)
	if err != nil {
		return
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	if _, err := gen1Client.GetStatus(ctx); err == nil {
		ios.Printf("    Status:  %s\n", theme.StatusOK().Render("available"))
	}
}
