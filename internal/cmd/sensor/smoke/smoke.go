// Package smoke provides smoke sensor commands.
package smoke

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/sensorutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the smoke command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "smoke",
		Aliases: []string{"detector"},
		Short:   "Manage smoke sensors",
		Long: `Manage smoke detection sensors on Shelly devices.

Smoke sensors provide alarm state detection and the ability
to mute active alarms.`,
		Example: `  # List smoke sensors
  shelly sensor smoke list kitchen

  # Check smoke status
  shelly sensor smoke status kitchen

  # Test smoke alarm
  shelly sensor smoke test kitchen

  # Mute active alarm
  shelly sensor smoke mute kitchen`,
	}

	cmd.AddCommand(newListCommand(f))
	cmd.AddCommand(newStatusCommand(f))
	cmd.AddCommand(newTestCommand(f))
	cmd.AddCommand(newMuteCommand(f))

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
		Short:   "List smoke sensors",
		Long:    `List all smoke sensors on a Shelly device.`,
		Example: `  # List smoke sensors
  shelly sensor smoke list kitchen

  # Output as JSON
  shelly sensor smoke list kitchen --json`,
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

	sensors := collectSmokeSensors(fullStatus, ios)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No smoke sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Smoke Sensors:"))
	ios.Println()
	for _, s := range sensors {
		status := theme.StatusOK().Render("Clear")
		if s.Alarm {
			status = theme.StatusError().Render("SMOKE DETECTED!")
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
		Short:   "Check smoke status",
		Long:    `Get the current status of a smoke sensor.`,
		Example: `  # Check smoke status
  shelly sensor smoke status kitchen

  # Check specific sensor
  shelly sensor smoke status kitchen --id 1

  # Output as JSON
  shelly sensor smoke status kitchen --json`,
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
	result, err := conn.Call(ctx, "Smoke.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get smoke status: %w", err)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Smoke Sensor %d:", opts.ID)))
	ios.Println()

	if status.Alarm {
		ios.Printf("  Status: %s\n", theme.StatusError().Render("SMOKE DETECTED!"))
	} else {
		ios.Printf("  Status: %s\n", theme.StatusOK().Render("Clear"))
	}

	if status.Mute {
		ios.Printf("  Alarm: %s\n", theme.Dim().Render("Muted"))
	} else {
		ios.Printf("  Alarm: %s\n", theme.Highlight().Render("Active"))
	}

	if len(status.Errors) > 0 {
		ios.Warning("Errors: %s", strings.Join(status.Errors, ", "))
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
		Short:   "Test smoke alarm",
		Long: `Test the smoke alarm on a Shelly device.

Note: The Smoke component may not have a dedicated test method.
This command provides instructions for manual testing.`,
		Example: `  # Test smoke alarm
  shelly sensor smoke test kitchen`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
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

	// The Smoke component doesn't have a Test method in the API
	// Provide instructions for manual testing
	ios.Info("Smoke sensor testing:")
	ios.Println()
	ios.Println("  The Smoke component does not have a programmatic test method.")
	ios.Println("  To test the smoke detector, use the device's physical test")
	ios.Println("  button or use appropriate test spray (follow manufacturer")
	ios.Println("  guidelines).")
	ios.Println()
	ios.Println("  Monitor status with:")
	ios.Printf("    shelly sensor smoke status %s --id %d\n", opts.Device, opts.ID)

	return nil
}

// MuteOptions holds mute command options.
type MuteOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newMuteCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &MuteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "mute <device>",
		Aliases: []string{"silence", "quiet"},
		Short:   "Mute smoke alarm",
		Long: `Mute an active smoke alarm.

The alarm will remain muted until the smoke condition clears
and potentially re-triggers.`,
		Example: `  # Mute smoke alarm
  shelly sensor smoke mute kitchen

  # Mute specific sensor
  shelly sensor smoke mute kitchen --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runMute(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 0, "Sensor ID (default 0)")

	return cmd
}

func runMute(ctx context.Context, opts *MuteOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	params := map[string]any{"id": opts.ID}
	_, err = conn.Call(ctx, "Smoke.Mute", params)
	if err != nil {
		return fmt.Errorf("failed to mute smoke alarm: %w", err)
	}

	ios.Success("Smoke alarm muted for sensor %d.", opts.ID)

	return nil
}

// Status represents smoke sensor status.
type Status struct {
	ID     int      `json:"id"`
	Alarm  bool     `json:"alarm"`
	Mute   bool     `json:"mute"`
	Errors []string `json:"errors,omitempty"`
}

func collectSmokeSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []Status {
	return sensorutil.CollectByPrefix[Status](status, "smoke:", ios)
}
