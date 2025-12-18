// Package list provides the input list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the input list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls"},
		Short:   "List input components",
		Long: `List all input components on a Shelly device with their current state.

Input components represent physical inputs (buttons, switches, sensors)
connected to the device. Each input has an ID, optional name, type
(button, switch, etc.), and current state (active/inactive).

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, Type, State (active/inactive)`,
		Example: `  # List all inputs on a device
  shelly input list living-room

  # Output as JSON for scripting
  shelly input list living-room -o json

  # Get active inputs only
  shelly input list living-room -o json | jq '.[] | select(.state == true)'

  # List inputs by type
  shelly input list living-room -o json | jq '.[] | select(.type == "button")'

  # Get input IDs only
  shelly input list living-room -o json | jq -r '.[].id'

  # Monitor input state across multiple devices
  for dev in switch-1 switch-2; do
    echo "=== $dev ==="
    shelly input list "$dev" --no-color
  done

  # Short form
  shelly input ls living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Getting inputs...",
		"inputs",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.InputInfo, error) {
			return svc.InputList(ctx, device)
		},
		cmdutil.DisplayInputList)
}
