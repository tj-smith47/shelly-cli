// Package off provides the quick off command.
package off

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewQuickCommand(f, factories.QuickOpts{
		Action:  factories.QuickOff,
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off a device (auto-detects type)",
		Long: `Turn off a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this closes them. For switches/lights/RGB, this turns them off.

By default, turns off all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Turn off all components on a device
  shelly off living-room

  # Turn off specific switch (for multi-switch devices)
  shelly off dual-switch --id 1

  # Close a cover
  shelly off bedroom-blinds`,
		SpinnerText:     "Turning off...",
		SuccessSingular: "Device %q turned off",
		SuccessPlural:   "Turned off %d components on %q",
	})
}
