// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/validate"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CreateOptions holds create command options.
type CreateOptions struct {
	Factory      *cmdutil.Factory
	Device       string
	ThermostatID int
	Timespec     string
	TargetC      float64
	TargetCSet   bool // Tracks if --target was explicitly provided
	Mode         string
	Enable       bool
	Disable      bool
	Enabled      bool
}

func newCreateCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &CreateOptions{Factory: f, Enabled: true}

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
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.TargetCSet = cmd.Flags().Changed("target")
			return runCreate(cmd.Context(), opts)
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

func runCreate(ctx context.Context, opts *CreateOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	if err := validateCreateOptions(opts); err != nil {
		return err
	}

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	params := buildScheduleParams(opts)

	ios.StartProgress("Creating schedule...")
	result, err := conn.Call(ctx, "Schedule.Create", params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	scheduleID, err := parseCreateResponse(result)
	if err != nil {
		return err
	}

	displayCreateSuccess(ios, opts, scheduleID)
	return nil
}

func validateCreateOptions(opts *CreateOptions) error {
	if !opts.TargetCSet && opts.Mode == "" && !opts.Enable && !opts.Disable {
		return fmt.Errorf("at least one of --target, --mode, --enable, or --disable must be specified")
	}

	if err := validate.ValidateMode(opts.Mode, true); err != nil {
		return err
	}

	return nil
}

func buildScheduleParams(opts *CreateOptions) map[string]any {
	config := buildThermostatConfig(opts)

	call := map[string]any{
		"method": "Thermostat.SetConfig",
		"params": map[string]any{
			"id":     opts.ThermostatID,
			"config": config,
		},
	}

	return map[string]any{
		"enable":   opts.Enabled,
		"timespec": opts.Timespec,
		"calls":    []any{call},
	}
}

func buildThermostatConfig(opts *CreateOptions) map[string]any {
	config := make(map[string]any)
	if opts.TargetCSet {
		config["target_C"] = opts.TargetC
	}
	if opts.Mode != "" {
		config["thermostat_mode"] = opts.Mode
	}
	if opts.Enable {
		config["enable"] = true
	}
	if opts.Disable {
		config["enable"] = false
	}
	return config
}

func parseCreateResponse(result any) (int, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal result: %w", err)
	}

	var createResp struct {
		ID  int `json:"id"`
		Rev int `json:"rev"`
	}
	if err := json.Unmarshal(jsonBytes, &createResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return createResp.ID, nil
}

func displayCreateSuccess(ios *iostreams.IOStreams, opts *CreateOptions, scheduleID int) {
	ios.Success("Created schedule %d", scheduleID)
	ios.Printf("  Timespec: %s\n", opts.Timespec)

	if opts.TargetCSet {
		ios.Printf("  Target: %.1f°C\n", opts.TargetC)
	}
	if opts.Mode != "" {
		ios.Printf("  Mode: %s\n", opts.Mode)
	}
	if opts.Enable {
		ios.Printf("  Action: enable thermostat\n")
	}
	if opts.Disable {
		ios.Printf("  Action: disable thermostat\n")
	}

	if !opts.Enabled {
		ios.Info("Schedule is disabled. Enable with: shelly thermostat schedule enable %s --id %d", opts.Device, scheduleID)
	}
}
