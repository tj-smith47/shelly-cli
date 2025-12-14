// Package stop provides the script stop subcommand.
package stop

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the script stop command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <device> <id>",
		Short: "Stop a running script",
		Long:  `Stop a running script on a Gen2+ Shelly device.`,
		Example: `  # Stop a script
  shelly script stop living-room 1`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: cmdutil.CompleteDeviceThenScriptID(),
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
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Stopping script...", func(ctx context.Context) error {
		if err := svc.StopScript(ctx, device, id); err != nil {
			return fmt.Errorf("failed to stop script: %w", err)
		}
		ios.Success("Script %d stopped", id)
		return nil
	})
}
