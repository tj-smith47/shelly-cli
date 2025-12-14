// Package zigbee provides Zigbee management commands.
package zigbee

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee/pair"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/zigbee/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the zigbee command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zigbee",
		Aliases: []string{"zb", "z2m"},
		Short:   "Manage Zigbee connectivity",
		Long: `Manage Zigbee connectivity on Shelly devices.

Zigbee support is available on Gen4 devices that can operate as
Zigbee end devices, connecting to Zigbee coordinators like
Home Assistant (ZHA), Zigbee2MQTT, or other compatible systems.

When operating in Zigbee mode, the device joins a Zigbee network
and can be controlled through the Zigbee coordinator instead of
or in addition to WiFi/HTTP control.`,
		Example: `  # Show Zigbee status
  shelly zigbee status living-room

  # Start pairing to join a network
  shelly zigbee pair living-room

  # List Zigbee-capable devices on network
  shelly zigbee list`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(pair.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))

	return cmd
}
