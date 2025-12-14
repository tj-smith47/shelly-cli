// Package ota provides Gen1 firmware update commands.
package ota

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the gen1 ota command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ota",
		Aliases: []string{"firmware", "fw", "update"},
		Short:   "Manage Gen1 firmware updates",
		Long: `Manage firmware updates on Gen1 Shelly devices.

Gen1 devices use HTTP-based OTA updates. The firmware can be
updated from the official Shelly servers or from a custom URL.

For Gen2+ devices, use 'shelly firmware' instead.`,
		Example: `  # Check for firmware updates
  shelly gen1 ota check living-room

  # Update firmware
  shelly gen1 ota update living-room

  # Update to specific firmware URL
  shelly gen1 ota update living-room --url http://archive.shelly-tools.de/...`,
	}

	cmd.AddCommand(newCheckCommand(f))
	cmd.AddCommand(newUpdateCommand(f))

	return cmd
}
