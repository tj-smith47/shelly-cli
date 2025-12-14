// Package list provides the action list command.
package list

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the action list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "show"},
		Short:   "List action URLs for a Gen1 device",
		Long: `List all configured action URLs for a Gen1 Shelly device.

Gen1 devices use HTTP-based settings for action URLs. This command shows
all configured actions and their target URLs.

Note: This feature is currently in development. Gen1 device support requires
direct HTTP communication rather than the RPC protocol used by Gen2 devices.

Workaround: Access the device's web interface at http://<device-ip>/settings
to view and configure action URLs.`,
		Example: `  # List actions for a device
  shelly action list living-room

  # Workaround: use curl to get settings
  curl http://192.168.1.100/settings | jq '.actions'`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams()
			ios.Warning("Gen1 action URL management is not yet fully implemented.")
			ios.Println()
			ios.Info("Gen1 devices use a different API than Gen2 devices.")
			ios.Info("To view action URLs, access the device's web interface:")
			ios.Info("  http://<device-ip>/settings")
			ios.Println()
			ios.Info("Or use curl:")
			ios.Info("  curl http://<device-ip>/settings | jq '.actions'")
			return nil
		},
	}

	return cmd
}
