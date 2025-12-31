// Package off provides the fleet off subcommand.
package off

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the fleet off command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewFleetRelayCommand(f, factories.FleetRelayOpts{
		Action:  shelly.RelayOff,
		Aliases: []string{"turn-off", "disable"},
		Short:   "Turn off devices via cloud",
		Long: `Turn off devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Turn off specific device
  shelly fleet off device-id

  # Turn off all devices in a group
  shelly fleet off --group living-room

  # Turn off all relay devices
  shelly fleet off --all`,
		SuccessVerb: "turned off",
	})
}
