// Package set provides the rgb set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the rgb set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		rgbID      int
		red        int
		green      int
		blue       int
		brightness int
		on         bool
	)

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
			return run(cmd.Context(), f, args[0], rgbID, red, green, blue, brightness, on)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &rgbID, "RGB")
	cmd.Flags().IntVarP(&red, "red", "r", -1, "Red value (0-255)")
	cmd.Flags().IntVarP(&green, "green", "g", -1, "Green value (0-255)")
	cmd.Flags().IntVarP(&blue, "blue", "b", -1, "Blue value (0-255)")
	cmd.Flags().IntVar(&brightness, "brightness", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&on, "on", false, "Turn on")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, rgbID, red, green, blue, brightness int, on bool) error {
	params := shelly.BuildRGBSetParams(red, green, blue, brightness, on)

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()
	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, "Setting RGB parameters...", func(ctx context.Context) error {
		return svc.RGBSet(ctx, device, rgbID, params)
	})
	if err != nil {
		return fmt.Errorf("failed to set RGB parameters: %w", err)
	}

	ios.Success("RGB %d parameters set", rgbID)
	return nil
}
