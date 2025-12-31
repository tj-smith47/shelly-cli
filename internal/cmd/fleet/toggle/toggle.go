// Package toggle provides the fleet toggle subcommand.
package toggle

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the fleet toggle command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewFleetRelayCommand(f, factories.FleetRelayOpts{
		Action:  shelly.RelayToggle,
		Aliases: []string{"flip", "switch"},
		Short:   "Toggle devices via cloud",
		Long: `Toggle devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Toggle specific device
  shelly fleet toggle device-id

  # Toggle all devices in a group
  shelly fleet toggle --group living-room

  # Toggle all relay devices
  shelly fleet toggle --all`,
		SuccessVerb: "toggled",
	})
}
