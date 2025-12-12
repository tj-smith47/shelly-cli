// Package reset provides the energy reset command.
package reset

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the energy reset command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID  int
		counterTypes []string
	)

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
			device := args[0]
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &componentID); err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), f, device, componentID, counterTypes)
		},
	}

	cmd.Flags().StringSliceVar(&counterTypes, "types", nil, "Counter types to reset (leave empty for all)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, counterTypes []string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Resetting energy counters...", func(ctx context.Context) error {
		if err := svc.ResetEMCounters(ctx, device, id, counterTypes); err != nil {
			return fmt.Errorf("failed to reset EM counters: %w", err)
		}
		ios.Success("Energy counters reset for EM #%d", id)
		return nil
	})
}
