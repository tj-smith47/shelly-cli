// Package rename provides the device rename subcommand.
package rename

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	OldName string
	NewName string
}

// NewCommand creates the device rename command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "rename <old-name> <new-name>",
		Aliases: []string{"mv", "move"},
		Short:   "Rename a device in the registry",
		Long: `Rename a Shelly device in the local registry.

This updates the device's friendly name and updates any group memberships
to use the new name. The device's address and configuration are preserved.`,
		Example: `  # Rename a device
  shelly device rename kitchen kitchen-light

  # Short form
  shelly dev mv bedroom master-bedroom`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.OldName = args[0]
			opts.NewName = args[1]
			return run(cmd.Context(), opts)
		},
	}

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if source device exists
	if _, exists := config.GetDevice(opts.OldName); !exists {
		return fmt.Errorf("device %q not found", opts.OldName)
	}

	if err := config.RenameDevice(opts.OldName, opts.NewName); err != nil {
		return fmt.Errorf("failed to rename device: %w", err)
	}

	ios.Success("Renamed %q to %q", opts.OldName, opts.NewName)
	return nil
}
