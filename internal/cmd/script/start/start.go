// Package start provides the script start subcommand.
package start

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// NewCommand creates the script start command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start <device> <id>",
		Aliases: []string{"run"},
		Short:   "Start a script",
		Long: `Start a script on a Gen2+ Shelly device.

The script must be enabled before it can be started.`,
		Example: `  # Start a script
  shelly script start living-room 1`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenScriptID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			return run(cmd.Context(), f, args[0], id)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Starting script...", func(ctx context.Context) error {
		if err := svc.StartScript(ctx, device, id); err != nil {
			return fmt.Errorf("failed to start script: %w", err)
		}
		ios.Success("Script %d started", id)
		return nil
	})
}
