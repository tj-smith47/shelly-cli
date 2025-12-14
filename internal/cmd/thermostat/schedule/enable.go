// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// EnableOptions holds enable command options.
type EnableOptions struct {
	Factory    *cmdutil.Factory
	Device     string
	ScheduleID int
}

func newEnableCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &EnableOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on"},
		Short:   "Enable a schedule",
		Long:    `Enable a disabled schedule so it will run at its scheduled times.`,
		Example: `  # Enable schedule by ID
  shelly thermostat schedule enable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runEnable(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ScheduleID, "id", 0, "Schedule ID to enable (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		cmd.Printf("Warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runEnable(ctx context.Context, opts *EnableOptions) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	conn, err := svc.Connect(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing connection", conn)

	params := map[string]any{
		"id":     opts.ScheduleID,
		"enable": true,
	}

	ios.StartProgress("Enabling schedule...")
	_, err = conn.Call(ctx, "Schedule.Update", params)
	ios.StopProgress()

	if err != nil {
		return fmt.Errorf("failed to enable schedule %d: %w", opts.ScheduleID, err)
	}

	ios.Success("Enabled schedule %d", opts.ScheduleID)
	return nil
}
