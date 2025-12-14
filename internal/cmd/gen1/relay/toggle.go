package relay

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// ToggleOptions holds toggle command options.
type ToggleOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newToggleCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &ToggleOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "toggle <device>",
		Aliases: []string{"flip", "t"},
		Short:   "Toggle relay state",
		Long:    `Toggle a Gen1 relay switch between on and off states.`,
		Example: `  # Toggle relay
  shelly gen1 relay toggle living-room

  # Toggle relay 1 (second relay)
  shelly gen1 relay toggle living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runToggle(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Relay")

	return cmd
}

func runToggle(ctx context.Context, opts *ToggleOptions) error {
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

	ios.StartProgress("Toggling relay...")

	err = relay.Toggle(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Relay %d toggled", opts.ID)

	return nil
}
