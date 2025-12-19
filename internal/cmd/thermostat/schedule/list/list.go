// Package list provides the thermostat schedule list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds list command options.
type Options struct {
	Factory      *cmdutil.Factory
	Device       string
	ThermostatID int
	All          bool
	JSON         bool
}

// NewCommand creates the thermostat schedule list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List thermostat schedules",
		Long: `List all schedules that control the thermostat.

By default, only shows schedules that target the thermostat component.
Use --all to show all device schedules.`,
		Example: `  # List thermostat schedules
  shelly thermostat schedule list gateway

  # List schedules for specific thermostat ID
  shelly thermostat schedule list gateway --thermostat-id 1

  # List all device schedules
  shelly thermostat schedule list gateway --all

  # Output as JSON
  shelly thermostat schedule list gateway --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ThermostatID, "thermostat-id", 0, "Filter by thermostat component ID")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Show all device schedules")
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
	defer iostreams.CloseWithDebug("closing connection", conn)

	ios.StartProgress("Getting schedules...")
	result, err := conn.Call(ctx, "Schedule.List", nil)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to list schedules: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var scheduleResp components.ScheduleListResponse
	if err := json.Unmarshal(jsonBytes, &scheduleResp); err != nil {
		return fmt.Errorf("failed to parse schedules: %w", err)
	}

	thermostatSchedules := shelly.FilterThermostatSchedules(scheduleResp.Jobs, opts.ThermostatID, opts.All)

	if opts.JSON {
		jsonBytes, err := json.MarshalIndent(thermostatSchedules, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonBytes))
		return nil
	}

	term.DisplayThermostatSchedules(ios, thermostatSchedules, opts.Device, opts.All)
	return nil
}
