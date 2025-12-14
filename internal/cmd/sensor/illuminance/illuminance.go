// Package illuminance provides illuminance sensor commands.
package illuminance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/sensorutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the illuminance command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "illuminance",
		Aliases: []string{"lux", "light-level", "brightness"},
		Short:   "Manage illuminance sensors",
		Long: `Manage illuminance (light level) sensors on Shelly devices.

Illuminance sensors provide light level readings in lux,
useful for automation based on ambient light conditions.`,
		Example: `  # List illuminance sensors
  shelly sensor illuminance list living-room

  # Get current light level
  shelly sensor illuminance status living-room`,
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
		Short:   "List illuminance sensors",
		Long:    `List all illuminance sensors on a Shelly device.`,
		Example: `  # List illuminance sensors
  shelly sensor illuminance list living-room

  # Output as JSON
  shelly sensor illuminance list living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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

	sensors := collectIlluminanceSensors(fullStatus, ios)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No illuminance sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Illuminance Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.Lux != nil {
			ios.Printf("    Light Level: %.0f lux\n", *s.Lux)
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
		Short:   "Get light level reading",
		Long:    `Get the current light level from an illuminance sensor.`,
		Example: `  # Get light level from default sensor (ID 0)
  shelly sensor illuminance status living-room

  # Get from specific sensor
  shelly sensor illuminance status living-room --id 1

  # Output as JSON
  shelly sensor illuminance status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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
	result, err := conn.Call(ctx, "Illuminance.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get illuminance status: %w", err)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Illuminance Sensor %d:", opts.ID)))
	ios.Println()
	if status.Lux != nil {
		level := getLightLevel(*status.Lux)
		ios.Printf("  Light Level: %s (%s)\n",
			theme.Highlight().Render(fmt.Sprintf("%.0f lux", *status.Lux)),
			theme.Dim().Render(level))
	} else {
		ios.Warning("No illuminance reading available.")
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", status.Errors)
	}

	return nil
}

// getLightLevel returns a human-readable description of the light level.
func getLightLevel(lux float64) string {
	switch {
	case lux < 1:
		return "Very dark"
	case lux < 50:
		return "Dark"
	case lux < 200:
		return "Dim"
	case lux < 500:
		return "Indoor light"
	case lux < 1000:
		return "Bright indoor"
	case lux < 10000:
		return "Overcast daylight"
	case lux < 25000:
		return "Daylight"
	default:
		return "Direct sunlight"
	}
}

// Status represents illuminance sensor status.
type Status struct {
	ID     int      `json:"id"`
	Lux    *float64 `json:"lux"`
	Errors []string `json:"errors,omitempty"`
}

func collectIlluminanceSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []Status {
	return sensorutil.CollectByPrefix[Status](status, "illuminance:", ios)
}
