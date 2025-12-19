// Package create provides the thermostat schedule create command.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds create command options.
type Options struct {
	Factory      *cmdutil.Factory
	Device       string
	ThermostatID int
	Timespec     string
	TargetC      float64
	TargetCSet   bool
	Mode         string
	Enable       bool
	Disable      bool
	Enabled      bool
}

// NewCommand creates the thermostat schedule create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f, Enabled: true}

	cmd := &cobra.Command{
		Use:     "create <device>",
		Aliases: []string{"add", "new"},
		Short:   "Create a thermostat schedule",
		Long: `Create a new schedule for thermostat control.

The schedule will execute at the specified time and set the thermostat
to the configured target temperature, mode, or enabled state.

Timespec format (cron-like):
  "ss mm hh DD WW" - seconds, minutes, hours, day of month, weekday

  Wildcards: * (any), ranges: 1-5, lists: 1,3,5, steps: 0-59/10
  Special: @sunrise, @sunset (with optional +/- offset in minutes)

Examples:
  "0 0 8 * *"     - Every day at 8:00 AM
  "0 0 7 * 1-5"   - Weekdays at 7:00 AM
  "0 30 22 * *"   - Every day at 10:30 PM
  "0 0 6 * 0,6"   - Weekends at 6:00 AM
  "@sunrise"      - At sunrise
  "@sunset-30"    - 30 minutes before sunset`,
		Example: `  # Set temperature to 22°C at 7:00 AM on weekdays
  shelly thermostat schedule create gateway --target 22 --time "0 0 7 * 1-5"

  # Set temperature to 18°C at 10:00 PM every day
  shelly thermostat schedule create gateway --target 18 --time "0 0 22 * *"

  # Switch to heat mode at sunrise
  shelly thermostat schedule create gateway --mode heat --time "@sunrise"

  # Disable thermostat at midnight
  shelly thermostat schedule create gateway --disable --time "0 0 0 * *"

  # Create a disabled schedule (won't run until enabled)
  shelly thermostat schedule create gateway --target 20 --time "0 0 9 * *" --disabled`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.TargetCSet = cmd.Flags().Changed("target")
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ThermostatID, "thermostat-id", 0, "Thermostat component ID")
	cmd.Flags().StringVarP(&opts.Timespec, "time", "t", "", "Schedule timespec (required)")
	cmd.Flags().Float64Var(&opts.TargetC, "target", 0, "Target temperature in Celsius")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "Thermostat mode (heat, cool, auto)")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable the thermostat")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable the thermostat")
	cmd.Flags().BoolVar(&opts.Enabled, "enabled", true, "Whether the schedule itself is enabled")

	if err := cmd.MarkFlagRequired("time"); err != nil {
		panic(fmt.Sprintf("failed to mark flag required: %v", err))
	}
	cmd.MarkFlagsMutuallyExclusive("enable", "disable")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if !opts.TargetCSet && opts.Mode == "" && !opts.Enable && !opts.Disable {
		return fmt.Errorf("at least one of --target, --mode, --enable, or --disable must be specified")
	}

	if err := validate.Mode(opts.Mode, true); err != nil {
		return err
	}

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	schedParams := shelly.ThermostatScheduleParams{
		ThermostatID: opts.ThermostatID,
		Timespec:     opts.Timespec,
		Enabled:      opts.Enabled,
		Mode:         opts.Mode,
	}
	if opts.TargetCSet {
		schedParams.TargetC = &opts.TargetC
	}
	if opts.Enable {
		t := true
		schedParams.EnableState = &t
	}
	if opts.Disable {
		f := false
		schedParams.EnableState = &f
	}

	params := shelly.BuildThermostatScheduleCall(schedParams)

	ios.StartProgress("Creating schedule...")
	result, err := conn.Call(ctx, "Schedule.Create", params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	resp, err := shelly.ParseScheduleCreateResponse(result)
	if err != nil {
		return err
	}

	term.DisplayThermostatScheduleCreate(ios, term.ThermostatScheduleCreateDisplay{
		Device:     opts.Device,
		ScheduleID: resp.ID,
		Timespec:   opts.Timespec,
		TargetC:    schedParams.TargetC,
		Mode:       opts.Mode,
		Enable:     opts.Enable,
		Disable:    opts.Disable,
		Enabled:    opts.Enabled,
	})
	return nil
}
