// Package remove provides the sensoraddon remove command.
package remove

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
	Device    string
	Component string
	Factory   *cmdutil.Factory
}

// NewCommand creates the sensoraddon remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <device> <component>",
		Aliases: []string{"rm", "del", "delete"},
		Short:   "Remove a peripheral",
		Long: `Remove a Sensor Add-on peripheral from a device.

The component key format is "type:id", for example "temperature:100" or "input:101".

Note: Changes require a device reboot to take effect.`,
		Example: `  # Remove a peripheral
  shelly sensoraddon remove kitchen temperature:100

  # Skip confirmation
  shelly sensoraddon remove kitchen input:101 --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Component = args[1]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddConfirmFlags(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.SensorAddonService()

	// Confirm removal
	if !opts.Yes {
		confirmed, err := ios.Confirm(fmt.Sprintf("Remove peripheral %s?", opts.Component), false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Removing peripheral...", func(ctx context.Context) error {
		return svc.RemovePeripheral(ctx, opts.Device, opts.Component)
	})
	if err != nil {
		return err
	}

	ios.Success("Removed peripheral %s", opts.Component)
	ios.Warning("Device reboot required for changes to take effect")
	return nil
}
