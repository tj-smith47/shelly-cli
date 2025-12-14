// Package disable provides the auth disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var yesFlag bool

// NewCommand creates the auth disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "remove"},
		Short:   "Disable authentication",
		Long: `Disable authentication for a device.

This removes the password requirement for accessing the device locally.
Use with caution in production environments.`,
		Example: `  # Disable authentication
  shelly auth disable living-room

  # Disable without confirmation prompt
  shelly auth disable living-room --yes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ios := f.IOStreams()

	// Confirm before disabling
	if !yesFlag {
		confirmed, err := ios.Confirm(
			fmt.Sprintf("Disable authentication on %s? This will allow unauthenticated access.", device),
			false,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Warning("Cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling authentication...", func(ctx context.Context) error {
		if err := svc.DisableAuth(ctx, device); err != nil {
			return fmt.Errorf("failed to disable authentication: %w", err)
		}
		ios.Success("Authentication disabled on %s", device)
		return nil
	})
}
