// Package mqtt provides MQTT configuration commands.
package mqtt

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mqtt/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the mqtt command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mqtt",
		Short: "Manage device MQTT configuration",
		Long: `Manage MQTT configuration for devices.

Enable and configure MQTT for integration with home automation systems
like Home Assistant, OpenHAB, or custom MQTT brokers.`,
		Example: `  # Show MQTT status
  shelly mqtt status living-room

  # Configure MQTT broker
  shelly mqtt set living-room --server "mqtt://broker:1883" --user user --password pass

  # Disable MQTT
  shelly mqtt disable living-room`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))

	return cmd
}
