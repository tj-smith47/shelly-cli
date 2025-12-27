// Package disable provides the thermostat disable command.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	flags.ComponentFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the thermostat disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "stop"},
		Short:   "Disable thermostat",
		Long: `Disable a thermostat component.

When disabled, the thermostat will not actively control the valve
position based on temperature. The valve will typically remain
in its current position or close.`,
		Example: `  # Disable thermostat
  shelly thermostat disable gateway

  # Disable specific thermostat
  shelly thermostat disable gateway --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Thermostat")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		thermostat := dev.Gen2().Thermostat(opts.ID)

		err := cmdutil.RunWithSpinner(ctx, ios, "Disabling thermostat...", func(ctx context.Context) error {
			return thermostat.Enable(ctx, false)
		})
		if err != nil {
			return fmt.Errorf("failed to disable thermostat: %w", err)
		}

		ios.Success("Thermostat %d disabled", opts.ID)
		return nil
	})
}
