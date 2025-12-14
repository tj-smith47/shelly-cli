// Package create provides the schedule create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	timespecFlag string
	callsFlag    string
	enableFlag   bool
)

// NewCommand creates the schedule create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&timespecFlag, "timespec", "", "Cron-like time specification (required)")
	cmd.Flags().StringVar(&callsFlag, "calls", "", "JSON array of RPC calls to execute (required)")
	cmd.Flags().BoolVar(&enableFlag, "enable", true, "Enable schedule after creation")

	if err := cmd.MarkFlagRequired("timespec"); err != nil {
		panic(fmt.Sprintf("failed to mark timespec flag required: %v", err))
	}
	if err := cmd.MarkFlagRequired("calls"); err != nil {
		panic(fmt.Sprintf("failed to mark calls flag required: %v", err))
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Parse calls JSON
	calls, err := shelly.ParseScheduleCalls(callsFlag)
	if err != nil {
		return err
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Creating schedule...", func(ctx context.Context) error {
		id, createErr := svc.CreateSchedule(ctx, device, enableFlag, timespecFlag, calls)
		if createErr != nil {
			return fmt.Errorf("failed to create schedule: %w", createErr)
		}

		ios.Success("Created schedule %d", id)
		ios.Info("Timespec: %s", timespecFlag)
		if enableFlag {
			ios.Info("Schedule is enabled")
		} else {
			ios.Info("Schedule is disabled")
		}
		return nil
	})
}
