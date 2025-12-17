// Package sensor provides sensor management commands.
package sensor

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/devicepower"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/flood"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/humidity"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/illuminance"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/smoke"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/status"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/temperature"
	"github.com/tj-smith47/shelly-cli/internal/cmd/sensor/voltmeter"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the sensor command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sensor",
		Aliases: []string{"sensors", "env"},
		Short:   "Manage device sensors",
		Long: `Manage environmental sensors on Shelly devices.

Supports reading from various sensor types available on Gen2+ devices:
- Device power (battery status)
- Flood sensors (water leak detection)
- Humidity sensors (DHT22, HTU21D)
- Illuminance sensors (light level)
- Smoke sensors (smoke detection with alarm)
- Temperature sensors (built-in or external DS18B20)
- Voltmeters (voltage measurement)

Use the status command to get a combined view of all sensors on a device,
or use specific subcommands for individual sensor types.`,
		Example: `  # Show all sensor readings
  shelly sensor status living-room

  # Get temperature reading
  shelly sensor temperature status living-room

  # Check flood sensor
  shelly sensor flood status bathroom

  # Mute smoke alarm
  shelly sensor smoke mute kitchen`,
	}

	cmd.AddCommand(devicepower.NewCommand(f))
	cmd.AddCommand(flood.NewCommand(f))
	cmd.AddCommand(humidity.NewCommand(f))
	cmd.AddCommand(illuminance.NewCommand(f))
	cmd.AddCommand(smoke.NewCommand(f))
	cmd.AddCommand(status.NewCommand(f))
	cmd.AddCommand(temperature.NewCommand(f))
	cmd.AddCommand(voltmeter.NewCommand(f))

	return cmd
}
