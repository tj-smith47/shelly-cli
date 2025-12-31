// Package remove provides the group remove subcommand.
package remove

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the options for the remove command.
type Options struct {
	Factory   *cmdutil.Factory
	Devices   []string
	GroupName string
}

// NewCommand creates the group remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.GroupName = args[0]
			opts.Devices = args[1:]
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if group exists
	if opts.Factory.GetGroup(opts.GroupName) == nil {
		return fmt.Errorf("group %q not found", opts.GroupName)
	}

	// Get config manager for mutations
	mgr, err := opts.Factory.ConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	removed := 0
	for _, device := range opts.Devices {
		err := mgr.RemoveDeviceFromGroup(opts.GroupName, device)
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
		ios.Success("Removed 1 device from group %q", opts.GroupName)
	} else {
		ios.Success("Removed %d devices from group %q", removed, opts.GroupName)
	}

	return nil
}
