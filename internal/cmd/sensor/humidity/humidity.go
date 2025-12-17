// Package humidity provides humidity sensor commands.
package humidity

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewCommand creates the humidity command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.HumidityReading]{
		Name:    "humidity",
		Aliases: []string{"humid", "rh"},
		Short:   "Manage humidity sensors",
		Long: `Manage humidity sensors on Shelly devices.

Humidity sensors (DHT22, HTU21D, or similar) provide relative humidity readings.`,
		Example: `  # List humidity sensors
  shelly sensor humidity list living-room

  # Get humidity reading
  shelly sensor humidity status living-room`,
		Prefix:        "humidity:",
		StatusMethod:  "Humidity.GetStatus",
		DisplayList:   cmdutil.DisplayHumidityList,
		DisplayStatus: cmdutil.DisplayHumidityStatus,
	})
}
