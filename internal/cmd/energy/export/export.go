// Package export provides the energy export command.
package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const (
	formatCSV  = "csv"
	formatJSON = "json"
	formatYAML = "yaml"

	typeAuto = "auto"
	typeEM   = "em"
	typeEM1  = "em1"
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

	cmd.Flags().StringVar(&componentType, "type", typeAuto, "Component type (auto, em, em1)")
	cmd.Flags().StringVarP(&format, "format", "f", formatCSV, "Output format (csv, json, yaml)")
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
	if format != formatCSV && format != formatJSON && format != formatYAML {
		return fmt.Errorf("invalid format: %s (use: csv, json, yaml)", format)
	}

	// Calculate time range
	startTS, endTS, err := calculateTimeRange(period, from, to)
	if err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}

	// Auto-detect type if not specified
	if componentType == typeAuto {
		componentType, err = detectEnergyDataComponentType(ctx, ios, svc, device, id)
		if err != nil {
			return err
		}
	}

	// Export based on type and format
	switch componentType {
	case typeEM:
		return exportEMData(ctx, ios, svc, device, id, startTS, endTS, format, outputFile)
	case typeEM1:
		return exportEM1Data(ctx, ios, svc, device, id, startTS, endTS, format, outputFile)
	default:
		return fmt.Errorf("no energy data components found")
	}
}

func detectEnergyDataComponentType(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int) (string, error) {
	// Try EMData first
	emRecords, err := svc.GetEMDataRecords(ctx, device, id, nil)
	if err == nil && emRecords != nil && len(emRecords.Records) > 0 {
		return typeEM, nil
	}
	ios.DebugErr("get EMData records", err)

	// Try EM1Data
	em1Records, err := svc.GetEM1DataRecords(ctx, device, id, nil)
	if err == nil && em1Records != nil && len(em1Records.Records) > 0 {
		return typeEM1, nil
	}
	ios.DebugErr("get EM1Data records", err)

	return "", fmt.Errorf("no energy data components found")
}

func calculateTimeRange(period, from, to string) (startTS, endTS *int64, err error) {
	// If explicit from/to provided, use those
	if from != "" || to != "" {
		startTS, endTS, err = parseExplicitTimeRange(from, to)
		return startTS, endTS, err
	}

	// Calculate based on period
	now := time.Now()
	var start time.Time

	switch period {
	case "hour":
		start = now.Add(-1 * time.Hour)
	case "day", "":
		start = now.Add(-24 * time.Hour)
	case "week":
		start = now.Add(-7 * 24 * time.Hour)
	case "month":
		start = now.Add(-30 * 24 * time.Hour)
	default:
		return nil, nil, fmt.Errorf("invalid period: %s (use: hour, day, week, month)", period)
	}

	startUnix := start.Unix()
	endUnix := now.Unix()
	return &startUnix, &endUnix, nil
}

func parseExplicitTimeRange(from, to string) (startTS, endTS *int64, err error) {
	if from != "" {
		t, err := parseTime(from)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --from time: %w", err)
		}
		ts := t.Unix()
		startTS = &ts
	}
	if to != "" {
		t, err := parseTime(to)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --to time: %w", err)
		}
		ts := t.Unix()
		endTS = &ts
	}
	return startTS, endTS, nil
}

func parseTime(s string) (time.Time, error) {
	// Try RFC3339 first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	// Try date-only format (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}

	// Try datetime format (YYYY-MM-DD HH:MM:SS)
	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time (use RFC3339, YYYY-MM-DD, or 'YYYY-MM-DD HH:MM:SS')")
}

func exportEMData(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int, startTS, endTS *int64, format, outputFile string) error {
	// Fetch data
	data, err := svc.GetEMDataHistory(ctx, device, id, startTS, endTS)
	if err != nil {
		return fmt.Errorf("failed to get EMData: %w", err)
	}

	// Export based on format
	switch format {
	case formatCSV:
		return exportEMDataCSV(ios, data, outputFile)
	case formatJSON:
		return exportToFile(ios, data, outputFile, output.FormatJSON, "JSON")
	case formatYAML:
		return exportToFile(ios, data, outputFile, output.FormatYAML, "YAML")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func exportEM1Data(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, id int, startTS, endTS *int64, format, outputFile string) error {
	// Fetch data
	data, err := svc.GetEM1DataHistory(ctx, device, id, startTS, endTS)
	if err != nil {
		return fmt.Errorf("failed to get EM1Data: %w", err)
	}

	// Export based on format
	switch format {
	case formatCSV:
		return exportEM1DataCSV(ios, data, outputFile)
	case formatJSON:
		return exportToFile(ios, data, outputFile, output.FormatJSON, "JSON")
	case formatYAML:
		return exportToFile(ios, data, outputFile, output.FormatYAML, "YAML")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func exportEMDataCSV(ios *iostreams.IOStreams, data *components.EMDataGetDataResult, outputFile string) error {
	if data == nil || len(data.Data) == 0 {
		return fmt.Errorf("no data to export")
	}

	writer, closer, err := getWriter(ios, outputFile)
	if err != nil {
		return err
	}
	defer closer()

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV headers for 3-phase energy meter
	headers := []string{
		"timestamp",
		"a_voltage", "a_current", "a_act_power", "a_aprt_power", "a_pf", "a_freq",
		"b_voltage", "b_current", "b_act_power", "b_aprt_power", "b_pf", "b_freq",
		"c_voltage", "c_current", "c_act_power", "c_aprt_power", "c_pf", "c_freq",
		"total_current", "total_act_power", "total_aprt_power",
		"total_act_energy", "total_act_ret_energy",
		"n_current",
	}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, block := range data.Data {
		ts := block.TS
		period := int64(block.Period)

		for i, v := range block.Values {
			// Calculate timestamp for this measurement
			measurementTS := ts + int64(i)*period
			timeStr := time.Unix(measurementTS, 0).UTC().Format(time.RFC3339)

			row := []string{
				timeStr,
				formatFloat(v.AVoltage),
				formatFloat(v.ACurrent),
				formatFloat(v.AActivePower),
				formatFloat(v.AApparentPower),
				formatFloatPtr(v.APowerFactor),
				formatFloatPtr(v.AFreq),
				formatFloat(v.BVoltage),
				formatFloat(v.BCurrent),
				formatFloat(v.BActivePower),
				formatFloat(v.BApparentPower),
				formatFloatPtr(v.BPowerFactor),
				formatFloatPtr(v.BFreq),
				formatFloat(v.CVoltage),
				formatFloat(v.CCurrent),
				formatFloat(v.CActivePower),
				formatFloat(v.CApparentPower),
				formatFloatPtr(v.CPowerFactor),
				formatFloatPtr(v.CFreq),
				formatFloat(v.TotalCurrent),
				formatFloat(v.TotalActivePower),
				formatFloat(v.TotalAprtPower),
				formatFloatPtr(v.TotalActEnergy),
				formatFloatPtr(v.TotalActRetEnergy),
				formatFloatPtr(v.NCurrent),
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	if outputFile != "" {
		ios.Success("Exported to %s (CSV)", outputFile)
	}
	return nil
}

func exportEM1DataCSV(ios *iostreams.IOStreams, data *components.EM1DataGetDataResult, outputFile string) error {
	if data == nil || len(data.Data) == 0 {
		return fmt.Errorf("no data to export")
	}

	writer, closer, err := getWriter(ios, outputFile)
	if err != nil {
		return err
	}
	defer closer()

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV headers for single-phase energy meter
	headers := []string{
		"timestamp",
		"voltage", "current", "act_power", "aprt_power", "pf", "freq",
		"act_energy", "act_ret_energy",
	}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, block := range data.Data {
		ts := block.TS
		period := int64(block.Period)

		for i, v := range block.Values {
			// Calculate timestamp for this measurement
			measurementTS := ts + int64(i)*period
			timeStr := time.Unix(measurementTS, 0).UTC().Format(time.RFC3339)

			row := []string{
				timeStr,
				formatFloat(v.Voltage),
				formatFloat(v.Current),
				formatFloat(v.ActivePower),
				formatFloat(v.ApparentPower),
				formatFloatPtr(v.PowerFactor),
				formatFloatPtr(v.Freq),
				formatFloatPtr(v.ActEnergy),
				formatFloatPtr(v.ActRetEnergy),
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	if outputFile != "" {
		ios.Success("Exported to %s (CSV)", outputFile)
	}
	return nil
}

func exportToFile(ios *iostreams.IOStreams, data any, outputFile string, format output.Format, formatName string) error {
	writer, closer, err := getWriter(ios, outputFile)
	if err != nil {
		return err
	}
	defer closer()

	formatter := output.NewFormatter(format)
	if err := formatter.Format(writer, data); err != nil {
		return fmt.Errorf("failed to encode %s: %w", formatName, err)
	}

	if outputFile != "" {
		ios.Success("Exported to %s (%s)", outputFile, formatName)
	}
	return nil
}

func getWriter(ios *iostreams.IOStreams, outputFile string) (io.Writer, func(), error) {
	if outputFile == "" {
		return ios.Out, func() {}, nil
	}

	//nolint:gosec // G304: User-provided file path is expected for CLI export functionality
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return file, func() {
		if err := file.Close(); err != nil {
			ios.DebugErr("close output file", err)
		}
	}, nil
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func formatFloatPtr(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}
