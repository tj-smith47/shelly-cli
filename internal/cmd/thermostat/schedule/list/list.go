// Package list provides the thermostat schedule list command.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds list command options.
type Options struct {
	flags.OutputFlags
	Factory      *cmdutil.Factory
	Device       string
	ThermostatID int
	All          bool
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
	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var thermostatSchedules []shelly.ThermostatSchedule
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		conn := dev.Gen2()

		var result any
		spinnerErr := cmdutil.RunWithSpinner(ctx, ios, "Getting schedules...", func(ctx context.Context) error {
			var callErr error
			result, callErr = conn.Call(ctx, "Schedule.List", nil)
			return callErr
		})
		if spinnerErr != nil {
			return fmt.Errorf("failed to list schedules: %w", spinnerErr)
		}

		jsonBytes, marshalErr := json.Marshal(result)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal result: %w", marshalErr)
		}

		var scheduleResp components.ScheduleListResponse
		if unmarshalErr := json.Unmarshal(jsonBytes, &scheduleResp); unmarshalErr != nil {
			return fmt.Errorf("failed to parse schedules: %w", unmarshalErr)
		}

		thermostatSchedules = shelly.FilterThermostatSchedules(scheduleResp.Jobs, opts.ThermostatID, opts.All)
		return nil
	})
	if err != nil {
		return err
	}

	if opts.Format == "json" {
		jsonBytes, jsonErr := json.MarshalIndent(thermostatSchedules, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("failed to format JSON: %w", jsonErr)
		}
		ios.Println(string(jsonBytes))
		return nil
	}

	term.DisplayThermostatSchedules(ios, thermostatSchedules, opts.Device, opts.All)
	return nil
}
