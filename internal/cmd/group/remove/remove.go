// Package remove provides the group remove subcommand.
package remove

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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

	// Get config manager for mutations
	mgr, err := f.ConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	removed := 0
	for _, device := range devices {
		err := mgr.RemoveDeviceFromGroup(groupName, device)
		if err != nil {
			ios.Warning("Failed to remove %q: %v", device, err)
			continue
		}
		removed++
	}

	if removed == 0 {
		return fmt.Errorf("no devices were removed")
	}

	if removed == 1 {
		ios.Success("Removed 1 device from group %q", groupName)
	} else {
		ios.Success("Removed %d devices from group %q", removed, groupName)
	}

	return nil
}
