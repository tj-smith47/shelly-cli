package roller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// PositionOptions holds position command options.
type PositionOptions struct {
	Factory  *cmdutil.Factory
	Device   string
	ID       int
	Position int
}

func newPositionCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &PositionOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "position <device> <position>",
		Aliases: []string{"pos", "goto"},
		Short:   "Move roller to position",
		Long: `Move a Gen1 roller/cover to a specific position.

Position range: 0 (fully closed) to 100 (fully open).

Note: The roller must be calibrated for position control to work.
Use 'shelly gen1 roller calibrate' to calibrate first.`,
		Example: `  # Move to 50% open
  shelly gen1 roller position living-room 50

  # Fully open
  shelly gen1 roller position living-room 100

  # Fully close
  shelly gen1 roller position living-room 0`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			pos, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid position: %s (must be 0-100)", args[1])
			}
			if pos < 0 || pos > 100 {
				return fmt.Errorf("position must be 0-100, got %d", pos)
			}
			opts.Position = pos
			return runPosition(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Roller")

	return cmd
}

func runPosition(ctx context.Context, opts *PositionOptions) error {
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

	ios.StartProgress("Moving roller to position...")

	err = roller.GoToPosition(ctx, opts.Position)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Roller %d moving to %d%%", opts.ID, opts.Position)

	return nil
}
