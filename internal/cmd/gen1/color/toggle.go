package color

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
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
		Use:     "toggle <device>",
		Aliases: []string{"flip", "t"},
		Short:   "Toggle color light state",
		Long:    `Toggle a Gen1 RGBW color light between on and off states.`,
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

	gen1Client, err := connutil.ConnectGen1(ctx, ios, opts.Device)
	if err != nil {
		return err
	}
	defer iostreams.CloseWithDebug("closing gen1 client", gen1Client)

	color, err := gen1Client.Color(opts.ID)
	if err != nil {
		return err
	}

	ios.StartProgress("Toggling color light...")

	err = color.Toggle(ctx)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d toggled", opts.ID)

	return nil
}
