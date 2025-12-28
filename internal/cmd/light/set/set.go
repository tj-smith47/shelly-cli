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
	On         bool
}

// NewCommand creates the light set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:    f,
		Brightness: -1,
	}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"brightness", "br"},
		Short:   "Set light parameters",
		Long: `Set parameters of a light component on the specified device.

You can set brightness and on/off state.
Values not specified will be left unchanged.`,
		Example: `  # Set brightness to 50%
  shelly light set kitchen --brightness 50

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

	var onPtr *bool
	if opts.On {
		onPtr = &opts.On
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting light parameters...", func(ctx context.Context) error {
		return svc.LightSet(ctx, opts.Device, opts.ID, brightnessPtr, onPtr)
	})
	if err != nil {
		return fmt.Errorf("failed to set light parameters: %w", err)
	}

	ios.Success("Light %d parameters set", opts.ID)
	return nil
}
