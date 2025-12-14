// Package lora provides LoRa management commands.
package lora

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/lora/config"
	"github.com/tj-smith47/shelly-cli/internal/cmd/lora/send"
	"github.com/tj-smith47/shelly-cli/internal/cmd/lora/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the lora command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lora",
		Aliases: []string{"lorawan"},
		Short:   "Manage LoRa add-on",
		Long: `Manage LoRa add-on connectivity on Shelly devices.

LoRa (Long Range) is a wireless modulation technique that enables
long-distance communication with low power consumption. The Shelly
LoRa add-on extends device connectivity for scenarios where WiFi
is not available or practical.

LoRa features:
- Long range (up to 15km in ideal conditions)
- Low power consumption
- Point-to-point or star network topology
- Configurable frequency, bandwidth, and spreading factor`,
		Example: `  # Show LoRa add-on status
  shelly lora status living-room

  # Configure LoRa settings
  shelly lora config living-room --freq 868000000 --power 14

  # Send a message
  shelly lora send living-room "Hello World"`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(config.NewCommand(f))
	cmd.AddCommand(send.NewCommand(f))

	return cmd
}
