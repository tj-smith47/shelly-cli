// Package create provides the schedule create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	timespecFlag string
	callsFlag    string
	enableFlag   bool
)

// NewCommand creates the schedule create command.
func NewCommand() *cobra.Command {
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	cmd.Flags().StringVar(&timespecFlag, "timespec", "", "Cron-like time specification (required)")
	cmd.Flags().StringVar(&callsFlag, "calls", "", "JSON array of RPC calls to execute (required)")
	cmd.Flags().BoolVar(&enableFlag, "enable", true, "Enable schedule after creation")

	//nolint:errcheck // cobra's MarkFlagRequired only returns errors for non-existent flags
	cmd.MarkFlagRequired("timespec")
	//nolint:errcheck // cobra's MarkFlagRequired only returns errors for non-existent flags
	cmd.MarkFlagRequired("calls")

	return cmd
}

func run(ctx context.Context, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

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
