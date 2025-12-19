// Package history provides the energy history command.
package history

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		return showEMDataHistory(ctx, ios, svc, device, id, startTS, endTS, limit)
	case shelly.ComponentTypeEM1:
		return showEM1DataHistory(ctx, ios, svc, device, id, startTS, endTS, limit)
	default:
		return fmt.Errorf("no energy data components found (device may not support historical data storage)")
	}
}

func showEMDataHistory(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int, startTS, endTS *int64, limit int) error {
	data, err := svc.GetEMDataHistory(ctx, device, id, startTS, endTS)
	if err != nil {
		return fmt.Errorf("failed to get EMData history: %w", err)
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, data)
	}

	// Human-readable format
	ios.Printf("Energy History (EM) #%d\n", id)
	if startTS != nil {
		ios.Printf("From: %s\n", time.Unix(*startTS, 0).Format(time.RFC3339))
	}
	if endTS != nil {
		ios.Printf("To:   %s\n", time.Unix(*endTS, 0).Format(time.RFC3339))
	}
	ios.Printf("\n")

	totalRecords := 0
	for _, block := range data.Data {
		totalRecords += len(block.Values)
	}

	if totalRecords == 0 {
		ios.Warning("No data available for the specified time range")
		return nil
	}

	ios.Printf("Total data points: %d\n", totalRecords)
	ios.Printf("Data blocks: %d\n\n", len(data.Data))

	count := 0
	for _, block := range data.Data {
		blockTime := time.Unix(block.TS, 0)
		for i, values := range block.Values {
			if limit > 0 && count >= limit {
				ios.Printf("\n(showing first %d of %d records, use --limit to see more)\n", limit, totalRecords)
				return nil
			}

			recordTime := blockTime.Add(time.Duration(i*block.Period) * time.Second)
			ios.Printf("[%s] Total: %.2fW (A: %.2fW, B: %.2fW, C: %.2fW) | Voltage: A=%.1fV B=%.1fV C=%.1fV\n",
				recordTime.Format("2006-01-02 15:04:05"),
				values.TotalActivePower,
				values.AActivePower,
				values.BActivePower,
				values.CActivePower,
				values.AVoltage,
				values.BVoltage,
				values.CVoltage,
			)
			count++
		}
	}

	// Calculate total energy consumption if possible
	if count > 0 {
		totalEnergy, _, _, _ := shelly.CalculateEMMetrics(data)
		ios.Printf("\nEstimated total energy consumption: %.2f kWh\n", totalEnergy)
	}

	return nil
}

func showEM1DataHistory(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int, startTS, endTS *int64, limit int) error {
	data, err := svc.GetEM1DataHistory(ctx, device, id, startTS, endTS)
	if err != nil {
		return fmt.Errorf("failed to get EM1Data history: %w", err)
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, data)
	}

	// Human-readable format
	ios.Printf("Energy History (EM1) #%d\n", id)
	if startTS != nil {
		ios.Printf("From: %s\n", time.Unix(*startTS, 0).Format(time.RFC3339))
	}
	if endTS != nil {
		ios.Printf("To:   %s\n", time.Unix(*endTS, 0).Format(time.RFC3339))
	}
	ios.Printf("\n")

	totalRecords := 0
	for _, block := range data.Data {
		totalRecords += len(block.Values)
	}

	if totalRecords == 0 {
		ios.Warning("No data available for the specified time range")
		return nil
	}

	ios.Printf("Total data points: %d\n", totalRecords)
	ios.Printf("Data blocks: %d\n\n", len(data.Data))

	count := 0
	for _, block := range data.Data {
		blockTime := time.Unix(block.TS, 0)
		for i, values := range block.Values {
			if limit > 0 && count >= limit {
				ios.Printf("\n(showing first %d of %d records, use --limit to see more)\n", limit, totalRecords)
				return nil
			}

			recordTime := blockTime.Add(time.Duration(i*block.Period) * time.Second)
			pf := ""
			if values.PowerFactor != nil {
				pf = fmt.Sprintf(" | PF: %.3f", *values.PowerFactor)
			}
			ios.Printf("[%s] Power: %.2fW | Voltage: %.1fV | Current: %.2fA%s\n",
				recordTime.Format("2006-01-02 15:04:05"),
				values.ActivePower,
				values.Voltage,
				values.Current,
				pf,
			)
			count++
		}
	}

	// Calculate total energy consumption if possible
	if count > 0 {
		totalEnergy, _, _, _ := shelly.CalculateEM1Metrics(data)
		ios.Printf("\nEstimated total energy consumption: %.2f kWh\n", totalEnergy)
	}

	return nil
}
