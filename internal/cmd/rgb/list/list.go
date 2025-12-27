// Package list provides the rgb list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the rgb list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewListCommand(f, factories.ListOpts[shelly.RGBInfo]{
		Component: "RGB",
		Long: `List all RGB light components on the specified device with their current status.

RGB components control color-capable lights (RGBW, RGBW2, etc.). Each
component has an ID, optional name, state (ON/OFF), RGB color values,
brightness level, and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Color (R:G:B), Brightness (%), Power`,
		Example: `  # List all RGB components on a device
  shelly rgb list living-room

  # List RGB components with JSON output
  shelly rgb list living-room -o json

  # Get RGB lights that are currently ON
  shelly rgb list living-room -o json | jq '.[] | select(.output == true)'

  # Get current color values
  shelly rgb list living-room -o json | jq '.[] | {id, r: .rgb.r, g: .rgb.g, b: .rgb.b}'

  # Find lights set to pure red
  shelly rgb list living-room -o json | jq '.[] | select(.rgb.r == 255 and .rgb.g == 0 and .rgb.b == 0)'

  # Get brightness levels
  shelly rgb list living-room -o json | jq '.[] | {id, brightness}'

  # Short forms
  shelly rgb ls living-room`,
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.RGBInfo, error) {
			return svc.RGBList(ctx, device)
		},
		Display: term.DisplayRGBList,
	})
}
