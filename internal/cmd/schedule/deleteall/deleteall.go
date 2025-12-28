// Package deleteall provides the schedule delete-all subcommand.
package deleteall

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the schedule delete-all command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete-all <device>",
		Aliases: []string{"clear"},
		Short:   "Delete all schedules",
		Long:    `Delete all schedules from a Gen2+ Shelly device.`,
		Example: `  # Delete all schedules
  shelly schedule delete-all living-room

  # Delete without confirmation
  shelly schedule delete-all living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	// Confirm unless --yes
	ios.Warning("This will delete ALL schedules from the device.")
	confirmed, err := opts.Factory.ConfirmAction("Delete all schedules?", opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Delete cancelled")
		return nil
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Deleting all schedules...", func(ctx context.Context) error {
		if err := svc.DeleteAllSchedules(ctx, opts.Device); err != nil {
			return fmt.Errorf("failed to delete schedules: %w", err)
		}
		ios.Success("All schedules deleted")
		return nil
	})
}
