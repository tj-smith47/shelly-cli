// Package toggle provides the quick toggle command.
package toggle

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
)

// NewCommand creates the toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewQuickCommand(f, factories.QuickOpts{
		Action:  factories.QuickToggle,
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle a device (auto-detects type)",
		Long: `Toggle a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this toggles between open and close based on current state.

By default, toggles all controllable components on the device.
Use --id to target a specific component (e.g., for multi-switch devices).`,
		Example: `  # Toggle all components on a device
  shelly toggle living-room

  # Toggle specific switch (for multi-switch devices)
  shelly toggle dual-switch --id 1

  # Toggle a cover
  shelly toggle bedroom-blinds`,
		SpinnerText:     "Toggling...",
		SuccessSingular: "Device %q toggled",
		SuccessPlural:   "Toggled %d components on %q",
	})
}
