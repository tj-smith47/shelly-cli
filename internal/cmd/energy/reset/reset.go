// Package reset provides the energy reset command.
package reset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds command options.
type Options struct {
	Factory      *cmdutil.Factory
	Device       string
	ComponentID  int
	CounterTypes []string
}

// NewCommand creates the energy reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "reset <device> [id]",
		Short: "Reset energy monitor counters",
		Long: `Reset energy counters for an EM (3-phase) energy monitor.

Note: Only EM components support counter reset. EM1 components
do not have a reset capability.`,
		Example: `  # Reset all counters for EM component 0
  shelly energy reset shelly-3em-pro 0

  # Reset specific counter types
  shelly energy reset shelly-3em-pro 0 --types active,reactive

  # Reset with device alias
  shelly energy reset basement-em`,
		Aliases: []string{"clear"},
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &opts.ComponentID); err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.CounterTypes, "types", nil, "Counter types to reset (leave empty for all)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Resetting energy counters...", func(ctx context.Context) error {
		if err := svc.ResetEMCounters(ctx, opts.Device, opts.ComponentID, opts.CounterTypes); err != nil {
			return fmt.Errorf("failed to reset EM counters: %w", err)
		}
		ios.Success("Energy counters reset for EM #%d", opts.ComponentID)
		return nil
	})
}
