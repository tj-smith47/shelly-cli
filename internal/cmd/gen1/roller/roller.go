// Package roller provides Gen1 roller/cover control commands.
package roller

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 roller command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "roller",
		Aliases: []string{"cover", "shutter", "blind"},
		Short:   "Control Gen1 roller/covers",
		Long: `Control roller/cover/shutter components on Gen1 Shelly devices.

Gen1 rollers are found in devices like:
- Shelly 2.5 (in roller mode)
- Shelly 2

For Gen2+ devices, use 'shelly cover' instead.`,
		Example: `  # Open roller
  shelly gen1 roller open living-room

  # Close roller
  shelly gen1 roller close living-room

  # Stop roller movement
  shelly gen1 roller stop living-room

  # Go to specific position (0-100)
  shelly gen1 roller position living-room 50

  # Check roller status
  shelly gen1 roller status living-room`,
	}

	cmd.AddCommand(newOpenCommand(f))
	cmd.AddCommand(newCloseCommand(f))
	cmd.AddCommand(newStopCommand(f))
	cmd.AddCommand(newPositionCommand(f))
	cmd.AddCommand(newStatusCommand(f))
	cmd.AddCommand(newCalibrateCommand(f))

	return cmd
}
