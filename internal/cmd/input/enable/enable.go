// Package enable provides the input enable command.
package enable

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the input enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewInputToggleCommand(f, factories.InputToggleOpts{
		Enable: true,
		Long: `Enable an input component on a Shelly device.

When enabled, the input will respond to physical button presses or switch
state changes and trigger associated actions.`,
		Example: `  # Enable input on a device
  shelly input enable kitchen

  # Enable specific input by ID
  shelly input enable living-room --id 1

  # Using alias
  shelly input on bedroom`,
	})
}
