// Package firmware provides firmware management commands.
package firmware

import (
	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"

	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware/check"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware/download"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware/rollback"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/firmware/update"
)

// NewCommand creates the firmware command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "firmware",
		Aliases: []string{"fw"},
		Short:   "Manage device firmware",
		Long: `Manage firmware for Shelly devices.

Check for updates, update to the latest version, or rollback to a previous version.`,
		Example: `  # Check for updates on a device
  shelly firmware check living-room

  # Check for updates on all registered devices
  shelly firmware check --all

  # Show firmware status
  shelly firmware status living-room

  # Update device firmware
  shelly firmware update living-room

  # Update to beta firmware
  shelly firmware update living-room --beta

  # Rollback to previous firmware
  shelly firmware rollback living-room

  # Download firmware file
  shelly firmware download ShellyPlus1PM 1.0.0 --output firmware.zip`,
	}

	cmd.AddCommand(check.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(update.NewCommand(f))
	cmd.AddCommand(rollback.NewCommand(f))
	cmd.AddCommand(download.NewCommand(f))

	return cmd
}
