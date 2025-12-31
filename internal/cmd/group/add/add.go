// Package add provides the group add subcommand.
package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// Options holds the options for the add command.
type Options struct {
	Factory   *cmdutil.Factory
	Devices   []string
	GroupName string
}

// NewCommand creates the group add command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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

	added := 0
	for _, device := range opts.Devices {
		err := mgr.AddDeviceToGroup(opts.GroupName, device)
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
		ios.Success("Added 1 device to group %q", opts.GroupName)
	} else {
		ios.Success("Added %d devices to group %q", added, opts.GroupName)
	}

	return nil
}
