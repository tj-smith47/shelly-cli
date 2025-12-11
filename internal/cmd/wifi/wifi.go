// Package wifi provides WiFi configuration commands.
package wifi

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi/ap"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi/scan"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/wifi/status"
)

// NewCommand creates the wifi command and its subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wifi",
		Short: "Manage device WiFi configuration",
		Long: `Manage device WiFi configuration settings.

Get WiFi status, scan for networks, and configure WiFi settings including
station mode (connecting to a network) and access point mode.`,
		Example: `  # Show WiFi status
  shelly wifi status living-room

  # Scan for available networks
  shelly wifi scan living-room

  # Configure WiFi connection
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Configure access point
  shelly wifi ap living-room --enable --ssid "ShellyAP"`,
	}

	cmd.AddCommand(status.NewCommand())
	cmd.AddCommand(scan.NewCommand())
	cmd.AddCommand(set.NewCommand())
	cmd.AddCommand(ap.NewCommand())

	return cmd
}
