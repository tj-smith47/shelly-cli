// Package power provides commands for managing power meter components (PM/PM1).
package power

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/power/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/power/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the power command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "Power meter operations (PM/PM1 components)",
		Long: `Manage and monitor power meter components.

This command works with PM and PM1 power meter components found on
Shelly devices that include energy metering functionality:
  - Shelly Plus PM (single-phase power meter)
  - Shelly Pro series with PM components
  - Various Plus/Pro devices with built-in power metering

PM/PM1 components provide real-time measurements including:
  - Voltage, current, and power
  - Frequency
  - Accumulated energy (total and by-minute)
  - Return energy (for bidirectional meters)

For professional energy monitors (EM/EM1 components), use 'shelly energy'.`,
		Aliases: []string{"pm"},
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))

	return cmd
}
