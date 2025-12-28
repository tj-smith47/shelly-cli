// Package disable provides the auth disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the auth disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Confirm before disabling
	confirmed, err := opts.Factory.ConfirmAction(
		fmt.Sprintf("Disable authentication on %s? This will allow unauthenticated access.", opts.Device),
		opts.Yes,
	)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Cancelled")
		return nil
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling authentication...", func(ctx context.Context) error {
		if err := svc.DisableAuth(ctx, opts.Device); err != nil {
			return fmt.Errorf("failed to disable authentication: %w", err)
		}
		ios.Success("Authentication disabled on %s", opts.Device)
		return nil
	})
}
