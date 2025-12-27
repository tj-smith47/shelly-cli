// Package list provides the cover list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/factories"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the cover list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	return factories.NewListCommand(f, factories.ListOpts[shelly.CoverInfo]{
		Component: "Cover",
		Long: `List all cover/roller components on the specified device with their current status.

Cover components control motorized blinds, shutters, and garage doors. Each
cover has an ID, optional name, state (open/closed/opening/closing/stopped),
position (percentage), and power consumption if supported.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, State, Position (%), Power (watts)`,
		Example: `  # List all covers on a device
  shelly cover list bedroom

  # List covers with JSON output
  shelly cover list bedroom -o json

  # Get covers that are fully open
  shelly cover list bedroom -o json | jq '.[] | select(.current_pos == 100)'

  # Find covers currently in motion
  shelly cover list bedroom -o json | jq '.[] | select(.state == "opening" or .state == "closing")'

  # Get position of all covers
  shelly cover list bedroom -o json | jq '.[] | {id, position: .current_pos}'

  # Check cover positions across multiple devices
  for dev in bedroom living-room; do
    echo "=== $dev ==="
    shelly cover list "$dev" --no-color
  done

  # Short forms
  shelly cover ls bedroom
  shelly cv ls bedroom`,
		Fetcher: func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.CoverInfo, error) {
			return svc.CoverList(ctx, device)
		},
		Display: term.DisplayCoverList,
	})
}
