// Package bthome provides BTHome management commands.
package bthome

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome/add"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/bthome/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the bthome command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bthome",
		Aliases: []string{"bt", "bluetooth"},
		Short:   "Manage BTHome Bluetooth devices",
		Long: `Manage BTHome Bluetooth devices on Shelly gateways.

BTHome is a protocol for Bluetooth Low Energy (BLE) devices that enables
local communication without cloud dependency. Supported on Gen2 Pro*,
Gen3, and Gen4 devices with Bluetooth capability.

BTHome devices include sensors (temperature, humidity, door/window),
buttons, and other BLE peripherals that broadcast data in BTHome format.`,
		Example: `  # List BTHome devices on a gateway
  shelly bthome list living-room

  # Start discovery for new devices
  shelly bthome add living-room

  # Get status of a specific BTHome device
  shelly bthome status living-room 200

  # Remove a BTHome device
  shelly bthome remove living-room 200`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(add.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))

	return cmd
}
