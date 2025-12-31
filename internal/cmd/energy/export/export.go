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

// Options holds command options.
type Options struct {
	Factory       *cmdutil.Factory
	Device        string
	ComponentID   int
	ComponentType string
	Format        string
	OutputFile    string
	Period        string
	From          string
	To            string
}

// NewCommand creates the energy export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:       f,
		ComponentType: shelly.ComponentTypeAuto,
		Format:        shellyexport.FormatCSV,
	}

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
			opts.Device = args[0]
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &opts.ComponentID); err != nil {
					return fmt.Errorf("invalid component ID: %w", err)
				}
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.ComponentType, "type", shelly.ComponentTypeAuto, "Component type (auto, em, em1)")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", shellyexport.FormatCSV, "Output format (csv, json, yaml)")
	cmd.Flags().StringVarP(&opts.OutputFile, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&opts.Period, "period", "p", "", "Time period (hour, day, week, month)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.To, "to", "", "End time (RFC3339 or YYYY-MM-DD)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate format
	if opts.Format != shellyexport.FormatCSV && opts.Format != shellyexport.FormatJSON && opts.Format != shellyexport.FormatYAML {
		return fmt.Errorf("invalid format: %s (use: csv, json, yaml)", opts.Format)
	}

	// Calculate time range
	startTS, endTS, err := shelly.CalculateTimeRange(opts.Period, opts.From, opts.To)
	if err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}

	// Auto-detect type if not specified
	componentType := opts.ComponentType
	if componentType == shelly.ComponentTypeAuto {
		componentType, err = svc.DetectEnergyComponentType(ctx, ios, opts.Device, opts.ComponentID)
		if err != nil {
			return err
		}
	}

	// Export based on type and format
	switch componentType {
	case shelly.ComponentTypeEM:
		return shellyexport.EMData(ctx, ios, svc.GetEMDataHistory, opts.Device, opts.ComponentID, startTS, endTS, opts.Format, opts.OutputFile)
	case shelly.ComponentTypeEM1:
		return shellyexport.EM1Data(ctx, ios, svc.GetEM1DataHistory, opts.Device, opts.ComponentID, startTS, endTS, opts.Format, opts.OutputFile)
	default:
		return fmt.Errorf("no energy data components found")
	}
}
