package roller

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// OpenOptions holds open command options.
type OpenOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Duration float64
}

func newOpenCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &OpenOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "open <device>",
		Aliases: []string{"up", "raise"},
		Short:   "Open roller/cover",
		Long: `Start opening a Gen1 roller/cover.

Optionally specify a duration to open for a specific time.`,
		Example: `  # Open roller fully
  shelly gen1 roller open living-room

  # Open for 5 seconds
  shelly gen1 roller open living-room --duration 5

  # Open roller 1 (second roller)
  shelly gen1 roller open living-room --id 1`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runOpen(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")
	cmd.Flags().Float64Var(&opts.Duration, "duration", 0, "Open for specified seconds (0 = fully open)")

	return cmd
}

func runOpen(ctx context.Context, opts *OpenOptions) error {
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

	ios.StartProgress("Opening roller...")

	if opts.Duration > 0 {
		err = roller.OpenForDuration(ctx, opts.Duration)
	} else {
		err = roller.Open(ctx)
	}

	ios.StopProgress()

	if err != nil {
		return err
	}

	if opts.Duration > 0 {
		ios.Success("Roller %d opening for %.1f seconds", opts.ID, opts.Duration)
	} else {
		ios.Success("Roller %d opening", opts.ID)
	}

	return nil
}
