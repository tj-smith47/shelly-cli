package light

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// OnOptions holds on command options.
type OnOptions struct {
	Factory    *cmdutil.Factory
	Device     string
	ID         int
	Brightness int
	Duration   int
}

func newOnCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OnOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "on <device>",
		Aliases: []string{"enable", "1"},
		Short:   "Turn light on",
		Long: `Turn on a Gen1 light/dimmer.

Optionally specify brightness (0-100) and/or duration for auto-off.`,
		Example: `  # Turn light on
  shelly gen1 light on living-room

  # Turn on at 50% brightness
  shelly gen1 light on living-room --brightness 50

  # Turn on for 60 seconds
  shelly gen1 light on living-room --duration 60`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOn(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Light")
	cmd.Flags().IntVarP(&opts.Brightness, "brightness", "b", 0, "Set brightness (0-100, 0 = unchanged)")
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

	light, err := gen1Client.Light(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Turning light on...")

	switch {
	case opts.Brightness > 0:
		err = light.TurnOnWithBrightness(ctx, opts.Brightness)
	case opts.Duration > 0:
		err = light.TurnOnForDuration(ctx, opts.Duration)
	default:
		err = light.TurnOn(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	switch {
	case opts.Brightness > 0:
		ios.Success("Light %d turned on at %d%% brightness", opts.ID, opts.Brightness)
	case opts.Duration > 0:
		ios.Success("Light %d turned on for %d seconds", opts.ID, opts.Duration)
	default:
		ios.Success("Light %d turned on", opts.ID)
	}

	return nil
}
