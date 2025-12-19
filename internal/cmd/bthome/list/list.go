// Package list provides the bthome list command.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the bthome list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "devices"},
		Short:   "List BTHome devices",
		Long: `List all BTHome devices connected to a Shelly gateway.

BTHome is a Bluetooth Low Energy (BLE) protocol for sensors and devices.
Shelly BLU sensors (motion, door/window, button) connect to a gateway
device that collects their readings.

Shows configured BTHomeDevice components with their current status,
signal strength (RSSI), battery level, and last update time.

Use 'shelly bthome add' to discover and pair new devices.
Use 'shelly bthome sensors' to view sensor readings.

Output is formatted as styled text by default. Use --json for
structured output suitable for scripting.`,
		Example: `  # List all BTHome devices
  shelly bthome list living-room

  # Output as JSON
  shelly bthome list living-room --json

  # Get devices with low battery
  shelly bthome list living-room --json | jq '.[] | select(.battery != null and .battery < 20)'

  # Find devices with weak signal
  shelly bthome list living-room --json | jq '.[] | select(.rssi != null and .rssi < -80)'

  # Get device addresses (MAC)
  shelly bthome list living-room --json | jq -r '.[].addr'

  # List device names and IDs
  shelly bthome list living-room --json | jq '.[] | {name, id}'

  # Short form
  shelly bthome ls living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	devices, err := svc.FetchBTHomeDevices(ctx, opts.Device, ios)
	if err != nil {
		return err
	}

	if opts.JSON {
		return output.JSON(ios.Out, devices)
	}

	term.DisplayBTHomeDevices(ios, devices, opts.Device)
	return nil
}
