// Package disable provides the schedule disable subcommand.
package disable

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

// NewCommand creates the schedule disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			opts.ID = id
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling schedule...", func(ctx context.Context) error {
		if err := svc.DisableSchedule(ctx, opts.Device, opts.ID); err != nil {
			return fmt.Errorf("failed to disable schedule: %w", err)
		}
		ios.Success("Schedule %d disabled", opts.ID)
		return nil
	})
}
