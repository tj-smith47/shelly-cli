// Package matter provides Matter management commands.
package matter

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/matter/code"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/matter/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the matter command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "matter",
		Short: "Manage Matter connectivity",
		Long: `Manage Matter connectivity on Gen4+ Shelly devices.

Matter is a unified smart home connectivity standard that allows
devices from different manufacturers to work together. Shelly Gen4+
devices support Matter, enabling integration with Apple Home,
Google Home, Amazon Alexa, and other Matter-compatible controllers.

Key concepts:
- Fabric: A Matter network/ecosystem (e.g., Apple Home, Google Home)
- Commissioner: The app/controller that adds devices to a fabric
- Commissionable: Device is ready to be added to a fabric

Note: Matter support requires Gen4+ devices.`,
		Example: `  # Show Matter status
  shelly matter status living-room

  # Enable Matter on a device
  shelly matter enable living-room

  # Show pairing code for commissioning
  shelly matter code living-room

  # Reset Matter configuration
  shelly matter reset living-room --yes`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(enable.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))
	cmd.AddCommand(code.NewCommand(f))

	return cmd
}
