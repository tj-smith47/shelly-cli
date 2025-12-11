// Package cloud provides cloud configuration commands.
package cloud

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/status"
)

// NewCommand creates the cloud command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage device cloud connection",
		Long: `Manage the Shelly Cloud connection for devices.

Enable or disable the cloud connection for remote access and monitoring
through the Shelly Cloud service.`,
		Example: `  # Show cloud connection status
  shelly cloud status living-room

  # Enable cloud connection
  shelly cloud enable living-room

  # Disable cloud connection
  shelly cloud disable living-room`,
	}

	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(enable.NewCommand())
	cmd.AddCommand(disable.NewCommand())

	return cmd
}
