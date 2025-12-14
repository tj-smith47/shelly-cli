// Package color provides Gen1 RGBW color control commands.
package color

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 color command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "color",
		Aliases: []string{"rgb", "rgbw"},
		Short:   "Control Gen1 RGBW lights",
		Long: `Control RGBW color lights on Gen1 Shelly devices.

Gen1 color lights are found in devices like:
- Shelly RGBW, RGBW2
- Shelly Bulb (color mode)

For Gen2+ devices, use 'shelly rgb' instead.`,
		Example: `  # Turn color light on
  shelly gen1 color on living-room

  # Set RGB color (red=255, green=128, blue=0)
  shelly gen1 color set living-room 255 128 0

  # Set gain (brightness)
  shelly gen1 color gain living-room 75

  # Check color status
  shelly gen1 color status living-room`,
	}

	cmd.AddCommand(newOnCommand(f))
	cmd.AddCommand(newOffCommand(f))
	cmd.AddCommand(newToggleCommand(f))
	cmd.AddCommand(newSetCommand(f))
	cmd.AddCommand(newGainCommand(f))
	cmd.AddCommand(newStatusCommand(f))

	return cmd
}
