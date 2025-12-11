// Package device provides the device command group for device management.
package device

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/device/factoryreset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/info"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/ping"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/reboot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/device/status"
)

// NewCommand creates the device command group.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"dev"},
		Short:   "Manage Shelly devices",
		Long: `Manage Shelly devices in your registry.

Device commands allow you to add, remove, list, and manage registered devices.
Registered devices can be referenced by name in other commands.`,
	}

	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(info.NewCommand())
	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(ping.NewCommand())
	cmd.AddCommand(reboot.NewCommand())
	cmd.AddCommand(factoryreset.NewCommand())

	return cmd
}
