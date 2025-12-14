// Package enable provides the schedule enable subcommand.
package enable

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the schedule enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <device> <id>",
		Short: "Enable a schedule",
		Long:  `Enable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Enable a schedule
  shelly schedule enable living-room 1`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: cmdutil.CompleteDeviceThenScheduleID(),
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

	return cmdutil.RunWithSpinner(ctx, ios, "Enabling schedule...", func(ctx context.Context) error {
		if err := svc.EnableSchedule(ctx, device, id); err != nil {
			return fmt.Errorf("failed to enable schedule: %w", err)
		}
		ios.Success("Schedule %d enabled", id)
		return nil
	})
}
