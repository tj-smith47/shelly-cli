// Package enable provides the thermostat schedule enable command.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Options holds enable command options.
type Options struct {
	Factory    *cmdutil.Factory
	Device     string
	ScheduleID int
}

// NewCommand creates the thermostat schedule enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on"},
		Short:   "Enable a schedule",
		Long:    `Enable a disabled schedule so it will run at its scheduled times.`,
		Example: `  # Enable schedule by ID
  shelly thermostat schedule enable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ScheduleID, "id", 0, "Schedule ID to enable (required)")
	utils.Must(cmd.MarkFlagRequired("id"))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		conn := dev.Gen2()

		params := map[string]any{
			"id":     opts.ScheduleID,
			"enable": true,
		}

		err := cmdutil.RunWithSpinner(ctx, ios, "Enabling schedule...", func(ctx context.Context) error {
			_, callErr := conn.Call(ctx, "Schedule.Update", params)
			return callErr
		})
		if err != nil {
			return fmt.Errorf("failed to enable schedule %d: %w", opts.ScheduleID, err)
		}

		ios.Success("Enabled schedule %d", opts.ScheduleID)
		return nil
	})
}
