// Package on provides the fleet on subcommand.
package on

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the fleet on command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewFleetRelayCommand(f, factories.FleetRelayOpts{
		Action:  shelly.RelayOn,
		Aliases: []string{"turn-on", "enable"},
		Short:   "Turn on devices via cloud",
		Long: `Turn on devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token`,
		Example: `  # Turn on specific device
  shelly fleet on device-id

  # Turn on all devices in a group
  shelly fleet on --group living-room

  # Turn on all relay devices
  shelly fleet on --all`,
		SuccessVerb: "turned on",
	})
}
