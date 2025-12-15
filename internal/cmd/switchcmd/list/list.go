// Package list provides the switch list subcommand.
package list

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the switch list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List switch components",
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
		"Fetching switch components...",
		"switch components",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.SwitchInfo, error) {
			return svc.SwitchList(ctx, device)
		},
		displayList)
}

func displayList(ios *iostreams.IOStreams, switches []shelly.SwitchInfo) {
	t := output.NewTable("ID", "Name", "State", "Power")
	for _, sw := range switches {
		name := sw.Name
		if name == "" {
			name = fmt.Sprintf("switch:%d", sw.ID)
		}

		state := theme.StatusError().Render("OFF")
		if sw.Output {
			state = theme.StatusOK().Render("ON")
		}

		power := "-"
		if sw.Power > 0 {
			power = fmt.Sprintf("%.1f W", sw.Power)
		}

		t.AddRow(fmt.Sprintf("%d", sw.ID), name, state, power)
	}
	t.Print()
}
