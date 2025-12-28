// Package list provides the rgbw list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the rgbw list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewListCommand(f, factories.ListOpts[shelly.RGBWInfo]{
		Component: "RGBW",
		Long: `List all RGBW light components on the specified device with their current status.

RGBW components control color-capable lights with an additional white channel.
Each component has an ID, optional name, state (ON/OFF), RGB color values,
white channel value, brightness level, and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State (ON/OFF), Color (R:G:B), White, Brightness (%), Power`,
		Example: `  # List all RGBW components on a device
  shelly rgbw list living-room

  # List RGBW components with JSON output
  shelly rgbw list living-room -o json

  # Get RGBW lights that are currently ON
  shelly rgbw list living-room -o json | jq '.[] | select(.output == true)'

  # Get current color and white values
  shelly rgbw list living-room -o json | jq '.[] | {id, r: .rgb.r, g: .rgb.g, b: .rgb.b, white}'

  # Find lights with white channel active
  shelly rgbw list living-room -o json | jq '.[] | select(.white > 0)'

  # Short forms
  shelly rgbw ls living-room`,
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.RGBWInfo, error) {
			return svc.RGBWList(ctx, device)
		},
		Display: term.DisplayRGBWList,
	})
}
