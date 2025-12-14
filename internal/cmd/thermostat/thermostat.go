// Package thermostat provides thermostat control commands.
package thermostat

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/boost"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/calibrate"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/disable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/enable"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/override"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/schedule"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/set"
	"github.com/tj-smith47/shelly-cli/internal/cmd/thermostat/status"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the thermostat command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "thermostat",
		Aliases: []string{"thermo", "trv", "hvac"},
		Short:   "Manage thermostats",
		Long: `Manage Shelly thermostat devices and components.

Thermostat support is available on specific devices like:
- Shelly BLU TRV (Thermostatic Radiator Valve) via BLU Gateway
- Devices with virtual thermostat components

Commands allow you to:
- View thermostat status (current/target temperature, valve position)
- Set target temperature and operating mode (heat/cool/auto)
- Enable/disable thermostat control
- Activate boost mode for rapid heating
- Override target temperature temporarily
- Calibrate valve position`,
		Example: `  # List thermostats on a device
  shelly thermostat list gateway

  # Show thermostat status
  shelly thermostat status gateway

  # Set target temperature to 22°C
  shelly thermostat set gateway --target 22

  # Enable thermostat in heat mode
  shelly thermostat enable gateway --mode heat

  # Activate boost for 5 minutes
  shelly thermostat boost gateway --duration 5m

  # Override temperature for 30 minutes
  shelly thermostat override gateway --target 25 --duration 30m`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(enable.NewCommand(f))
	cmd.AddCommand(disable.NewCommand(f))
	cmd.AddCommand(boost.NewCommand(f))
	cmd.AddCommand(override.NewCommand(f))
	cmd.AddCommand(calibrate.NewCommand(f))
	cmd.AddCommand(schedule.NewCommand(f))

	return cmd
}
