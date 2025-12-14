// Package relay provides Gen1 relay control commands.
package relay

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 relay command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relay",
		Aliases: []string{"sw", "switch"},
		Short:   "Control Gen1 relay switches",
		Long: `Control relay switches on Gen1 Shelly devices.

Gen1 relays are the basic switching elements found in devices like:
- Shelly 1, 1PM, 1L
- Shelly 2, 2.5
- Shelly Plug, Plug S
- Shelly 4Pro

For Gen2+ devices, use 'shelly switch' instead.`,
		Example: `  # Turn relay on
  shelly gen1 relay on living-room

  # Turn relay off
  shelly gen1 relay off living-room

  # Toggle relay state
  shelly gen1 relay toggle living-room

  # Check relay status
  shelly gen1 relay status living-room

  # Control specific relay (for multi-relay devices)
  shelly gen1 relay on living-room --id 1`,
	}

	cmd.AddCommand(newOnCommand(f))
	cmd.AddCommand(newOffCommand(f))
	cmd.AddCommand(newToggleCommand(f))
	cmd.AddCommand(newStatusCommand(f))

	return cmd
}
