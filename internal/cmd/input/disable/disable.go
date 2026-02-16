// Package disable provides the input disable command.
package disable

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the input disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewInputToggleCommand(f, factories.InputToggleOpts{
		Enable: false,
		Long: `Disable an input component on a Shelly device.

When disabled, the input will not respond to physical button presses or
switch state changes. This can be useful for maintenance or to prevent
accidental triggers.`,
		Example: `  # Disable input on a device
  shelly input disable kitchen

  # Disable specific input by ID
  shelly input disable living-room --id 1

  # Using alias
  shelly input off bedroom`,
	})
}
