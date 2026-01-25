// Package remove provides the device remove subcommand.
package remove

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
	Name    string
	Force   bool
}

// NewCommand creates the device remove command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete", "unregister"},
		Short:   "Remove a device from the registry",
		Long: `Remove a Shelly device from the local registry.

The device will also be removed from any groups it belongs to.
This does not affect the physical device itself.`,
		Example: `  # Remove a device
  shelly device remove kitchen

  # Remove with force (skip confirmation)
  shelly device remove kitchen --force

  # Short form
  shelly dev rm bedroom`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Check if device exists
	dev, exists := config.GetDevice(opts.Name)
	if !exists {
		return fmt.Errorf("device %q not found", opts.Name)
	}

	// Confirm unless forced
	if !opts.Force {
		displayName := dev.Name
		if displayName == "" {
			displayName = opts.Name
		}
		confirmed, err := ios.Confirm(fmt.Sprintf("Remove device %q?", displayName), false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	if err := config.UnregisterDevice(opts.Name); err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}

	ios.Success("Removed device %q", opts.Name)
	return nil
}
