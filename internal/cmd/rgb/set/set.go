// Package set provides the rgb set subcommand.
package set

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
	Factory    *cmdutil.Factory
	Device     string
	Red        int
	Green      int
	Blue       int
	Brightness int
	On         bool
}

// NewCommand creates the rgb set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:    f,
		Red:        -1,
		Green:      -1,
		Blue:       -1,
		Brightness: -1,
	}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"color", "c"},
		Short:   "Set RGB parameters",
		Long: `Set parameters of an RGB light component on the specified device.

You can set color values (red, green, blue), brightness, and on/off state.
Values not specified will be left unchanged.`,
		Example: `  # Set RGB color to red
  shelly rgb set living-room --red 255 --green 0 --blue 0

  # Set RGB with brightness
  shelly rgb color living-room -r 0 -g 255 -b 128 --brightness 75`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, "RGB")
	cmd.Flags().IntVarP(&opts.Red, "red", "r", -1, "Red value (0-255)")
	cmd.Flags().IntVarP(&opts.Green, "green", "g", -1, "Green value (0-255)")
	cmd.Flags().IntVarP(&opts.Blue, "blue", "b", -1, "Blue value (0-255)")
	cmd.Flags().IntVar(&opts.Brightness, "brightness", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&opts.On, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	f := opts.Factory
	params := shelly.BuildRGBSetParams(opts.Red, opts.Green, opts.Blue, opts.Brightness, opts.On)

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting RGB parameters...", func(ctx context.Context) error {
		return svc.RGBSet(ctx, opts.Device, opts.ID, params)
	})
	if err != nil {
		return fmt.Errorf("failed to set RGB parameters: %w", err)
	}

	ios.Success("RGB %d parameters set", opts.ID)
	return nil
}
