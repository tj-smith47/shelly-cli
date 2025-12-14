// Package flood provides flood sensor commands.
package flood

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

// NewCommand creates the flood command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flood",
		Aliases: []string{"water", "leak"},
		Short:   "Manage flood sensors",
		Long: `Manage flood (water leak) sensors on Shelly devices.

Flood sensors detect water leaks and can trigger alarms with
different modes: disabled, normal, intense, or rain detection.`,
		Example: `  # List flood sensors
  shelly sensor flood list bathroom

  # Check flood status
  shelly sensor flood status bathroom

  # Test flood alarm
  shelly sensor flood test bathroom`,
	}

	cmd.AddCommand(newListCommand(f))
	cmd.AddCommand(newStatusCommand(f))
	cmd.AddCommand(newTestCommand(f))

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
		Short:   "List flood sensors",
		Long:    `List all flood sensors on a Shelly device.`,
		Example: `  # List flood sensors
  shelly sensor flood list bathroom

  # Output as JSON
  shelly sensor flood list bathroom --json`,
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

	sensors := collectFloodSensors(fullStatus, ios)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No flood sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Flood Sensors:"))
	ios.Println()
	for _, s := range sensors {
		status := theme.StatusOK().Render("Clear")
		if s.Alarm {
			status = theme.StatusError().Render("WATER DETECTED!")
		}
		muteStr := ""
		if s.Mute {
			muteStr = " " + theme.Dim().Render("(muted)")
		}
		ios.Printf("  Sensor %d: %s%s\n", s.ID, status, muteStr)
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
		Aliases: []string{"st", "get", "check"},
		Short:   "Check flood status",
		Long:    `Get the current status of a flood sensor.`,
		Example: `  # Check flood status
  shelly sensor flood status bathroom

  # Check specific sensor
  shelly sensor flood status bathroom --id 1

  # Output as JSON
  shelly sensor flood status bathroom --json`,
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
	result, err := conn.Call(ctx, "Flood.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get flood status: %w", err)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Flood Sensor %d:", opts.ID)))
	ios.Println()

	if status.Alarm {
		ios.Printf("  Status: %s\n", theme.StatusError().Render("WATER DETECTED!"))
	} else {
		ios.Printf("  Status: %s\n", theme.StatusOK().Render("Clear"))
	}

	if status.Mute {
		ios.Printf("  Alarm: %s\n", theme.Dim().Render("Muted"))
	} else {
		ios.Printf("  Alarm: %s\n", theme.Highlight().Render("Active"))
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", status.Errors)
	}

	return nil
}

// TestOptions holds test command options.
type TestOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newTestCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &TestOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "test <device>",
		Aliases: []string{"trigger"},
		Short:   "Test flood alarm",
		Long: `Test the flood alarm on a Shelly device.

Note: The Flood component may not have a dedicated test method.
This command provides instructions for manual testing.`,
		Example: `  # Test flood alarm
  shelly sensor flood test bathroom`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runTest(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Sensor ID (default 0)")

	return cmd
}

func runTest(_ context.Context, opts *TestOptions) error {
	ios := opts.Factory.IOStreams()

	// The Flood component doesn't have a Test method like Smoke does
	// Provide instructions for manual testing
	ios.Info("Flood sensor testing:")
	ios.Println()
	ios.Println("  The Flood component does not have a programmatic test method.")
	ios.Println("  To test the flood sensor, briefly apply water to the sensor")
	ios.Println("  contacts or use the device's physical test button if available.")
	ios.Println()
	ios.Println("  Monitor status with:")
	ios.Printf("    shelly sensor flood status %s --id %d\n", opts.Device, opts.ID)

	return nil
}

// Status represents flood sensor status.
type Status struct {
	ID     int      `json:"id"`
	Alarm  bool     `json:"alarm"`
	Mute   bool     `json:"mute"`
	Errors []string `json:"errors,omitempty"`
}

func collectFloodSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []Status {
	return shelly.CollectSensorsByPrefix[Status](status, "flood:", ios)
}
