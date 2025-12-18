// Package list provides the light list subcommand.
package list

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the light list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List light components",
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
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Fetching light components...",
		"light components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.LightInfo, error) {
			return svc.LightList(ctx, device)
		},
		cmdutil.DisplayLightList)
}
