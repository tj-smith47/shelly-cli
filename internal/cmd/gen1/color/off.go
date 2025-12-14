package color

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// OffOptions holds off command options.
type OffOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newOffCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OffOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "off <device>",
		Short: "Turn color light off",
		Long:  `Turn off a Gen1 RGBW color light.`,
		Example: `  # Turn color light off
  shelly gen1 color off living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOff(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")

	return cmd
}

func runOff(ctx context.Context, opts *OffOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Turning color light off...")

	color := gen1Client.Color(opts.ID)
	err = color.TurnOff(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d turned off", opts.ID)

	return nil
}
