// Package list provides the schedule list subcommand.
package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the schedule list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls"},
		Short:   "List schedules on a device",
		Long: `List all schedules on a Gen2+ Shelly device.

Shows schedule ID, enabled status, timespec (cron-like syntax), and the
RPC calls to execute. Schedules allow time-based automation of device
actions without external home automation systems.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Enabled, Timespec, Calls`,
		Example: `  # List all schedules
  shelly schedule list living-room

  # Output as JSON
  shelly schedule list living-room -o json

  # Get enabled schedules only
  shelly schedule list living-room -o json | jq '.[] | select(.enable)'

  # Extract timespecs for enabled schedules
  shelly schedule list living-room -o json | jq -r '.[] | select(.enable) | .timespec'

  # Count total schedules
  shelly schedule list living-room -o json | jq length

  # Short form
  shelly sched ls living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting schedules...",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.ScheduleJob, error) {
			return svc.ListSchedules(ctx, device)
		},
		displaySchedules)
}

func displaySchedules(ios *iostreams.IOStreams, schedules []shelly.ScheduleJob) {
	if len(schedules) == 0 {
		ios.Info("No schedules found on this device")
		return
	}

	table := output.NewTable("ID", "Enabled", "Timespec", "Calls")
	for _, s := range schedules {
		enabled := theme.Dim().Render("no")
		if s.Enable {
			enabled = theme.StatusOK().Render("yes")
		}

		// Format calls summary
		callsSummary := formatCallsSummary(s.Calls)

		table.AddRow(fmt.Sprintf("%d", s.ID), enabled, s.Timespec, callsSummary)
	}
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

func formatCallsSummary(calls []shelly.ScheduleCall) string {
	if len(calls) == 0 {
		return theme.Dim().Render("(none)")
	}

	if len(calls) == 1 {
		call := calls[0]
		if len(call.Params) == 0 {
			return call.Method
		}
		params, err := json.Marshal(call.Params)
		if err != nil {
			return call.Method
		}
		return fmt.Sprintf("%s %s", call.Method, string(params))
	}

	return fmt.Sprintf("%d calls", len(calls))
}
