// Package create provides the schedule create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Timespec string
	Calls    string
	Enable   bool
}

// NewCommand creates the schedule create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Enable:  true,
	}

	cmd := &cobra.Command{
		Use:     "create <device>",
		Aliases: []string{"new"},
		Short:   "Create a new schedule",
		Long: `Create a new schedule on a Gen2+ Shelly device.

Schedules use cron-like timespec expressions:
  - Format: "ss mm hh DD WW" (seconds, minutes, hours, day of month, weekday)
  - Wildcards: * (any), ranges: 1-5, lists: 1,3,5, steps: 0-59/10
  - Special: @sunrise, @sunset (with optional offset like @sunrise+30)

The calls parameter is a JSON array of RPC calls to execute.`,
		Example: `  # Turn on switch at 8:00 AM every day
  shelly schedule create living-room --timespec "0 0 8 * *" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":true}}]'

  # Turn off at sunset
  shelly schedule create living-room --timespec "@sunset" \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":false}}]'

  # Toggle every 30 minutes
  shelly schedule create living-room --timespec "0 */30 * * *" \
    --calls '[{"method":"Switch.Toggle","params":{"id":0}}]'

  # Run script at sunrise + 30 minutes
  shelly schedule create living-room --timespec "@sunrise+30" \
    --calls '[{"method":"Script.Start","params":{"id":1}}]'`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Timespec, "timespec", "", "Cron-like time specification (required)")
	cmd.Flags().StringVar(&opts.Calls, "calls", "", "JSON array of RPC calls to execute (required)")
	cmd.Flags().BoolVar(&opts.Enable, "enable", true, "Enable schedule after creation")

	utils.Must(cmd.MarkFlagRequired("timespec"))
	utils.Must(cmd.MarkFlagRequired("calls"))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Parse calls JSON
	calls, err := automation.ParseScheduleCalls(opts.Calls)
	if err != nil {
		return err
	}

	err = cmdutil.RunWithSpinner(ctx, ios, "Creating schedule...", func(ctx context.Context) error {
		id, createErr := svc.CreateSchedule(ctx, opts.Device, opts.Enable, opts.Timespec, calls)
		if createErr != nil {
			return fmt.Errorf("failed to create schedule: %w", createErr)
		}

		ios.Success("Created schedule %d", id)
		ios.Info("Timespec: %s", opts.Timespec)
		if opts.Enable {
			ios.Info("Schedule is enabled")
		} else {
			ios.Info("Schedule is disabled")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cached schedule list
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeSchedules)
	return nil
}
