// Package cloud provides cloud configuration and API commands.
package cloud

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/authstatus"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/control"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/device"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/devices"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/events"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/login"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/logout"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cloud/token"
)

// NewCommand creates the cloud command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Manage cloud connection and Shelly Cloud API",
		Long: `Manage device cloud connection and interact with the Shelly Cloud API.

Device cloud commands:
  status    Show device cloud connection status
  enable    Enable cloud connection on a device
  disable   Disable cloud connection on a device

Cloud API commands (requires login):
  login       Authenticate with Shelly Cloud
  logout      Clear cloud credentials
  auth-status Show authentication status
  token       Show/manage access token
  devices     List cloud-registered devices
  device      Show cloud device details
  control     Control devices via cloud
  events      Subscribe to real-time cloud events`,
		Example: `  # Device cloud configuration
  shelly cloud status living-room
  shelly cloud enable living-room
  shelly cloud disable living-room

  # Cloud API authentication
  shelly cloud login
  shelly cloud auth-status
  shelly cloud logout

  # Cloud API device management
  shelly cloud devices
  shelly cloud device abc123
  shelly cloud control abc123 on`,
	}

	// Device cloud configuration commands
	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(enable.NewCommand())
	cmd.AddCommand(disable.NewCommand())

	// Cloud API authentication commands
	cmd.AddCommand(login.NewCommand())
	cmd.AddCommand(logout.NewCommand())
	cmd.AddCommand(authstatus.NewCommand())
	cmd.AddCommand(token.NewCommand())

	// Cloud API device management commands
	cmd.AddCommand(devices.NewCommand())
	cmd.AddCommand(device.NewCommand())
	cmd.AddCommand(control.NewCommand())
	cmd.AddCommand(events.NewCommand())

	return cmd
}
