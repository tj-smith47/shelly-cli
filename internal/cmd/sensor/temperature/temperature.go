// Package temperature provides temperature sensor commands.
package temperature

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewCommand creates the temperature command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.TemperatureReading]{
		Name:    "temperature",
		Aliases: []string{"temp", "t"},
		Short:   "Manage temperature sensors",
		Long: `Manage temperature sensors on Shelly devices.

Temperature sensors can be built-in or external (DS18B20).
Readings are provided in both Celsius and Fahrenheit.`,
		Example: `  # List temperature sensors
  shelly sensor temperature list living-room

  # Get temperature reading
  shelly sensor temperature status living-room`,
		Prefix:        "temperature:",
		StatusMethod:  "Temperature.GetStatus",
		DisplayList:   cmdutil.DisplayTemperatureList,
		DisplayStatus: cmdutil.DisplayTemperatureStatus,
	})
}
