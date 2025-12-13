// Package ethernet provides Ethernet configuration commands.
package ethernet

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/ethernet/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/ethernet/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the ethernet command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ethernet",
		Aliases: []string{"eth"},
		Short:   "Manage device Ethernet configuration",
		Long: `Manage device Ethernet configuration settings.

Ethernet is available on Shelly Pro devices that have an Ethernet port.
It provides wired network connectivity as an alternative to WiFi.`,
		Example: `  # Show Ethernet status
  shelly ethernet status living-room-pro

  # Configure Ethernet with DHCP
  shelly ethernet set living-room-pro --enable

  # Configure Ethernet with static IP
  shelly ethernet set living-room-pro --enable --static-ip "192.168.1.50" \
    --gateway "192.168.1.1" --netmask "255.255.255.0"`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))

	return cmd
}
