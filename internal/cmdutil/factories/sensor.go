// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// SensorListDisplay is a function that displays a sensor list in human-readable format.
type SensorListDisplay[T any] func(ios *iostreams.IOStreams, sensors []T)

// SensorStatusDisplay is a function that displays sensor status in human-readable format.
type SensorStatusDisplay[T any] func(ios *iostreams.IOStreams, status T, id int)

// SensorOpts configures a sensor command group.
type SensorOpts[T any] struct {
	// Name is the sensor name used in commands and messages (e.g., "temperature", "smoke").
	Name string

	// Aliases are command aliases for the parent command.
	Aliases []string

	// Short is the short description for the parent command.
	Short string

	// Long is the long description for the parent command.
	Long string

	// Example is the example text for the parent command.
	Example string

	// Prefix is the API status key prefix (e.g., "temperature:", "smoke:").
	Prefix string

	// StatusMethod is the RPC method name (e.g., "Temperature.GetStatus").
	StatusMethod string

	// DisplayList formats the sensor list for human-readable output.
	// Not needed for alarm sensors when AlarmSensorTitle is set.
	DisplayList SensorListDisplay[T]

	// DisplayStatus formats single sensor status for human-readable output.
	// Not needed for alarm sensors when AlarmSensorTitle is set.
	DisplayStatus SensorStatusDisplay[T]

	// AlarmSensorTitle is the display title for alarm-type sensors (e.g., "Smoke", "Flood").
	// When set, the factory uses generic alarm display functions instead of DisplayList/DisplayStatus.
	// Requires T to implement model.AlarmSensor interface.
	AlarmSensorTitle string

	// AlarmMessage is the alarm message for alarm-type sensors (e.g., "SMOKE DETECTED!").
	// Only used when AlarmSensorTitle is set.
	AlarmMessage string

	// HasTest indicates if a test subcommand should be added.
	HasTest bool

	// TestHint is the instruction text for the test command.
	// Only used if HasTest is true.
	TestHint string

	// HasMute indicates if a mute subcommand should be added.
	HasMute bool

	// MuteMethod is the RPC method for mute (e.g., "Smoke.Mute").
	// Only used if HasMute is true.
	MuteMethod string
}

// NewSensorCommand creates a sensor command group with list, status, and optional test/mute subcommands.
func NewSensorCommand[T any](f *cmdutil.Factory, opts SensorOpts[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     opts.Name,
		Aliases: opts.Aliases,
		Short:   opts.Short,
		Long:    opts.Long,
		Example: opts.Example,
	}

	cmd.AddCommand(newSensorListCommand(f, opts))
	cmd.AddCommand(newSensorStatusCommand(f, opts))

	if opts.HasTest {
		cmd.AddCommand(newSensorTestCommand(f, opts))
	}

	if opts.HasMute {
		cmd.AddCommand(newSensorMuteCommand(f, opts))
	}

	return cmd
}

func newSensorListCommand[T any](f *cmdutil.Factory, opts SensorOpts[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   fmt.Sprintf("List %s sensors", opts.Name),
		Long:    fmt.Sprintf("List all %s sensors on a Shelly device.", opts.Name),
		Example: fmt.Sprintf(`  # List %s sensors
  shelly sensor %s list <device>

  # Output as JSON
  shelly sensor %s list <device> -o json`, opts.Name, opts.Name, opts.Name),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSensorList(cmd.Context(), f, opts, args[0])
		},
	}

	return cmd
}

func runSensorList[T any](ctx context.Context, f *cmdutil.Factory, opts SensorOpts[T], device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	spinnerMsg := fmt.Sprintf("Fetching %s sensors...", opts.Name)
	emptyMsg := fmt.Sprintf("No %s sensors found on this device.", opts.Name)

	fetcher := func(ctx context.Context, svc *shelly.Service, device string) ([]T, error) {
		return fetchSensorList[T](ctx, svc, device, opts.Prefix, ios)
	}

	return cmdutil.RunList(ctx, ios, svc, device, spinnerMsg, emptyMsg, fetcher, buildSensorListDisplay(opts))
}

// buildSensorListDisplay creates a list display function for the sensor factory.
// For alarm sensors (AlarmSensorTitle set), uses the generic alarm display.
// For other sensors, uses the provided DisplayList function.
func buildSensorListDisplay[T any](opts SensorOpts[T]) cmdutil.ListDisplay[T] {
	if opts.AlarmSensorTitle != "" {
		return func(ios *iostreams.IOStreams, sensors []T) {
			if alarmSensors, ok := any(sensors).([]model.AlarmSensorReading); ok {
				term.DisplayAlarmSensorList(ios, alarmSensors, opts.AlarmSensorTitle, opts.AlarmMessage)
			}
		}
	}
	return cmdutil.ListDisplay[T](opts.DisplayList)
}

func fetchSensorList[T any](ctx context.Context, svc *shelly.Service, device, prefix string, ios *iostreams.IOStreams) ([]T, error) {
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}

	result, err := conn.Call(ctx, "Shelly.GetStatus", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var fullStatus map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &fullStatus); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return shelly.CollectSensorsByPrefix[T](fullStatus, prefix, ios), nil
}

func newSensorStatusCommand[T any](f *cmdutil.Factory, opts SensorOpts[T]) *cobra.Command {
	var sensorID int

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "get"},
		Short:   fmt.Sprintf("Get %s sensor status", opts.Name),
		Long:    fmt.Sprintf("Get the current status of a %s sensor.", opts.Name),
		Example: fmt.Sprintf(`  # Get %s status
  shelly sensor %s status <device>

  # Get specific sensor
  shelly sensor %s status <device> --id 1

  # Output as JSON
  shelly sensor %s status <device> -o json`, opts.Name, opts.Name, opts.Name, opts.Name),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSensorStatus(cmd.Context(), f, opts, args[0], sensorID)
		},
	}

	cmd.Flags().IntVar(&sensorID, "id", 0, "Sensor ID (default 0)")

	return cmd
}

func runSensorStatus[T any](ctx context.Context, f *cmdutil.Factory, opts SensorOpts[T], device string, sensorID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	spinnerMsg := fmt.Sprintf("Fetching %s status...", opts.Name)

	fetcher := func(ctx context.Context, svc *shelly.Service, device string, id int) (T, error) {
		return fetchSensorStatus[T](ctx, svc, device, opts.StatusMethod, id)
	}

	return cmdutil.RunStatus(ctx, ios, svc, device, sensorID, spinnerMsg, fetcher, buildSensorStatusDisplay(opts, sensorID))
}

// buildSensorStatusDisplay creates a status display function for the sensor factory.
// For alarm sensors (AlarmSensorTitle set), uses the generic alarm display.
// For other sensors, uses the provided DisplayStatus function.
func buildSensorStatusDisplay[T any](opts SensorOpts[T], sensorID int) cmdutil.StatusDisplay[T] {
	if opts.AlarmSensorTitle != "" {
		return func(ios *iostreams.IOStreams, status T) {
			if alarmStatus, ok := any(status).(model.AlarmSensorReading); ok {
				term.DisplayAlarmSensorStatus(ios, alarmStatus, sensorID, opts.AlarmSensorTitle, opts.AlarmMessage)
			}
		}
	}
	return func(ios *iostreams.IOStreams, status T) {
		opts.DisplayStatus(ios, status, sensorID)
	}
}

func fetchSensorStatus[T any](ctx context.Context, svc *shelly.Service, device, method string, id int) (T, error) {
	var zero T

	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return zero, fmt.Errorf("failed to connect to device: %w", err)
	}

	params := map[string]any{"id": id}
	result, err := conn.Call(ctx, method, params)
	if err != nil {
		return zero, fmt.Errorf("failed to get sensor status: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return zero, fmt.Errorf("failed to marshal result: %w", err)
	}

	var status T
	if err := json.Unmarshal(jsonBytes, &status); err != nil {
		return zero, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

func newSensorTestCommand[T any](f *cmdutil.Factory, opts SensorOpts[T]) *cobra.Command {
	var sensorID int

	titleName := strings.ToTitle(opts.Name[:1]) + opts.Name[1:]

	cmd := &cobra.Command{
		Use:     "test <device>",
		Aliases: []string{"trigger"},
		Short:   fmt.Sprintf("Test %s sensor", opts.Name),
		Long: fmt.Sprintf(`Test the %s sensor on a Shelly device.

Note: The %s component may not have a dedicated test method.
This command provides instructions for manual testing.`, opts.Name, titleName),
		Example: fmt.Sprintf(`  # Test %s sensor
  shelly sensor %s test <device>`, opts.Name, opts.Name),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSensorTest(f, opts, args[0], sensorID)
		},
	}

	cmd.Flags().IntVar(&sensorID, "id", 0, "Sensor ID (default 0)")

	return cmd
}

func runSensorTest[T any](f *cmdutil.Factory, opts SensorOpts[T], device string, sensorID int) error {
	ios := f.IOStreams()

	titleName := strings.ToTitle(opts.Name[:1]) + opts.Name[1:]

	ios.Info("%s sensor testing:", titleName)
	ios.Println()
	ios.Printf("  %s\n", opts.TestHint)
	ios.Println()
	ios.Println("  Monitor status with:")
	ios.Printf("    shelly sensor %s status %s --id %d\n", opts.Name, device, sensorID)

	return nil
}

func newSensorMuteCommand[T any](f *cmdutil.Factory, opts SensorOpts[T]) *cobra.Command {
	var sensorID int

	cmd := &cobra.Command{
		Use:     "mute <device>",
		Aliases: []string{"silence", "quiet"},
		Short:   fmt.Sprintf("Mute %s alarm", opts.Name),
		Long: fmt.Sprintf(`Mute an active %s alarm.

The alarm will remain muted until the condition clears
and potentially re-triggers.`, opts.Name),
		Example: fmt.Sprintf(`  # Mute %s alarm
  shelly sensor %s mute <device>

  # Mute specific sensor
  shelly sensor %s mute <device> --id 1`, opts.Name, opts.Name, opts.Name),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSensorMute(cmd.Context(), f, opts, args[0], sensorID)
		},
	}

	cmd.Flags().IntVar(&sensorID, "id", 0, "Sensor ID (default 0)")

	return cmd
}

func runSensorMute[T any](ctx context.Context, f *cmdutil.Factory, opts SensorOpts[T], device string, sensorID int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	titleName := strings.ToTitle(opts.Name[:1]) + opts.Name[1:]

	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	params := map[string]any{"id": sensorID}
	_, err = conn.Call(ctx, opts.MuteMethod, params)
	if err != nil {
		return fmt.Errorf("failed to mute %s alarm: %w", opts.Name, err)
	}

	ios.Success("%s alarm muted for sensor %d.", titleName, sensorID)

	return nil
}
