// Package illuminance provides illuminance sensor commands.
package illuminance

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the illuminance command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.IlluminanceReading]{
		Name:    "illuminance",
		Aliases: []string{"lux", "light-level", "brightness"},
		Short:   "Manage illuminance sensors",
		Long: `Manage illuminance (light level) sensors on Shelly devices.

Illuminance sensors provide light level readings in lux,
useful for automation based on ambient light conditions.`,
		Example: `  # List illuminance sensors
  shelly sensor illuminance list living-room

  # Get current light level
  shelly sensor illuminance status living-room`,
		Prefix:        "illuminance:",
		StatusMethod:  "Illuminance.GetStatus",
		DisplayList:   term.DisplayIlluminanceList,
		DisplayStatus: term.DisplayIlluminanceStatus,
	})
}
