// Package devicepower provides device power (battery) sensor commands.
package devicepower

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the devicepower command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewSensorCommand(f, factories.SensorOpts[model.DevicePowerReading]{
		Name:    "devicepower",
		Aliases: []string{"power", "battery", "bat"},
		Short:   "Manage device power sensors",
		Long: `Manage device power sensors on Shelly devices.

Device power sensors provide battery and external power status for
battery-powered Shelly devices (Plus HT, H&T, etc.).`,
		Example: `  # List device power sensors
  shelly sensor devicepower list sensor1

  # Get battery status
  shelly sensor devicepower status sensor1`,
		Prefix:        "devicepower:",
		StatusMethod:  "DevicePower.GetStatus",
		DisplayList:   term.DisplayDevicePowerList,
		DisplayStatus: term.DisplayDevicePowerStatus,
	})
}
