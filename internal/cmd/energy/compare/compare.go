// Package compare provides the energy compare command.
package compare

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the energy compare command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		devices []string
		period  string
		from    string
		to      string
	)

	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare energy usage between devices",
		Long: `Compare energy consumption across multiple devices for a specified time period.

Shows each device's total energy consumption, average power, and percentage
of the total consumption. Useful for identifying high-energy consumers.

By default, compares all registered devices. Use --devices to specify a subset.`,
		Example: `  # Compare all devices for the last day
  shelly energy compare

  # Compare specific devices for the last week
  shelly energy compare --devices kitchen,living-room,garage --period week

  # Compare for a specific date range
  shelly energy compare --from "2025-01-01" --to "2025-01-07"

  # Output as JSON
  shelly energy compare -o json`,
		Aliases: []string{"cmp", "diff"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, devices, period, from, to)
		},
	}

	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to compare (default: all registered)")
	cmd.Flags().StringVarP(&period, "period", "p", "day", "Time period (hour, day, week, month)")
	cmd.Flags().StringVar(&from, "from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End time (RFC3339 or YYYY-MM-DD)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, period, from, to string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(devices) == 0 {
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices found. Register devices using 'shelly device add' or specify --devices")
		return nil
	}

	if len(devices) < 2 {
		ios.Warning("At least 2 devices are required for comparison. Found: %d", len(devices))
		return nil
	}

	sort.Strings(devices)

	startTS, endTS, err := shelly.CalculateTimeRange(period, from, to)
	if err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}

	// Collect comparison data using service layer
	comparison := svc.CollectComparisonData(ctx, ios, devices, period, startTS, endTS)

	// Calculate percentages
	if comparison.TotalEnergy > 0 {
		for i := range comparison.Devices {
			comparison.Devices[i].Percentage = (comparison.Devices[i].Energy / comparison.TotalEnergy) * 100
		}
	}

	return cmdutil.PrintResult(ios, comparison, term.DisplayComparison)
}
