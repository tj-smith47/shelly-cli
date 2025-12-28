// Package sensoraddon provides the sensoraddon command group.
package sensoraddon

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/sensoraddon/add"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensoraddon/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensoraddon/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensoraddon/scan"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the sensoraddon command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sensoraddon",
		Aliases: []string{"addon", "sa"},
		Short:   "Manage Sensor Add-on peripherals",
		Long: `Manage Sensor Add-on peripherals on Shelly Gen2+ devices.

The Sensor Add-on board allows connecting external sensors:
  - DS18B20: Dallas 1-Wire temperature sensors
  - DHT22: Temperature and humidity sensor
  - Digital inputs
  - Analog inputs

Supported devices: Plus1, Plus1PM, Plus2PM, PlusI4, Plus10V, PlusRGBWPM,
Dimmer0110VPM G3, Shelly1G3, Shelly1PMG3, Shelly2PMG3, ShellyI4G3

Note: Peripheral changes require a device reboot to take effect.`,
		Example: `  # List configured peripherals
  shelly sensoraddon list kitchen

  # Scan for OneWire devices
  shelly sensoraddon scan kitchen

  # Add a DS18B20 sensor
  shelly sensoraddon add kitchen ds18b20 --addr "40:255:100:6:199:204:149:177"

  # Add a DHT22 sensor
  shelly sensoraddon add kitchen dht22

  # Remove a peripheral
  shelly sensoraddon remove kitchen temperature:100`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(add.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))
	cmd.AddCommand(scan.NewCommand(f))

	return cmd
}
