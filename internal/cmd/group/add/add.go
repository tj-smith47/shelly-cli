// Package add provides the group add subcommand.
package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the group add command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <group> <device>...",
		Aliases: []string{"append", "include"},
		Short:   "Add devices to a group",
		Long: `Add one or more devices to a group.

Devices can be specified by their registered name or IP address.
Devices can belong to multiple groups.`,
		Example: `  # Add a single device to a group
  shelly group add living-room light-1

  # Add multiple devices
  shelly group add living-room light-1 light-2 switch-1

  # Add by IP address
  shelly group add office 192.168.1.100

  # Short form
  shelly grp add bedroom lamp`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0], args[1:])
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, groupName string, devices []string) error {
	ios := f.IOStreams()

	// Check if group exists
	if f.GetGroup(groupName) == nil {
		return fmt.Errorf("group %q not found", groupName)
	}

	added := 0
	for _, device := range devices {
		err := config.AddDeviceToGroup(groupName, device)
		if err != nil {
			ios.Warning("Failed to add %q: %v", device, err)
			continue
		}
		added++
	}

	if added == 0 {
		return fmt.Errorf("no devices were added")
	}

	if added == 1 {
		ios.Success("Added 1 device to group %q", groupName)
	} else {
		ios.Success("Added %d devices to group %q", added, groupName)
	}

	return nil
}
