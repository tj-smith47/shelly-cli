// Package export provides the energy export command.
package export

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	shellyexport "github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// NewCommand creates the energy export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		componentID   int
		componentType string
		format        string
		outputFile    string
		period        string
		from          string
		to            string
	)

	cmd := &cobra.Command{
		Use:   "export <device> [id]",
		Short: "Export energy data to file",
		Long: `Export historical energy consumption data to a file.

Supports multiple output formats:
  - CSV: Comma-separated values (default)
  - JSON: Structured JSON format
  - YAML: Human-readable YAML format

The exported data includes timestamp, voltage, current, power, and energy
measurements for the specified time range.`,
		Example: `  # Export last 24 hours as CSV
  shelly energy export shelly-3em-pro > data.csv

  # Export specific time range as JSON
  shelly energy export shelly-em --format json --from "2025-01-01" --to "2025-01-07" --output energy.json

  # Export last week as YAML
  shelly energy export shelly-3em-pro 0 --format yaml --period week --output weekly.yaml`,
		Aliases: []string{"exp", "dump"},
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &componentID); err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), f, device, componentID, componentType, format, outputFile, period, from, to)
		},
	}

	cmd.Flags().StringVar(&componentType, "type", shelly.ComponentTypeAuto, "Component type (auto, em, em1)")
	cmd.Flags().StringVarP(&format, "format", "f", shellyexport.FormatCSV, "Output format (csv, json, yaml)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&period, "period", "p", "", "Time period (hour, day, week, month)")
	cmd.Flags().StringVar(&from, "from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End time (RFC3339 or YYYY-MM-DD)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, componentType, format, outputFile, period, from, to string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Validate format
	if format != shellyexport.FormatCSV && format != shellyexport.FormatJSON && format != shellyexport.FormatYAML {
		return fmt.Errorf("invalid format: %s (use: csv, json, yaml)", format)
	}

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

	// Export based on type and format
	switch componentType {
	case shelly.ComponentTypeEM:
		return shellyexport.EMData(ctx, ios, svc.GetEMDataHistory, device, id, startTS, endTS, format, outputFile)
	case shelly.ComponentTypeEM1:
		return shellyexport.EM1Data(ctx, ios, svc.GetEM1DataHistory, device, id, startTS, endTS, format, outputFile)
	default:
		return fmt.Errorf("no energy data components found")
	}
}
