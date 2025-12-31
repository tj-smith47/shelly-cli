// Package update provides the schedule update subcommand.
package update

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Timespec string
	Calls    string
	Enable   bool
	Disable  bool
}

// NewCommand creates the schedule update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "update <device> <id>",
		Aliases: []string{"up"},
		Short:   "Update a schedule",
		Long: `Update an existing schedule on a Gen2+ Shelly device.

You can update the timespec, calls, or enabled status.`,
		Example: `  # Update timespec
  shelly schedule update living-room 1 --timespec "0 0 9 * *"

  # Update calls
  shelly schedule update living-room 1 \
    --calls '[{"method":"Switch.Set","params":{"id":0,"on":false}}]'

  # Enable/disable
  shelly schedule update living-room 1 --enable
  shelly schedule update living-room 1 --disable`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenScheduleID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid schedule ID: %s", args[1])
			}
			opts.Device = args[0]
			opts.ID = id
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Timespec, "timespec", "", "Cron-like time specification")
	cmd.Flags().StringVar(&opts.Calls, "calls", "", "JSON array of RPC calls to execute")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable the schedule")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable the schedule")

	cmd.MarkFlagsMutuallyExclusive("enable", "disable")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Parse calls if provided
	var calls []automation.ScheduleCall
	if opts.Calls != "" {
		var err error
		calls, err = automation.ParseScheduleCalls(opts.Calls)
		if err != nil {
			return err
		}
	}

	// Build update params
	var enablePtr *bool
	if opts.Enable {
		enable := true
		enablePtr = &enable
	} else if opts.Disable {
		enable := false
		enablePtr = &enable
	}

	var timespecPtr *string
	if opts.Timespec != "" {
		timespecPtr = &opts.Timespec
	}

	// Check if anything was specified
	if enablePtr == nil && timespecPtr == nil && calls == nil {
		ios.Warning("No changes specified")
		return nil
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Updating schedule...", func(ctx context.Context) error {
		if updateErr := svc.UpdateSchedule(ctx, opts.Device, opts.ID, enablePtr, timespecPtr, calls); updateErr != nil {
			return fmt.Errorf("failed to update schedule: %w", updateErr)
		}

		ios.Success("Schedule %d updated", opts.ID)
		if timespecPtr != nil {
			ios.Info("Timespec: %s", *timespecPtr)
		}
		if enablePtr != nil {
			if *enablePtr {
				ios.Info("Schedule enabled")
			} else {
				ios.Info("Schedule disabled")
			}
		}
		if calls != nil {
			ios.Info("Calls updated")
		}
		return nil
	})
}
