// Package disable provides the schedule disable subcommand.
package disable

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the schedule disable command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <device> <id>",
		Short: "Disable a schedule",
		Long:  `Disable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Disable a schedule
  shelly schedule disable living-room 1`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid schedule ID: %s", args[1])
			}
			return run(cmd.Context(), args[0], id)
		},
	}

	return cmd
}

func run(ctx context.Context, device string, id int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := iostreams.System()
	svc := shelly.NewService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling schedule...", func(ctx context.Context) error {
		if err := svc.DisableSchedule(ctx, device, id); err != nil {
			return fmt.Errorf("failed to disable schedule: %w", err)
		}
		ios.Success("Schedule %d disabled", id)
		return nil
	})
}
