// Package status provides the power status command.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	cmdutil.PowerStatusOptions
}

// NewCommand creates the power status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:            f,
		PowerStatusOptions: cmdutil.PowerStatusOptions{Type: shelly.ComponentTypeAuto},
	}

	cmd := &cobra.Command{
		Use:   "status [device] [id]",
		Short: "Show power meter status",
		Long: `Show the live power reading of one component on a device.

Reads whichever component meters power: a PM or PM1 power meter, a
switch, cover or light that meters its load (Plus 1PM, Plus 2PM, Plus
Plug, Pro 4PM, dimmers, RGBW PM), an EM or EM1 energy monitor, or the
meters of a Gen1 device (Shelly 1PM, Plug S, Duo bulbs, EM). Shows
voltage, current, power, frequency and accumulated energy, as far as
the component reports them.

Without an ID the first component that meters power is shown; 'shelly
power list' lists them all. Use --type when two component types share
an ID.

With --all, shows every power reading on every registered device as one
list. Devices that are offline or meter nothing are skipped with a note
on stderr. With -o json or -o yaml each reading carries name, type, id,
power (watts) and the full reading under em, em1 or meter.`,
		Example: `  # Show the first power reading of a device
  shelly power status living-room

  # Show channel 1 of a two-channel switch
  shelly power status living-room 1

  # Pick the component type explicitly
  shelly power status living-room 0 --type pm1

  # Output as JSON for scripting
  shelly power status living-room -o json

  # Read just the power in watts
  shelly power status living-room -o json | jq '.power'

  # Every power reading on every registered device
  shelly power status --all`,
		Aliases:           []string{"st"},
		Args:              cobra.MaximumNArgs(2),
		ValidArgsFunction: completion.DeviceThenNoComplete(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opts.Device, opts.ID, err = cmdutil.ParsePowerStatusArgs(args, opts.All)
			if err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmdutil.AddPowerTypeFlag(cmd, &opts.Type)
	flags.AddAllFlag(cmd, &opts.All)
	cmd.MarkFlagsMutuallyExclusive("all", "type")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	return cmdutil.RunPowerStatus(ctx, opts.Factory, opts.PowerStatusOptions)
}
