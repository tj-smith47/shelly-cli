// Package enable provides the thermostat enable command.
package enable

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
	Mode    string
}

// NewCommand creates the thermostat enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on", "start"},
		Short:   "Enable thermostat",
		Long: `Enable a thermostat component.

Optionally set the operating mode when enabling:
- heat: Heating mode (opens valve when below target)
- cool: Cooling mode (opens valve when above target)
- auto: Automatic mode (device decides based on conditions)`,
		Example: `  # Enable thermostat with current settings
  shelly thermostat enable gateway

  # Enable in heat mode
  shelly thermostat enable gateway --mode heat

  # Enable specific thermostat in auto mode
  shelly thermostat enable gateway --id 1 --mode auto`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Thermostat")
	cmd.Flags().StringVar(&opts.Mode, "mode", "", "Operating mode (heat, cool, auto)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate mode if provided
	if err := shelly.ValidateThermostatMode(opts.Mode, true); err != nil {
		return err
	}

	return svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		if dev.IsGen1() {
			return fmt.Errorf("thermostat component requires Gen2+ device")
		}

		thermostat := dev.Gen2().Thermostat(opts.ID)

		err := cmdutil.RunWithSpinner(ctx, ios, "Enabling thermostat...", func(ctx context.Context) error {
			if enableErr := thermostat.Enable(ctx, true); enableErr != nil {
				return fmt.Errorf("failed to enable thermostat: %w", enableErr)
			}
			if opts.Mode != "" {
				if modeErr := thermostat.SetMode(ctx, opts.Mode); modeErr != nil {
					return fmt.Errorf("failed to set thermostat mode: %w", modeErr)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		if opts.Mode != "" {
			ios.Success("Thermostat %d enabled in %s mode", opts.ID, opts.Mode)
		} else {
			ios.Success("Thermostat %d enabled", opts.ID)
		}

		return nil
	})
}
