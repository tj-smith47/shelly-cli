// Package enable provides the schedule enable subcommand.
package enable

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

// NewCommand creates the schedule enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device> <id>",
		Aliases: []string{"on", "activate"},
		Short:   "Enable a schedule",
		Long:    `Enable a schedule on a Gen2+ Shelly device.`,
		Example: `  # Enable a schedule
  shelly schedule enable living-room 1`,
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

	return cmdutil.RunWithSpinner(ctx, ios, "Enabling schedule...", func(ctx context.Context) error {
		if err := svc.EnableSchedule(ctx, opts.Device, opts.ID); err != nil {
			return fmt.Errorf("failed to enable schedule: %w", err)
		}
		ios.Success("Schedule %d enabled", opts.ID)
		return nil
	})
}
