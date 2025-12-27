// Package list provides the light list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the light list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewListCommand(f, factories.ListOpts[shelly.LightInfo]{
		Component: "Light",
		Long: `List all light components on the specified device with their current status.

Light components control dimmable lights. Each light has an ID, optional
name, state (ON/OFF), brightness level (percentage), and power consumption
if supported. Some lights also support color temperature or RGB values.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Brightness (%), Power (watts)`,
		Example: `  # List all lights on a device
  shelly light list kitchen

  # List lights with JSON output
  shelly light list kitchen -o json

  # Get lights that are currently ON
  shelly light list kitchen -o json | jq '.[] | select(.output == true)'

  # Find lights below 50% brightness
  shelly light list kitchen -o json | jq '.[] | select(.brightness < 50)'

  # Calculate total light power consumption
  shelly light list kitchen -o json | jq '[.[].apower // 0] | add'

  # Get all light IDs
  shelly light list kitchen -o json | jq -r '.[].id'

  # Check lights across multiple devices
  for dev in kitchen bedroom living-room; do
    echo "=== $dev ==="
    shelly light list "$dev" --no-color
  done

  # Short forms
  shelly light ls kitchen
  shelly lt ls kitchen`,
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.LightInfo, error) {
			return svc.LightList(ctx, device)
		},
		Display: term.DisplayLightList,
	})
}
