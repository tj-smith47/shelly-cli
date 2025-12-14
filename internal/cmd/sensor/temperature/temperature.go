// Package temperature provides temperature sensor commands.
package temperature

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the temperature command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "temperature",
		Aliases: []string{"temp", "t"},
		Short:   "Manage temperature sensors",
		Long: `Manage temperature sensors on Shelly devices.

Temperature sensors can be built-in or external (DS18B20).
Readings are provided in both Celsius and Fahrenheit.`,
		Example: `  # List temperature sensors
  shelly sensor temperature list living-room

  # Get temperature reading
  shelly sensor temperature status living-room`,
	}

	cmd.AddCommand(newListCommand(f))
	cmd.AddCommand(newStatusCommand(f))

	return cmd
}

// ListOptions holds list command options.
type ListOptions struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

func newListCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &ListOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List temperature sensors",
		Long:    `List all temperature sensors on a Shelly device.`,
		Example: `  # List temperature sensors
  shelly sensor temperature list living-room

  # Output as JSON
  shelly sensor temperature list living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runList(ctx context.Context, opts *ListOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return fmt.Errorf("failed to get device status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var fullStatus map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &fullStatus); err != nil {
		return fmt.Errorf("failed to parse status: %w", err)
	}

	sensors := collectTempSensors(fullStatus, ios)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No temperature sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Temperature Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.TC != nil {
			ios.Printf("    Temperature: %.1f°C", *s.TC)
			if s.TF != nil {
				ios.Printf(" (%.1f°F)", *s.TF)
			}
			ios.Println()
		}
	}

	return nil
}

// StatusOptions holds status command options.
type StatusOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	JSON    bool
}

func newStatusCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &StatusOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "get", "read"},
		Short:   "Get temperature reading",
		Long:    `Get the current temperature reading from a sensor.`,
		Example: `  # Get temperature from default sensor (ID 0)
  shelly sensor temperature status living-room

  # Get temperature from specific sensor
  shelly sensor temperature status living-room --id 1

  # Output as JSON
  shelly sensor temperature status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runStatus(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Sensor ID (default 0)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func runStatus(ctx context.Context, opts *StatusOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	params := map[string]any{"id": opts.ID}
	result, err := conn.Call(ctx, "Temperature.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get temperature status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var status TempStatus
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return fmt.Errorf("failed to parse status: %w", err)
	}

	if opts.JSON {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Temperature Sensor %d:", opts.ID)))
	ios.Println()
	if status.TC != nil {
		ios.Printf("  Temperature: %s", theme.Highlight().Render(fmt.Sprintf("%.1f°C", *status.TC)))
		if status.TF != nil {
			ios.Printf(" (%s)", theme.Dim().Render(fmt.Sprintf("%.1f°F", *status.TF)))
		}
		ios.Println()
	} else {
		ios.Warning("No temperature reading available.")
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", status.Errors)
	}

	return nil
}

// TempStatus represents temperature sensor status.
type TempStatus struct {
	ID     int      `json:"id"`
	TC     *float64 `json:"tC"`
	TF     *float64 `json:"tF"`
	Errors []string `json:"errors,omitempty"`
}

func collectTempSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []TempStatus {
	return shelly.CollectSensorsByPrefix[TempStatus](status, "temperature:", ios)
}
