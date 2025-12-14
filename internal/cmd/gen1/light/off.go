package light

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
		Short:   "Turn light off",
		Long:    `Turn off a Gen1 light/dimmer.`,
		Example: `  # Turn light off
  shelly gen1 light off living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOff(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Light")

	return cmd
}

func runOff(ctx context.Context, opts *OffOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	light, err := gen1Client.Light(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Turning light off...")

	err = light.TurnOff(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Light %d turned off", opts.ID)

	return nil
}
