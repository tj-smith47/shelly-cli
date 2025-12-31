// Package list provides the script list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the script list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.AutomationService()

	items, err := cmdutil.RunWithSpinnerResult(ctx, ios, "Getting scripts...", func(ctx context.Context) ([]automation.ScriptInfo, error) {
		return svc.ListScripts(ctx, opts.Device)
	})
	if err != nil {
		return err
	}

	if len(items) == 0 {
		ios.NoResults("scripts")
		return nil
	}

	return cmdutil.PrintListResult(ios, items, term.DisplayScriptList)
}
