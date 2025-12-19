// Package add provides the bthome add command.
package add

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
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
	ios := opts.Factory.IOStreams()

	// If address provided, add device directly
	if opts.Addr != "" {
		result, err := svc.BTHomeAddDevice(ctx, opts.Device, opts.Addr, opts.Name)
		if err != nil {
			return err
		}
		term.DisplayBTHomeAddResult(ios, term.BTHomeAddResult{
			Key:  result.Key,
			Name: opts.Name,
			Addr: opts.Addr,
		})
		return nil
	}

	// Start discovery scan
	if err := svc.BTHomeStartDiscovery(ctx, opts.Device, opts.Duration); err != nil {
		return err
	}
	term.DisplayBTHomeDiscoveryStarted(ios, opts.Device, opts.Duration)

	return nil
}
