package color

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// SetOptions holds set command options.
type SetOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	Red     int
	Green   int
	Blue    int
	White   int
	Gain    int
}

func newSetCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &SetOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device> <red> <green> <blue>",
		Aliases: []string{"rgb", "c"},
		Short:   "Set RGB color",
		Long: `Set the RGB color values of a Gen1 RGBW light.

RGB values range from 0-255. Optionally include white channel and gain.`,
		Example: `  # Set to red
  shelly gen1 color set living-room 255 0 0

  # Set to purple with gain
  shelly gen1 color set living-room 128 0 255 --gain 75

  # Set RGBW (with white channel)
  shelly gen1 color set living-room 255 200 100 --white 50`,
		Args:              cobra.ExactArgs(4),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]

			red, err := strconv.Atoi(args[1])
			if err != nil || red < 0 || red > 255 {
				return fmt.Errorf("invalid red value: %s (must be 0-255)", args[1])
			}
			opts.Red = red

			green, err := strconv.Atoi(args[2])
			if err != nil || green < 0 || green > 255 {
				return fmt.Errorf("invalid green value: %s (must be 0-255)", args[2])
			}
			opts.Green = green

			blue, err := strconv.Atoi(args[3])
			if err != nil || blue < 0 || blue > 255 {
				return fmt.Errorf("invalid blue value: %s (must be 0-255)", args[3])
			}
			opts.Blue = blue

			return runSet(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")
	cmd.Flags().IntVarP(&opts.White, "white", "w", 0, "White channel value (0-255)")
	cmd.Flags().IntVarP(&opts.Gain, "gain", "g", 0, "Brightness/gain (0-100)")

	return cmd
}

func runSet(ctx context.Context, opts *SetOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	color, err := gen1Client.Color(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Setting color...")

	switch {
	case opts.Gain > 0:
		err = color.TurnOnWithRGB(ctx, opts.Red, opts.Green, opts.Blue, opts.Gain)
	case opts.White > 0:
		err = color.SetRGBW(ctx, opts.Red, opts.Green, opts.Blue, opts.White)
	default:
		err = color.SetRGB(ctx, opts.Red, opts.Green, opts.Blue)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d set to RGB(%d, %d, %d)", opts.ID, opts.Red, opts.Green, opts.Blue)

	return nil
}
