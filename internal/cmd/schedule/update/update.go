// Package update provides the schedule update subcommand.
package update

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	timespecFlag string
	callsFlag    string
	enableFlag   bool
	disableFlag  bool
)

// NewCommand creates the schedule update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0], id)
		},
	}

	cmd.Flags().StringVar(&timespecFlag, "timespec", "", "Cron-like time specification")
	cmd.Flags().StringVar(&callsFlag, "calls", "", "JSON array of RPC calls to execute")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable the schedule")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Disable the schedule")

	cmd.MarkFlagsMutuallyExclusive("enable", "disable")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Parse calls if provided
	var calls []shelly.ScheduleCall
	if callsFlag != "" {
		var err error
		calls, err = shelly.ParseScheduleCalls(callsFlag)
		if err != nil {
			return err
		}
	}

	// Build update params
	var enablePtr *bool
	if enableFlag {
		enable := true
		enablePtr = &enable
	} else if disableFlag {
		enable := false
		enablePtr = &enable
	}

	var timespecPtr *string
	if timespecFlag != "" {
		timespecPtr = &timespecFlag
	}

	// Check if anything was specified
	if enablePtr == nil && timespecPtr == nil && calls == nil {
		ios.Warning("No changes specified")
		return nil
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Updating schedule...", func(ctx context.Context) error {
		if updateErr := svc.UpdateSchedule(ctx, device, id, enablePtr, timespecPtr, calls); updateErr != nil {
			return fmt.Errorf("failed to update schedule: %w", updateErr)
		}

		ios.Success("Schedule %d updated", id)
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
