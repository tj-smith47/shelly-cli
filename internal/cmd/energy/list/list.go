// Package list provides the energy list command.
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

// NewCommand creates the energy list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "list <device>",
		Short: "List energy monitoring components",
		Long: `List every component on a device that meters power, with its live reading.

That is any EM (3-phase) or EM1 (single-phase) energy monitor, any PM
or PM1 power meter, any switch, cover or light that meters its load
(Plus 1PM, Plus 2PM, Plus Plug, Pro 4PM, dimmers, RGBW PM), and the
meters of a Gen1 device (Shelly 1PM, Plug S, Duo bulbs, EM). Energy
monitors are listed first.

Use 'shelly energy status' with a component ID for one component in
full.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output; each item carries name, type, id, power (watts) and
the full reading under em, em1 or meter.

Columns: Device, Component, Voltage, Current, Power, Energy`,
		Example: `  # List the components that meter power on a device
  shelly energy list shelly-3em-pro

  # Output as JSON for scripting
  shelly energy list shelly-3em-pro -o json

  # IDs of the 3-phase monitors
  shelly energy list shelly-3em-pro -o json | jq -r '.[] | select(.type == "em") | .id'

  # Count the components that meter power
  shelly energy list shelly-3em-pro -o json | jq length

  # Short form
  shelly energy ls shelly-3em-pro`,
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
