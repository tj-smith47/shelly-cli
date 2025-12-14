// Package light provides Gen1 light/dimmer control commands.
package light

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 light command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "light",
		Aliases: []string{"dimmer", "lamp"},
		Short:   "Control Gen1 lights/dimmers",
		Long: `Control light/dimmer components on Gen1 Shelly devices.

Gen1 lights are found in devices like:
- Shelly Dimmer, Dimmer 2
- Shelly Bulb, Vintage
- Shelly Duo

For Gen2+ devices, use 'shelly light' instead.`,
		Example: `  # Turn light on
  shelly gen1 light on living-room

  # Set brightness
  shelly gen1 light brightness living-room 75

  # Toggle light
  shelly gen1 light toggle living-room

  # Check light status
  shelly gen1 light status living-room`,
	}

	cmd.AddCommand(newOnCommand(f))
	cmd.AddCommand(newOffCommand(f))
	cmd.AddCommand(newToggleCommand(f))
	cmd.AddCommand(newBrightnessCommand(f))
	cmd.AddCommand(newStatusCommand(f))

	return cmd
}
