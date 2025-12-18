// Package status provides the sensor status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the sensor status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "all", "readings"},
		Short:   "Show all sensor readings",
		Long: `Show all sensor readings from a Shelly device.

Displays a combined view of all available sensors including:
- Temperature (°C/°F)
- Humidity (%)
- Flood detection status
- Smoke detection status
- Illuminance (lux)
- Voltage readings

Only sensors present on the device will be shown.`,
		Example: `  # Show all sensor readings
  shelly sensor status living-room

  # Output as JSON
  shelly sensor status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Get full device status
	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	// Marshal result to JSON for parsing
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var fullStatus map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &fullStatus); err != nil {
		return fmt.Errorf("failed to parse status: %w", err)
	}

	// Collect sensor data using service layer
	data := shelly.CollectSensorData(fullStatus)

	if opts.JSON {
		jsonOut, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonOut))
		return nil
	}

	cmdutil.DisplayAllSensorData(ios, data, opts.Device)
	return nil
}
