// Package disable provides the schedule disable subcommand.
package disable

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the schedule disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable <device> <id>",
		Aliases: []string{"off", "deactivate"},
		Short:   "Disable a schedule",
		Long:    `Disable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Disable a schedule
  shelly schedule disable living-room 1`,
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

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling schedule...", func(ctx context.Context) error {
		if err := svc.DisableSchedule(ctx, device, id); err != nil {
			return fmt.Errorf("failed to disable schedule: %w", err)
		}
		ios.Success("Schedule %d disabled", id)
		return nil
	})
}
