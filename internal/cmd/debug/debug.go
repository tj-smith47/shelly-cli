// Package debug provides debug and diagnostic commands.
package debug

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/debug/coiot"
	"github.com/tj-smith47/shelly-cli/internal/cmd/debug/log"
	"github.com/tj-smith47/shelly-cli/internal/cmd/debug/websocket"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the debug command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug",
		Aliases: []string{"dbg", "diag"},
		Short:   "Debug and diagnostic commands",
		Long: `Debug and diagnostic commands for troubleshooting Shelly devices.

These commands provide low-level access to device communication protocols
and diagnostic information. Use them for debugging issues or exploring
device capabilities.

For direct API calls, use 'shelly api' instead.

WARNING: Some debug commands may affect device behavior. Use with caution.`,
		Example: `  # Get Gen1 device debug log
  shelly debug log living-room-gen1

  # Show CoIoT status
  shelly debug coiot living-room

  # Debug WebSocket connection
  shelly debug websocket living-room`,
	}

	cmd.AddCommand(log.NewCommand(f))
	cmd.AddCommand(coiot.NewCommand(f))
	cmd.AddCommand(websocket.NewCommand(f))

	return cmd
}
