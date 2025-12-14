package relay

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
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration int
}

func newOffCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OffOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "off <device>",
		Aliases: []string{"disable", "0"},
		Short:   "Turn relay off",
		Long: `Turn off a Gen1 relay switch.

Optionally specify a duration to auto-turn-on after the specified seconds.`,
		Example: `  # Turn relay off
  shelly gen1 relay off living-room

  # Turn off for 60 seconds (then auto-on)
  shelly gen1 relay off living-room --duration 60

  # Turn off relay 1 (second relay)
  shelly gen1 relay off living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOff(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Relay")
	cmd.Flags().IntVar(&opts.Duration, "duration", 0, "Auto-on after seconds (0 = disabled)")

	return cmd
}

func runOff(ctx context.Context, opts *OffOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	relay, err := gen1Client.Relay(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Turning relay off...")

	if opts.Duration > 0 {
		err = relay.TurnOffForDuration(ctx, opts.Duration)
	} else {
		err = relay.TurnOff(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Relay %d turned off for %d seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Relay %d turned off", opts.ID)
	}

	return nil
}
