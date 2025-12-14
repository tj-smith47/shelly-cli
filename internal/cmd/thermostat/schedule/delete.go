// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DeleteOptions holds delete command options.
type DeleteOptions struct {
	Factory    *cmdutil.Factory
	Device     string
	ScheduleID int
	All        bool
}

func newDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &DeleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete <device>",
		Aliases: []string{"del", "rm", "remove"},
		Short:   "Delete a thermostat schedule",
		Long: `Delete a schedule from the device.

Use --id to specify the schedule ID to delete.
Use --all to delete all schedules (use with caution).`,
		Example: `  # Delete schedule by ID
  shelly thermostat schedule delete gateway --id 1

  # Delete all schedules
  shelly thermostat schedule delete gateway --all`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runDelete(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ScheduleID, "id", 0, "Schedule ID to delete")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Delete all schedules")

	cmd.MarkFlagsMutuallyExclusive("id", "all")

	return cmd
}

func runDelete(ctx context.Context, opts *DeleteOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate that either --id or --all is specified
	if opts.ScheduleID == 0 && !opts.All {
		return fmt.Errorf("either --id or --all must be specified")
	}

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	if opts.All {
		ios.StartProgress("Deleting all schedules...")
		_, err = conn.Call(ctx, "Schedule.DeleteAll", nil)
		ios.StopProgress()

		if err != nil {
			return fmt.Errorf("failed to delete all schedules: %w", err)
		}

		ios.Success("Deleted all schedules from %s", opts.Device)
		return nil
	}

	// Delete specific schedule
	params := map[string]any{
		"id": opts.ScheduleID,
	}

	ios.StartProgress("Deleting schedule...")
	_, err = conn.Call(ctx, "Schedule.Delete", params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to delete schedule %d: %w", opts.ScheduleID, err)
	}

	ios.Success("Deleted schedule %d from %s", opts.ScheduleID, opts.Device)
	return nil
}
