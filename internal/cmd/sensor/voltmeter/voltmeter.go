// Package voltmeter provides voltmeter sensor commands.
package voltmeter

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewCommand creates the voltmeter command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.VoltmeterReading]{
		Name:    "voltmeter",
		Aliases: []string{"volt", "voltage", "v"},
		Short:   "Manage voltmeter sensors",
		Long: `Manage voltmeter sensors on Shelly devices.

Voltmeter sensors provide voltage readings, useful for monitoring
power supplies, batteries, or other voltage sources.`,
		Example: `  # List voltmeters
  shelly sensor voltmeter list device1

  # Get voltage reading
  shelly sensor voltmeter status device1`,
		Prefix:        "voltmeter:",
		StatusMethod:  "Voltmeter.GetStatus",
		DisplayList:   cmdutil.DisplayVoltmeterList,
		DisplayStatus: cmdutil.DisplayVoltmeterStatus,
	})
}
