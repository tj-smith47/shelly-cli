package color

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/gen1/connutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// GainOptions holds gain command options.
type GainOptions struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	Gain    int
}

func newGainCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &GainOptions{Factory: f}

	cmd := &cobra.Command{
		Use:     "gain <device> <level>",
		Aliases: []string{"brightness", "br"},
		Short:   "Set color light gain/brightness",
		Long: `Set the gain (brightness) level of a Gen1 RGBW light.

Gain range: 0 (off) to 100 (full brightness).`,
		Example: `  # Set gain to 75%
  shelly gen1 color gain living-room 75`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			gain, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid gain: %s (must be 0-100)", args[1])
			}
			if gain < 0 || gain > 100 {
				return fmt.Errorf("gain must be 0-100, got %d", gain)
			}
			opts.Gain = gain
			return runGain(cmd.Context(), opts)
		},
	}

	cmdutil.AddComponentIDFlag(cmd, &opts.ID, "Color")

	return cmd
}

func runGain(ctx context.Context, opts *GainOptions) error {
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

	ios.StartProgress("Setting gain...")

	err = color.SetGain(ctx, opts.Gain)

	ios.StopProgress()

	if err != nil {
		return err
	}

	ios.Success("Color %d gain set to %d%%", opts.ID, opts.Gain)

	return nil
}
