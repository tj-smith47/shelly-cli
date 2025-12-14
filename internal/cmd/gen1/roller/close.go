package roller

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CloseOptions holds close command options.
type CloseOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration float64
}

func newCloseCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &CloseOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "close <device>",
		Aliases: []string{"down", "lower"},
		Short:   "Close roller/cover",
		Long: `Start closing a Gen1 roller/cover.

Optionally specify a duration to close for a specific time.`,
		Example: `  # Close roller fully
  shelly gen1 roller close living-room

  # Close for 5 seconds
  shelly gen1 roller close living-room --duration 5`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runClose(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")
	cmd.Flags().Float64Var(&opts.Duration, "duration", 0, "Close for specified seconds (0 = fully close)")

	return cmd
}

func runClose(ctx context.Context, opts *CloseOptions) error {
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

	ios.StartProgress("Closing roller...")

	if opts.Duration > 0 {
		err = roller.CloseForDuration(ctx, opts.Duration)
	} else {
		err = roller.Close(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Roller %d closing for %.1f seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Roller %d closing", opts.ID)
	}

	return nil
}
