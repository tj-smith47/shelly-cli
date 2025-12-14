// Package clearcmd provides the action clear command.
package clearcmd

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the action clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear <device> <action>",
		Aliases: []string{"delete", "remove", "rm"},
		Short:   "Clear an action URL for a Gen1 device",
		Long: `Clear (remove) an action URL for a Gen1 Shelly device.

This removes the configured URL for the specified action, disabling the
HTTP callback for that event.

Note: This feature is currently in development.

Workaround: Use curl to clear action URLs directly:
  curl "http://<device-ip>/settings?<action>="`,
		Example: `  # Clear output on action
  shelly action clear living-room out_on_url

  # Workaround: use curl
  curl "http://192.168.1.100/settings?out_on_url="`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams()
			action := args[1]

			ios.Warning("Gen1 action URL management is not yet fully implemented.")
			ios.Println()
			ios.Info("To clear the action URL, use curl:")
			ios.Info("  curl \"http://<device-ip>/settings?%s=\"", action)
			return nil
		},
	}

	return cmd
}
