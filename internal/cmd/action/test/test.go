// Package test provides the action test command.
package test

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the action test command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "test <device> <action>",
		Aliases: []string{"trigger", "fire"},
		Short:   "Test/trigger an action on a Gen1 device",
		Long: `Test (trigger) an action on a Gen1 Shelly device.

This simulates the event that would trigger the action URL, causing
the device to make the configured HTTP request.

Note: This feature is currently in development.

For Gen2 devices, actions are triggered differently - use the device's
built-in test functionality via the web interface.`,
		Example: `  # Test output on action
  shelly action test living-room out_on_url

  # For Gen1 devices, triggering actions typically requires
  # actually changing the device state:
  shelly switch on living-room  # triggers out_on_url`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams()

			ios.Warning("Gen1 action testing is not yet fully implemented.")
			ios.Println()
			ios.Info("Gen1 devices trigger actions based on state changes.")
			ios.Info("To trigger an action, change the device state:")
			ios.Info("  shelly on <device>   # triggers out_on_url")
			ios.Info("  shelly off <device>  # triggers out_off_url")
			ios.Println()
			ios.Info("Or use the device's web interface to test actions.")
			return nil
		},
	}

	return cmd
}
