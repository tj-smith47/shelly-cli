// Package del provides the script delete subcommand.
package del

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var yesFlag bool

// NewCommand creates the script delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <device> <id>",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a script",
		Long: `Delete a script from a Gen2+ Shelly device.

This permanently removes the script and its code from the device.`,
		Example: `  # Delete a script
  shelly script delete living-room 1

  # Delete without confirmation
  shelly script delete living-room 1 --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			return run(cmd.Context(), f, args[0], id)
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Confirm unless --yes
	if !yesFlag {
		ios.Warning("This will permanently delete script %d.", id)
		confirmed, err := ios.Confirm("Delete script?", false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Warning("Delete cancelled")
			return nil
		}
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Deleting script...", func(ctx context.Context) error {
		if err := svc.DeleteScript(ctx, device, id); err != nil {
			return fmt.Errorf("failed to delete script: %w", err)
		}
		ios.Success("Script %d deleted", id)
		return nil
	})
}
