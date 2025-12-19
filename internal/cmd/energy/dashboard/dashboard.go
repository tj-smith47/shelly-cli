// Package dashboard provides the energy dashboard command.
package dashboard

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the energy dashboard command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		devices      []string
		costPerKwh   float64
		costCurrency string
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show energy dashboard for all devices",
		Long: `Display an aggregated energy dashboard showing power consumption across all devices.

Shows total power consumption, per-device breakdown, and optional cost estimation.
By default, queries all registered devices. Use --devices to specify a subset.

Examples:
  # Show dashboard for all registered devices
  shelly energy dashboard

  # Show dashboard for specific devices
  shelly energy dashboard --devices kitchen,living-room,garage

  # Include cost estimation at $0.12 per kWh
  shelly energy dashboard --cost 0.12 --currency USD`,
		Aliases: []string{"dash", "summary"},
		Example: `  # Show dashboard for all registered devices
  shelly energy dashboard

  # Show dashboard for specific devices
  shelly energy dashboard --devices kitchen,living-room

  # Include cost estimation
  shelly energy dashboard --cost 0.15 --currency EUR

  # Output as JSON
  shelly energy dashboard -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, devices, costPerKwh, costCurrency)
		},
	}

	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().Float64Var(&costPerKwh, "cost", 0, "Cost per kWh for estimation")
	cmd.Flags().StringVar(&costCurrency, "currency", "USD", "Currency for cost display")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, costPerKwh float64, currency string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get device list
	if len(devices) == 0 {
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices found. Register devices using 'shelly device add' or specify --devices")
		return nil
	}

	sort.Strings(devices)

	// Collect data from all devices concurrently using service layer
	dashboard := svc.CollectDashboardData(ctx, ios, devices)

	// Add cost estimation if configured
	if costPerKwh > 0 {
		dashboard.CostPerKwh = costPerKwh
		dashboard.CostCurrency = currency
		cost := (dashboard.TotalEnergy / 1000) * costPerKwh
		dashboard.EstimatedCost = &cost
	}

	return cmdutil.PrintResult(ios, dashboard, term.DisplayDashboard)
}
