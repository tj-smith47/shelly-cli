// Package set provides the light set subcommand.
package set

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
	flags.ComponentFlags
	Factory    *cmdutil.Factory
	Device     string
	Brightness int
	Temp       int
	On         bool
}

// NewCommand creates the light set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:    f,
		Brightness: -1,
		Temp:       -1,
	}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"brightness", "br"},
		Short:   "Set light parameters",
		Long: `Set parameters of a light component on the specified device.

You can set brightness, white color temperature (Gen1 white-temp bulbs such as
the Duo), and on/off state. Values not specified are left unchanged.`,
		Example: `  # Set brightness to 50%
  shelly light set kitchen --brightness 50

  # Set white color temperature to 4200K (Gen1 Duo)
  shelly light set master-bath --temp 4200

  # Turn on and set brightness
  shelly light br kitchen -b 75 --on`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Light")
	cmd.Flags().IntVarP(&opts.Brightness, "brightness", "b", -1, "Brightness (0-100)")
	cmd.Flags().IntVarP(&opts.Temp, "temp", "t", -1, "White color temperature in Kelvin (Gen1 Duo: 2700-6500)")
	cmd.Flags().BoolVar(&opts.On, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	var brightnessPtr *int
	if opts.Brightness >= 0 && opts.Brightness <= 100 {
		brightnessPtr = &opts.Brightness
	}

	var tempPtr *int
	if opts.Temp > 0 {
		tempPtr = &opts.Temp
	}

	var onPtr *bool
	if opts.On {
		onPtr = &opts.On
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting light parameters...", func(ctx context.Context) error {
		return svc.LightSet(ctx, opts.Device, opts.ID, brightnessPtr, tempPtr, onPtr)
	})
	if err != nil {
		return fmt.Errorf("failed to set light parameters: %w", err)
	}

	ios.Success("Light %d parameters set", opts.ID)
	return nil
}
