// Package flood provides flood sensor commands.
package flood

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewCommand creates the flood command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.AlarmSensorReading]{
		Name:    "flood",
		Aliases: []string{"water", "leak"},
		Short:   "Manage flood sensors",
		Long: `Manage flood (water leak) sensors on Shelly devices.

Flood sensors detect water leaks and can trigger alarms with
different modes: disabled, normal, intense, or rain detection.`,
		Example: `  # List flood sensors
  shelly sensor flood list bathroom

  # Check flood status
  shelly sensor flood status bathroom

  # Test flood alarm
  shelly sensor flood test bathroom`,
		Prefix:           "flood:",
		StatusMethod:     "Flood.GetStatus",
		AlarmSensorTitle: "Flood",
		AlarmMessage:     "WATER DETECTED!",
		HasTest:          true,
		TestHint: `The Flood component does not have a programmatic test method.
  To test the flood sensor, briefly apply water to the sensor
  contacts or use the device's physical test button if available.`,
	})
}
