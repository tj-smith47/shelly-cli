// Package switchcmd provides the switch command group for controlling relay switches.
package switchcmd

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/switchcmd/toggle"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the switch command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "switch",
		Aliases: []string{"sw"},
		Short:   "Control switch components",
		Long: `Control Shelly switch (relay) components.

Switches are the basic on/off relays found in most Shelly devices.
Use these commands to control individual switches or list all
switches on a device.`,
	}

	cmd.AddCommand(on.NewCommand(f))
	cmd.AddCommand(off.NewCommand(f))
	cmd.AddCommand(toggle.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))

	return cmd
}
