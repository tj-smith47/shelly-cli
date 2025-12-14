// Package energy provides commands for managing professional energy monitoring components (EM/EM1).
package energy

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/compare"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/dashboard"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/history"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/reset"
	"github.com/tj-smith47/shelly-cli/internal/cmd/energy/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the energy command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "energy",
		Short: "Energy monitoring operations (EM/EM1 components)",
		Long: `Manage and monitor professional energy monitoring components.

This command works with EM (3-phase) and EM1 (single-phase) energy monitor
components found on professional Shelly devices like:
  - Shelly Pro 3EM (3-phase energy monitor)
  - Shelly Pro EM (single-phase or dual-phase monitor)
  - Shelly Pro EM-50 (professional energy monitor)

These components provide real-time measurements including:
  - Voltage, current, power (active/apparent)
  - Power factor and frequency
  - Per-phase data for 3-phase monitors
  - Total power and neutral current

For power meters with energy totals (PM/PM1 components), use 'shelly power'.`,
		Aliases: []string{"em"},
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(history.NewCommand(f))
	cmd.AddCommand(export.NewCommand(f))
	cmd.AddCommand(reset.NewCommand(f))
	cmd.AddCommand(dashboard.NewCommand(f))
	cmd.AddCommand(compare.NewCommand(f))

	return cmd
}
