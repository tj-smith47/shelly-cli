// Package add provides the bthome add command.
package add

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Duration int
	Addr     string
	Name     string
}

// NewCommand creates the bthome add command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "add <device>",
		Aliases: []string{"discover", "scan"},
		Short:   "Add a BTHome device",
		Long: `Add a BTHome device to a Shelly gateway.

This command can either:
1. Start a discovery scan to find nearby BTHome devices
2. Add a specific device by MAC address

When scanning, the gateway will broadcast discovery requests and listen
for BTHome device advertisements. Discovered devices emit events that
can be monitored with 'shelly monitor events'.`,
		Example: `  # Start 30-second discovery scan
  shelly bthome add living-room

  # Custom scan duration (60 seconds)
  shelly bthome add living-room --duration 60

  # Add specific device by MAC address
  shelly bthome add living-room --addr 3c:2e:f5:71:d5:2a

  # Add device with a name
  shelly bthome add living-room --addr 3c:2e:f5:71:d5:2a --name "Door Sensor"`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Duration, "duration", 30, "Discovery scan duration in seconds")
	cmd.Flags().StringVar(&opts.Addr, "addr", "", "MAC address of device to add directly")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Name for the device (with --addr)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(opts.Duration+10)*time.Second)
	defer cancel()

	svc := opts.Factory.ShellyService()

	// If address provided, add device directly
	if opts.Addr != "" {
		return addDeviceDirectly(ctx, svc, opts)
	}

	// Otherwise start discovery scan
	return startDiscovery(ctx, svc, opts)
}

func addDeviceDirectly(ctx context.Context, svc *shelly.Service, opts *Options) error {
	ios := opts.Factory.IOStreams()

	result, err := svc.BTHomeAddDevice(ctx, opts.Device, opts.Addr, opts.Name)
	if err != nil {
		return err
	}

	ios.Success("BTHome device added: %s", result.Key)
	if opts.Name != "" {
		ios.Info("Name: %s", opts.Name)
	}
	ios.Info("Address: %s", opts.Addr)

	return nil
}

func startDiscovery(ctx context.Context, svc *shelly.Service, opts *Options) error {
	ios := opts.Factory.IOStreams()

	if err := svc.BTHomeStartDiscovery(ctx, opts.Device, opts.Duration); err != nil {
		return err
	}

	ios.Println(theme.Bold().Render("BTHome Device Discovery Started"))
	ios.Println()
	ios.Info("Scanning for %d seconds...", opts.Duration)
	ios.Println()
	ios.Info("Discovered devices will emit 'device_discovered' events.")
	ios.Info("Monitor events with: shelly monitor events %s", opts.Device)
	ios.Println()
	ios.Info("When discovery completes, a 'discovery_done' event will be emitted.")
	ios.Info("Then use 'shelly bthome add %s --addr <mac>' to add discovered devices.", opts.Device)

	return nil
}
