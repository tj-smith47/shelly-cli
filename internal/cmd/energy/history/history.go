// Package history provides the energy history command.
package history

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the energy history command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID   int
		componentType string
		period        string
		from          string
		to            string
		limit         int
	)

	cmd := &cobra.Command{
		Use:   "history <device> [id]",
		Short: "Show energy consumption history",
		Long: `Retrieve and display historical energy consumption data.

Shows voltage, current, power, and energy measurements stored by the
device over time (up to 60 days of 1-minute interval data).

Works with:
  - EM components (3-phase energy monitors)
  - EM1 components (single-phase energy monitors)

The device must have EMData or EM1Data components that store historical
measurements. Not all Shelly devices support historical data storage.`,
		Example: `  # Show last 24 hours of energy data
  shelly energy history shelly-3em-pro

  # Show specific time range
  shelly energy history shelly-em --from "2025-01-01" --to "2025-01-07"

  # Show last week for specific component
  shelly energy history shelly-3em-pro 0 --period week

  # Limit number of records shown
  shelly energy history shelly-em --limit 100`,
		Aliases: []string{"hist", "events"},
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &componentID); err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), f, device, componentID, componentType, period, from, to, limit)
		},
	}

	cmd.Flags().StringVar(&componentType, "type", shelly.ComponentTypeAuto, "Component type (auto, em, em1)")
	cmd.Flags().StringVarP(&period, "period", "p", "", "Time period (hour, day, week, month)")
	cmd.Flags().StringVar(&from, "from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of data points (0 = no limit)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, componentType, period, from, to string, limit int) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Calculate time range
	startTS, endTS, err := shelly.CalculateTimeRange(period, from, to)
	if err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}

	// Auto-detect type if not specified
	if componentType == shelly.ComponentTypeAuto {
		componentType, err = svc.DetectEnergyComponentType(ctx, ios, device, id)
		if err != nil {
			return err
		}
	}

	switch componentType {
	case shelly.ComponentTypeEM:
		data, err := svc.GetEMDataHistory(ctx, device, id, startTS, endTS)
		if err != nil {
			return fmt.Errorf("failed to get EMData history: %w", err)
		}
		term.DisplayEMDataHistory(ios, data, id, startTS, endTS, limit)
		return nil
	case shelly.ComponentTypeEM1:
		data, err := svc.GetEM1DataHistory(ctx, device, id, startTS, endTS)
		if err != nil {
			return fmt.Errorf("failed to get EM1Data history: %w", err)
		}
		term.DisplayEM1DataHistory(ios, data, id, startTS, endTS, limit)
		return nil
	default:
		return fmt.Errorf("no energy data components found (device may not support historical data storage)")
	}
}
