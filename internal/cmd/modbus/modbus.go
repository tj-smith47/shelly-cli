// Package modbus provides the modbus command group.
package modbus

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/modbus/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/modbus/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/modbus/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the modbus command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modbus",
		Aliases: []string{"mb"},
		Short:   "Manage Modbus-TCP configuration",
		Long: `Manage Modbus-TCP server on Shelly Gen2+ devices.

The Modbus-TCP server runs on port 502 when enabled, allowing integration
with industrial automation systems and SCADA software.

Device info registers (when enabled):
  30000: Device MAC (6 registers / 12 bytes)
  30006: Device model (10 registers / 20 bytes)
  30016: Device name (32 registers / 64 bytes)

Additional component-specific registers are documented per-component.`,
		Example: `  # Check Modbus status
  shelly modbus status kitchen

  # Enable Modbus
  shelly modbus enable kitchen

  # Disable Modbus
  shelly modbus disable kitchen`,
	}

	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(enable.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))

	return cmd
}
