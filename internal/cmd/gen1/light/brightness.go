package light

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// BrightnessOptions holds brightness command options.
type BrightnessOptions struct {
	Factory    *cmdutil.Factory
	Device     string
	ID         int
	Brightness int
	Transition int
}

func newBrightnessCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &BrightnessOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "brightness <device> <level>",
		Aliases: []string{"br", "dim"},
		Short:   "Set light brightness",
		Long: `Set the brightness level of a Gen1 light/dimmer.

Brightness range: 0 (off) to 100 (full brightness).

Optionally specify a transition time for smooth dimming.`,
		Example: `  # Set brightness to 75%
  shelly gen1 light brightness living-room 75

  # Set brightness with 1 second transition
  shelly gen1 light brightness living-room 50 --transition 1000`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			brightness, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid brightness: %s (must be 0-100)", args[1])
			}
			if brightness < 0 || brightness > 100 {
				return fmt.Errorf("brightness must be 0-100, got %d", brightness)
			}
			opts.Brightness = brightness
			return runBrightness(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Light")
	cmd.Flags().IntVar(&opts.Transition, "transition", 0, "Transition time in milliseconds (0 = instant)")

	return cmd
}

func runBrightness(ctx context.Context, opts *BrightnessOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Setting brightness...")

	light := gen1Client.Light(opts.ID)
	if opts.Transition > 0 {
		err = light.SetBrightnessWithTransition(ctx, opts.Brightness, opts.Transition)
	} else {
		err = light.SetBrightness(ctx, opts.Brightness)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Light %d brightness set to %d%%", opts.ID, opts.Brightness)

	return nil
}
