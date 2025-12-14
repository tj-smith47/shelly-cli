// Package schedule provides thermostat schedule management commands.
package schedule

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the thermostat schedule command group.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schedule",
		Aliases: []string{"sched", "sc"},
		Short:   "Manage thermostat schedules",
		Long: `Manage time-based schedules for thermostat control.

Schedules allow automatic temperature adjustments at specific times.
You can set different target temperatures for different times of day
or days of the week.

Schedule timespec format (cron-like):
  "ss mm hh DD WW" - seconds, minutes, hours, day of month, weekday

Special values:
  @sunrise  - At sunrise (with optional offset like @sunrise+30)
  @sunset   - At sunset (with optional offset like @sunset-15)

Examples:
  "0 0 8 * 1-5"   - 8:00 AM on weekdays
  "0 30 22 * *"   - 10:30 PM every day
  "0 0 6 * 0,6"   - 6:00 AM on weekends`,
		Example: `  # List all thermostat schedules
  shelly thermostat schedule list gateway

  # Create a morning schedule (22°C at 7:00 AM on weekdays)
  shelly thermostat schedule create gateway --target 22 --time "0 0 7 * 1-5"

  # Create a night schedule (18°C at 10:00 PM every day)
  shelly thermostat schedule create gateway --target 18 --time "0 0 22 * *"

  # Delete a schedule
  shelly thermostat schedule delete gateway --id 1`,
	}

	cmd.AddCommand(newListCommand(f))
	cmd.AddCommand(newCreateCommand(f))
	cmd.AddCommand(newDeleteCommand(f))
	cmd.AddCommand(newEnableCommand(f))
	cmd.AddCommand(newDisableCommand(f))

	return cmd
}
