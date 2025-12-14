// Package del provides the schedule delete subcommand.
package del

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var yesFlag bool

// NewCommand creates the schedule delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <device> <id>",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a schedule",
		Long:    `Delete a schedule from a Gen2+ Shelly device.`,
		Example: `  # Delete a schedule
  shelly schedule delete living-room 1

  # Delete without confirmation
  shelly schedule delete living-room 1 --yes`,
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

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Confirm unless --yes
	if !yesFlag {
		ios.Warning("This will delete schedule %d.", id)
		confirmed, err := ios.Confirm("Delete schedule?", false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Warning("Delete cancelled")
			return nil
		}
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Deleting schedule...", func(ctx context.Context) error {
		if err := svc.DeleteSchedule(ctx, device, id); err != nil {
			return fmt.Errorf("failed to delete schedule: %w", err)
		}
		ios.Success("Schedule %d deleted", id)
		return nil
	})
}
