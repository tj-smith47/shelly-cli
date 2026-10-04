// Package reset provides the energy reset command.
package reset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory       *cmdutil.Factory
	Device        string
	ComponentID   int
	ComponentType string
	CounterTypes  []string
}

// NewCommand creates the energy reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f, ComponentType: shelly.ComponentTypeAuto}

	cmd := &cobra.Command{
		Use:   "reset <device> [id]",
		Short: "Reset energy monitor counters",
		Long: `Reset the accumulated energy counters of a component that meters power.

Works on EM (3-phase) energy monitors, PM and PM1 power meters, and
switches, covers and lights that meter their load (Plus 1PM, Plus 2PM,
Plug, dimmers, RGBW PM). The component is chosen as 'shelly energy
status' chooses it: the first one that meters power, or the one with
the given ID and --type.

EM1 energy monitors and Gen1 meters have no counter reset.`,
		Example: `  # Reset all counters for EM component 0
  shelly energy reset shelly-3em-pro 0

  # Reset specific counter types
  shelly energy reset shelly-3em-pro 0 --types active,reactive

  # Reset the energy total of switch channel 1 on a Plus 2PM
  shelly energy reset kitchen 1 --type switch

  # Reset with device alias
  shelly energy reset basement-em`,
		Aliases:           []string{"clear"},
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.DeviceThenNoComplete(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opts.Device, opts.ComponentID, err = cmdutil.ParsePowerStatusArgs(args, false)
			if err != nil {
				return err
			}
			return run(cmd.Context(), opts)
		},
	}

	cmdutil.AddPowerTypeFlag(cmd, &opts.ComponentType)
	cmd.Flags().StringSliceVar(&opts.CounterTypes, "types", nil, "Counter types to reset (leave empty for all)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Resetting energy counters...", func(ctx context.Context) error {
		r, err := svc.ResetPowerCounters(ctx, opts.Device, opts.ComponentType, opts.ComponentID, opts.CounterTypes)
		if err != nil {
			return fmt.Errorf("failed to reset energy counters: %w", err)
		}
		ios.Success("Energy counters reset for %s #%d", term.MeterLabel(r.Type), r.ID)
		return nil
	})
}
