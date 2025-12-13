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
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the device command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"dev"},
		Short:   "Manage Shelly devices",
		Long: `Manage Shelly devices in your registry.

Device commands allow you to add, remove, list, and manage registered devices.
Registered devices can be referenced by name in other commands.`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(info.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(ping.NewCommand(f))
	cmd.AddCommand(reboot.NewCommand(f))
	cmd.AddCommand(factoryreset.NewCommand(f))

	return cmd
}
