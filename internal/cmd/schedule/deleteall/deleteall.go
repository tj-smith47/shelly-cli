// Package deleteall provides the schedule delete-all subcommand.
package deleteall

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var yesFlag bool

// NewCommand creates the schedule delete-all command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Confirm unless --yes
	if !yesFlag {
		ios.Warning("This will delete ALL schedules from the device.")
		confirmed, err := ios.Confirm("Delete all schedules?", false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Warning("Delete cancelled")
			return nil
		}
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Deleting all schedules...", func(ctx context.Context) error {
		if err := svc.DeleteAllSchedules(ctx, device); err != nil {
			return fmt.Errorf("failed to delete schedules: %w", err)
		}
		ios.Success("All schedules deleted")
		return nil
	})
}
