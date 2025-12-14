package color

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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
		Use:   "toggle <device>",
		Short: "Toggle color light state",
		Long:  `Toggle a Gen1 RGBW color light between on and off states.`,
		Example: `  # Toggle color light
  shelly gen1 color toggle living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runToggle(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")

	return cmd
}

func runToggle(ctx context.Context, opts *ToggleOptions) error {
	ios := opts.Factory.IOStreams()

	gen1Client, err := connectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	ios.StartProgress("Toggling color light...")

	color := gen1Client.Color(opts.ID)
	err = color.Toggle(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d toggled", opts.ID)

	return nil
}
