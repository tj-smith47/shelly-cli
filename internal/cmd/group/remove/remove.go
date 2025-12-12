// Package remove provides the group remove subcommand.
package remove

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the group remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <group> <device>...",
		Aliases: []string{"rm"},
		Short:   "Remove devices from a group",
		Long: `Remove one or more devices from a group.

Devices can be specified by their registered name or IP address.
Removing a device from a group does not delete the device.`,
		Example: `  # Remove a single device from a group
  shelly group remove living-room light-1

  # Remove multiple devices
  shelly group remove living-room light-1 light-2 switch-1

  # Using alias
  shelly group rm bedroom lamp

  # Short form
  shelly grp rm office 192.168.1.100`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(args[0], args[1:])
		},
	}

	return cmd
}

func run(groupName string, devices []string) error {
	// Check if group exists
	if _, exists := config.GetGroup(groupName); !exists {
		return fmt.Errorf("group %q not found", groupName)
	}

	removed := 0
	for _, device := range devices {
		err := config.RemoveDeviceFromGroup(groupName, device)
		if err != nil {
			iostreams.Warning("Failed to remove %q: %v", device, err)
			continue
		}
		removed++
	}

	if removed == 0 {
		return fmt.Errorf("no devices were removed")
	}

	if removed == 1 {
		iostreams.Success("Removed 1 device from group %q", groupName)
	} else {
		iostreams.Success("Removed %d devices from group %q", removed, groupName)
	}

	return nil
}
