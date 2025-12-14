// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DisableOptions holds disable command options.
type DisableOptions struct {
	Factory    *cmdutil.Factory
	Device     string
	ScheduleID int
}

func newDisableCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &DisableOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off"},
		Short:   "Disable a schedule",
		Long:    `Disable a schedule so it will not run until re-enabled.`,
		Example: `  # Disable schedule by ID
  shelly thermostat schedule disable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runDisable(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ScheduleID, "id", 0, "Schedule ID to disable (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		cmd.Printf("Warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runDisable(ctx context.Context, opts *DisableOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	params := map[string]any{
		"id":     opts.ScheduleID,
		"enable": false,
	}

	ios.StartProgress("Disabling schedule...")
	_, err = conn.Call(ctx, "Schedule.Update", params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to disable schedule %d: %w", opts.ScheduleID, err)
	}

	ios.Success("Disabled schedule %d", opts.ScheduleID)
	return nil
}
