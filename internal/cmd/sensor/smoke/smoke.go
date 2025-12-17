// Package smoke provides smoke sensor commands.
package smoke

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewCommand creates the smoke command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.AlarmSensorReading]{
		Name:    "smoke",
		Aliases: []string{"detector"},
		Short:   "Manage smoke sensors",
		Long: `Manage smoke detection sensors on Shelly devices.

Smoke sensors provide alarm state detection and the ability
to mute active alarms.`,
		Example: `  # List smoke sensors
  shelly sensor smoke list kitchen

  # Check smoke status
  shelly sensor smoke status kitchen

  # Test smoke alarm
  shelly sensor smoke test kitchen

  # Mute active alarm
  shelly sensor smoke mute kitchen`,
		Prefix:           "smoke:",
		StatusMethod:     "Smoke.GetStatus",
		AlarmSensorTitle: "Smoke",
		AlarmMessage:     "SMOKE DETECTED!",
		HasTest:          true,
		TestHint: `The Smoke component does not have a programmatic test method.
  To test the smoke detector, use the device's physical test
  button or use appropriate test spray (follow manufacturer
  guidelines).`,
		HasMute:    true,
		MuteMethod: "Smoke.Mute",
	})
}
