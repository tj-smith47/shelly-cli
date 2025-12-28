// Package virtual provides the virtual component commands.
package virtual

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual/del"
	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual/get"
	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/virtual/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the virtual command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "virtual",
		Aliases: []string{"virt", "vc"},
		Short:   "Manage virtual components",
		Long: `Manage virtual components on Shelly Gen2+ devices.

Virtual components allow you to create custom boolean, number, text, enum,
button, and group components that can be used in scripts and automations.

Virtual component IDs are automatically assigned in the range 200-299.`,
		Example: `  # List virtual components on a device
  shelly virtual list kitchen

  # Create a virtual boolean
  shelly virtual create kitchen boolean --name "Away Mode"

  # Get a virtual component value
  shelly virtual get kitchen boolean:200

  # Set a virtual component value
  shelly virtual set kitchen boolean:200 true

  # Delete a virtual component
  shelly virtual delete kitchen boolean:200`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(get.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(del.NewCommand(f))

	return cmd
}
