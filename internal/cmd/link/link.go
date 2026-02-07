// Package link provides the link command group for device link management.
package link

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/link/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/link/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/link/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/link/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the link command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "link",
		Aliases: []string{"ln"},
		Short:   "Manage device power links",
		Long: `Manage parent-child power relationships between devices.

Links define which switch controls the power to another device.
When a linked child device is offline, its state is derived from
the parent switch state. Control commands (on/off/toggle) automatically
proxy to the parent switch when the child is unreachable.`,
		Example: `  # Link a bulb to a switch (bulb is powered by switch:0)
  shelly link set bulb-duo bedroom-2pm

  # Link with a specific switch ID
  shelly link set garage-light garage-switch --switch-id 1

  # List all links
  shelly link list

  # Show link status with derived state
  shelly link status

  # Remove a link
  shelly link delete bulb-duo`,
	}

	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))

	return cmd
}
