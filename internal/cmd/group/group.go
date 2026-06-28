// Package group provides the group command group for device group management.
package group

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/group/add"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/members"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/off"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/on"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/toggle"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the group command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "group",
		Aliases: []string{"grp"},
		Short:   "Manage device groups",
		Long: `Manage device groups for batch operations.

Groups allow you to organize devices and perform bulk operations on them.
Devices can belong to multiple groups.`,
		Example: `  # List all groups
  shelly group list

  # Create a new group
  shelly group create living-room

  # Add devices to a group
  shelly group add living-room light-1 switch-2

  # Show group members
  shelly group members living-room

  # Remove a device from a group
  shelly group remove living-room switch-2

  # Delete a group
  shelly group delete living-room

  # Set all members to 100% and turn on
  shelly group set guest-bath-bulbs -b 100 --on

  # Turn a group on, off, or toggle it
  shelly group on living-room
  shelly group off living-room
  shelly group toggle living-room`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(add.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))
	cmd.AddCommand(members.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(on.NewCommand(f))
	cmd.AddCommand(off.NewCommand(f))
	cmd.AddCommand(toggle.NewCommand(f))

	return cmd
}
