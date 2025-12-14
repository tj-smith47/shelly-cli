package color

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// OnOptions holds on command options.
type OnOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration int
}

func newOnCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OnOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "on <device>",
		Aliases: []string{"enable", "1"},
		Short:   "Turn color light on",
		Long:    `Turn on a Gen1 RGBW color light.`,
		Example: `  # Turn color light on
  shelly gen1 color on living-room

  # Turn on for 60 seconds
  shelly gen1 color on living-room --duration 60`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOn(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")
	cmd.Flags().IntVar(&opts.Duration, "duration", 0, "Auto-off after seconds (0 = disabled)")

	return cmd
}

func runOn(ctx context.Context, opts *OnOptions) error {
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

	ios.StartProgress("Turning color light on...")

	if opts.Duration > 0 {
		err = color.TurnOnForDuration(ctx, opts.Duration)
	} else {
		err = color.TurnOn(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Color %d turned on for %d seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Color %d turned on", opts.ID)
	}

	return nil
}
