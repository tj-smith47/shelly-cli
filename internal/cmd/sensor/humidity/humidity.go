// Package humidity provides humidity sensor commands.
package humidity

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

// NewCommand creates the humidity command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "humidity",
		Aliases: []string{"humid", "rh"},
		Short:   "Manage humidity sensors",
		Long: `Manage humidity sensors on Shelly devices.

Humidity sensors (DHT22, HTU21D, or similar) provide relative humidity readings.`,
		Example: `  # List humidity sensors
  shelly sensor humidity list living-room

  # Get humidity reading
  shelly sensor humidity status living-room`,
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
		Short:   "List humidity sensors",
		Long:    `List all humidity sensors on a Shelly device.`,
		Example: `  # List humidity sensors
  shelly sensor humidity list living-room

  # Output as JSON
  shelly sensor humidity list living-room --json`,
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

	sensors := collectHumiditySensors(fullStatus, ios)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No humidity sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Humidity Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.RH != nil {
			ios.Printf("    Humidity: %.1f%%\n", *s.RH)
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
		Short:   "Get humidity reading",
		Long:    `Get the current humidity reading from a sensor.`,
		Example: `  # Get humidity from default sensor (ID 0)
  shelly sensor humidity status living-room

  # Get humidity from specific sensor
  shelly sensor humidity status living-room --id 1

  # Output as JSON
  shelly sensor humidity status living-room --json`,
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
	result, err := conn.Call(ctx, "Humidity.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get humidity status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var status Status
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Humidity Sensor %d:", opts.ID)))
	ios.Println()
	if status.RH != nil {
		ios.Printf("  Humidity: %s\n", theme.Highlight().Render(fmt.Sprintf("%.1f%%", *status.RH)))
	} else {
		ios.Warning("No humidity reading available.")
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", status.Errors)
	}

	return nil
}

// Status represents humidity sensor status.
type Status struct {
	ID     int      `json:"id"`
	RH     *float64 `json:"rh"`
	Errors []string `json:"errors,omitempty"`
}

func collectHumiditySensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []Status {
	return shelly.CollectSensorsByPrefix[Status](status, "humidity:", ios)
}
