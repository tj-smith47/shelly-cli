// Package ble provides BLE-based device provisioning.
package ble

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the provision ble command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ble <device-address>",
		Aliases: []string{"bluetooth"},
		Short:   "BLE-based provisioning (not yet implemented)",
		Long: `Provision a device using Bluetooth Low Energy (BLE).

This command is intended for provisioning devices in AP mode that are not yet
connected to your network. BLE provisioning requires Bluetooth hardware support.

NOTE: This feature is not yet implemented. For now, connect to the device's AP
network and use 'shelly provision wifi' with the device's AP IP address (usually
192.168.33.1).`,
		Example: `  # BLE provisioning (future)
  shelly provision ble ShellyPlus1-ABCD1234

  # Current workaround: connect to device AP and use wifi provisioning
  # 1. Connect your computer to "ShellyPlus1-ABCD1234" WiFi network
  # 2. Run: shelly provision wifi 192.168.33.1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ios := f.IOStreams()
			ios.Warning("BLE provisioning is not yet implemented.")
			ios.Println()
			ios.Info("Workaround for provisioning devices in AP mode:")
			ios.Info("  1. Connect your computer to the device's WiFi AP network")
			ios.Info("  2. Run: shelly provision wifi 192.168.33.1")
			ios.Println()
			ios.Info("The device's AP network name is usually shown on the device")
			ios.Info("or in the format 'ShellyPlus<Model>-<ID>'")
			return nil
		},
	}

	return cmd
}
