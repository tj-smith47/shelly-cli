// Package provision provides device provisioning commands.
package provision

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/bulk"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/wifi"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the provision command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "provision",
		Aliases: []string{"prov", "setup"},
		Short:   "Provision device settings",
		Long: `Provision device settings including WiFi, network, and bulk configuration.

The provision commands provide an interactive workflow for setting up devices,
including WiFi network scanning and selection, bulk provisioning from config files,
and BLE-based provisioning for devices in AP mode.`,
		Example: `  # Interactive WiFi provisioning
  shelly provision wifi living-room

  # Bulk provision from config file
  shelly provision bulk devices.yaml

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1`,
	}

	cmd.AddCommand(wifi.NewCommand(f))
	cmd.AddCommand(bulk.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))

	return cmd
}
