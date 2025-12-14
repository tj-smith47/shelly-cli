package color

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
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
		Use:     "off <device>",
		Aliases: []string{"disable", "0"},
		Short:   "Turn color light off",
		Long:    `Turn off a Gen1 RGBW color light.`,
		Example: `  # Turn color light off
  shelly gen1 color off living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
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

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	color, err := gen1Client.Color(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Turning color light off...")

	err = color.TurnOff(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d turned off", opts.ID)

	return nil
}
