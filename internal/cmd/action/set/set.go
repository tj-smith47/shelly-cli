// Package set provides the action set command.
package set

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the action set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set <device> <action> <url>",
		Aliases: []string{"add", "configure"},
		Short:   "Set an action URL for a Gen1 device",
		Long: `Set an action URL for a Gen1 Shelly device.

Gen1 devices support various action types:
  - btn_on_url, btn_off_url: Button toggle actions
  - out_on_url, out_off_url: Output state change actions
  - roller_open_url, roller_close_url, roller_stop_url: Roller actions
  - longpush_url, shortpush_url: Button press duration actions

Note: This feature is currently in development.

Workaround: Use curl to set action URLs directly:
  curl "http://<device-ip>/settings?<action>=<url>"`,
		Example: `  # Set output on action
  shelly action set living-room out_on_url "http://homeserver/api/light-on"

  # Workaround: use curl
  curl "http://192.168.1.100/settings?out_on_url=http%3A%2F%2Fhomeserver%2Fapi%2Flight-on"`,
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams()
			action := args[1]
			url := args[2]

			ios.Warning("Gen1 action URL management is not yet fully implemented.")
			ios.Println()
			ios.Info("To set the action URL, use curl:")
			ios.Info("  curl \"http://<device-ip>/settings?%s=%s\"", action, url)
			return nil
		},
	}

	return cmd
}
