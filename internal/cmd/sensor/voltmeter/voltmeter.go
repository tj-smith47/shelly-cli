// Package voltmeter provides voltmeter sensor commands.
package voltmeter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the voltmeter command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "voltmeter",
		Aliases: []string{"volt", "voltage", "v"},
		Short:   "Manage voltmeter sensors",
		Long: `Manage voltmeter sensors on Shelly devices.

Voltmeter sensors provide voltage readings, useful for monitoring
power supplies, batteries, or other voltage sources.`,
		Example: `  # List voltmeters
  shelly sensor voltmeter list device1

  # Get voltage reading
  shelly sensor voltmeter status device1`,
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
		Short:   "List voltmeters",
		Long:    `List all voltmeter sensors on a Shelly device.`,
		Example: `  # List voltmeters
  shelly sensor voltmeter list device1

  # Output as JSON
  shelly sensor voltmeter list device1 --json`,
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

	sensors := collectVoltmeterSensors(fullStatus)

	if opts.JSON {
		output, err := json.MarshalIndent(sensors, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	if len(sensors) == 0 {
		ios.Info("No voltmeter sensors found on this device.")
		return nil
	}

	ios.Println(theme.Bold().Render("Voltmeter Sensors:"))
	ios.Println()
	for _, s := range sensors {
		ios.Printf("  Sensor %d:\n", s.ID)
		if s.Voltage != nil {
			ios.Printf("    Voltage: %.2f V\n", *s.Voltage)
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
		Short:   "Get voltage reading",
		Long:    `Get the current voltage reading from a voltmeter sensor.`,
		Example: `  # Get voltage from default sensor (ID 0)
  shelly sensor voltmeter status device1

  # Get from specific sensor
  shelly sensor voltmeter status device1 --id 1

  # Output as JSON
  shelly sensor voltmeter status device1 --json`,
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
	result, err := conn.Call(ctx, "Voltmeter.GetStatus", params)
	if err != nil {
		return fmt.Errorf("failed to get voltmeter status: %w", err)
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

	ios.Println(theme.Bold().Render(fmt.Sprintf("Voltmeter Sensor %d:", opts.ID)))
	ios.Println()
	if status.Voltage != nil {
		ios.Printf("  Voltage: %s\n", theme.Highlight().Render(fmt.Sprintf("%.3f V", *status.Voltage)))
	} else {
		ios.Warning("No voltage reading available.")
	}

	if len(status.Errors) > 0 {
		ios.Println()
		ios.Warning("Errors: %v", status.Errors)
	}

	return nil
}

// Status represents voltmeter sensor status.
type Status struct {
	ID      int      `json:"id"`
	Voltage *float64 `json:"voltage"`
	Errors  []string `json:"errors,omitempty"`
}

func collectVoltmeterSensors(status map[string]json.RawMessage) []Status {
	var sensors []Status
	for key, raw := range status {
		if strings.HasPrefix(key, "voltmeter:") {
			var s Status
			if err := json.Unmarshal(raw, &s); err == nil {
				sensors = append(sensors, s)
			}
		}
	}
	return sensors
}
