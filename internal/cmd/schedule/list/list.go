// Package list provides the schedule list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the schedule list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// Check for flag conflict early
	if err := cmdutil.CheckCacheFlags(); err != nil {
		return err
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	result, err := cmdutil.CachedFetchList(ctx, opts.Factory, opts.Device,
		cache.TypeSchedules, cache.TTLAutomation,
		func(ctx context.Context) ([]automation.ScheduleJob, error) {
			ios.StartProgress("Getting schedules...")
			defer ios.StopProgress()
			return svc.ListSchedules(ctx, opts.Device)
		})
	if err != nil {
		return err
	}

	if len(result.Data) == 0 {
		ios.NoResults("schedules")
		return nil
	}

	return cmdutil.PrintListResult(ios, result.Data, term.DisplayScheduleList)
}
