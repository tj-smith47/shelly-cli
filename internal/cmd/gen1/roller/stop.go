package roller

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// StopOptions holds stop command options.
type StopOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
}

func newStopCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &StopOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "stop <device>",
		Aliases: []string{"halt", "pause"},
		Short:   "Stop roller movement",
		Long:    `Stop a moving Gen1 roller/cover.`,
		Example: `  # Stop roller
  shelly gen1 roller stop living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runStop(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")

	return cmd
}

func runStop(ctx context.Context, opts *StopOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	roller, err := gen1Client.Roller(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Stopping roller...")

	err = roller.Stop(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Roller %d stopped", opts.ID)

	return nil
}
