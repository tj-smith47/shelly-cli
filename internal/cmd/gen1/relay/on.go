package relay

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
		Short:   "Turn relay on",
		Long: `Turn on a Gen1 relay switch.

Optionally specify a duration to auto-turn-off after the specified seconds.`,
		Example: `  # Turn relay on
  shelly gen1 relay on living-room

  # Turn on for 60 seconds
  shelly gen1 relay on living-room --duration 60

  # Turn on relay 1 (second relay)
  shelly gen1 relay on living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOn(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Relay")
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

	relay, err := gen1Client.Relay(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Turning relay on...")

	if opts.Duration > 0 {
		err = relay.TurnOnForDuration(ctx, opts.Duration)
	} else {
		err = relay.TurnOn(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Relay %d turned on for %d seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Relay %d turned on", opts.ID)
	}

	return nil
}
