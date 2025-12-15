// Package fleet provides cloud-based fleet management commands.
package fleet

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/connect"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/disconnect"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/health"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/stats"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/fleet/toggle"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the fleet command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fleet",
		Aliases: []string{"cloud-fleet", "cf"},
		Short:   "Cloud-based fleet management",
		Long: `Manage devices through Shelly Cloud using the integrator API.

Fleet commands use WebSocket connections to the Shelly Cloud for
real-time device monitoring and control across your entire fleet.

Configuration:
  Set SHELLY_INTEGRATOR_TAG and SHELLY_INTEGRATOR_TOKEN environment variables,
  or configure them in the config file.

Note: This differs from local batch commands. Fleet operations go through
the cloud, while batch commands communicate directly with devices on your LAN.`,
		Example: `  # Connect to cloud hosts
  shelly fleet connect

  # View fleet status
  shelly fleet status

  # View fleet statistics
  shelly fleet stats

  # Check device health
  shelly fleet health

  # Control devices via cloud
  shelly fleet on --group living-room
  shelly fleet off --all`,
	}

	cmd.AddCommand(connect.NewCommand(f))
	cmd.AddCommand(disconnect.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(stats.NewCommand(f))
	cmd.AddCommand(health.NewCommand(f))
	cmd.AddCommand(on.NewCommand(f))
	cmd.AddCommand(off.NewCommand(f))
	cmd.AddCommand(toggle.NewCommand(f))

	return cmd
}
