// Package list provides the power list command.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Device  string
	Factory *cmdutil.Factory
}

// NewCommand creates the power list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List power meter components",
		Long: `List every component on a device that meters power, with its live reading.

That is any PM or PM1 power meter, any switch, cover or light that
meters its load (Plus 1PM, Plus 2PM, Plus Plug, Pro 4PM, dimmers, RGBW
PM), any EM or EM1 energy monitor, and the meters of a Gen1 device
(Shelly 1PM, Plug S, Duo bulbs, EM).

Use 'shelly power status' with a component ID for one component in
full.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output; each item carries name, type, id, power (watts) and
the full reading under em, em1 or meter.

Columns: Device, Component, Voltage, Current, Power, Energy`,
		Example: `  # List the components that meter power on a device
  shelly power list living-room

  # Output as JSON for scripting
  shelly power list living-room -o json

  # Count the components that meter power
  shelly power list living-room -o json | jq length

  # IDs of the switch channels that meter power
  shelly power list living-room -o json | jq -r '.[] | select(.type == "switch") | .id'

  # Total power of a device in watts
  shelly power list living-room -o json | jq '[.[].power] | add'

  # Short form
  shelly power ls living-room`,
		Aliases:           []string{"ls"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	return cmdutil.RunPowerList(ctx, opts.Factory, opts.Device)
}
