// Package group provides the group command group for device group management.
package group

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/group/add"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/members"
	"github.com/tj-smith47/shelly-cli/internal/cmd/group/remove"
)

// NewCommand creates the group command group.
func NewCommand() *cobra.Command {
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
  shelly group delete living-room`,
	}

	cmd.AddCommand(list.NewCommand())
	cmd.AddCommand(create.NewCommand())
	cmd.AddCommand(deletecmd.NewCommand())
	cmd.AddCommand(add.NewCommand())
	cmd.AddCommand(remove.NewCommand())
	cmd.AddCommand(members.NewCommand())

	return cmd
}
