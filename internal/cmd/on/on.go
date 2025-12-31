// Package on provides the quick on command.
package on

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewQuickCommand(f, factories.QuickOpts{
		Action:  factories.QuickOn,
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on a device (auto-detects type)",
		Long: `Turn on a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this opens them. For switches/lights/RGB, this turns them on.

By default, turns on all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Turn on all components on a device
  shelly on living-room

  # Turn on specific switch (for multi-switch devices)
  shelly on dual-switch --id 1

  # Open a cover
  shelly on bedroom-blinds`,
		SpinnerText:     "Turning on...",
		SuccessSingular: "Device %q turned on",
		SuccessPlural:   "Turned on %d components on %q",
	})
}
