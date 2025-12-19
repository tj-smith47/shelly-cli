// Package disable provides the thermostat schedule disable command.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds disable command options.
type Options struct {
	Factory    *cmdutil.Factory
	Device     string
	ScheduleID int
}

// NewCommand creates the thermostat schedule disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off"},
		Short:   "Disable a schedule",
		Long:    `Disable a schedule so it will not run until re-enabled.`,
		Example: `  # Disable schedule by ID
  shelly thermostat schedule disable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ScheduleID, "id", 0, "Schedule ID to disable (required)")
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(fmt.Sprintf("failed to mark flag required: %v", err))
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
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
