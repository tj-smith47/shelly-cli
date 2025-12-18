// Package list provides the script list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the script list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls"},
		Short:   "List scripts on a device",
		Long: `List all scripts on a Gen2+ Shelly device.

Shows script ID, name, enabled status, and running status. Scripts are
user-programmable JavaScript code that runs directly on the device for
automation and custom logic.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Name, Enabled (yes/no), Running (running/stopped)`,
		Example: `  # List all scripts
  shelly script list living-room

  # Output as JSON
  shelly script list living-room -o json

  # Get IDs of running scripts
  shelly script list living-room -o json | jq -r '.[] | select(.running) | .id'

  # Check if any script is running
  shelly script list living-room -o json | jq -e 'any(.running)' > /dev/null && echo "Scripts running"

  # Short form
  shelly script ls living-room`,
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
		"Getting scripts...",
		"scripts",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.ScriptInfo, error) {
			return svc.ListScripts(ctx, device)
		},
		cmdutil.DisplayScriptList)
}
