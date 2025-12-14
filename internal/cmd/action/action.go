// Package action provides Gen1 action URL management commands.
package action

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/action/clearcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/action/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/action/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/action/test"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the action command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "action",
		Aliases: []string{"actions"},
		Short:   "Manage Gen1 device action URLs",
		Long: `Manage action URLs for Gen1 Shelly devices.

Gen1 devices support action URLs that trigger HTTP requests when specific events
occur (e.g., button press, output state change). This is different from Gen2
webhooks which have a more structured API.

Common Gen1 actions include:
  - btn_on_url, btn_off_url: Button toggle actions
  - out_on_url, out_off_url: Output state change actions
  - roller_open_url, roller_close_url: Roller actions
  - longpush_url, shortpush_url: Button press duration actions

Note: For Gen2 devices, use 'shelly webhook' instead.`,
		Example: `  # List all action URLs for a device
  shelly action list living-room

  # Set an action URL
  shelly action set living-room out_on_url "http://homeserver/api/light-on"

  # Clear an action URL
  shelly action clear living-room out_on_url

  # Test an action
  shelly action test living-room out_on_url`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(clearcmd.NewCommand(f))
	cmd.AddCommand(test.NewCommand(f))

	return cmd
}
