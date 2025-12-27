// Package list provides the switch list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the switch list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewListCommand(f, factories.ListOpts[shelly.SwitchInfo]{
		Component: "Switch",
		Long: `List all switch components on the specified device with their current status.

Switch components control relay outputs (on/off). Each switch has an ID,
optional name, current state (ON/OFF), and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Power (watts)`,
		Example: `  # List all switches on a device
  shelly switch list kitchen

  # List switches with JSON output
  shelly switch list kitchen -o json

  # Get switches that are currently ON
  shelly switch list kitchen -o json | jq '.[] | select(.output == true)'

  # Calculate total power consumption
  shelly switch list kitchen -o json | jq '[.[].power] | add'

  # Get switch IDs only
  shelly switch list kitchen -o json | jq -r '.[].id'

  # Check all switches across multiple devices
  for dev in kitchen bedroom living-room; do
    echo "=== $dev ==="
    shelly switch list "$dev" --no-color
  done

  # Short forms
  shelly switch ls kitchen
  shelly sw ls kitchen`,
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.SwitchInfo, error) {
			return svc.SwitchList(ctx, device)
		},
		Display: term.DisplaySwitchList,
	})
}
