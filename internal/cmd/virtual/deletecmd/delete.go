// Package deletecmd provides the virtual delete command.
package deletecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Device  string
	Key     string
	Factory *cmdutil.Factory
}

// NewCommand creates the virtual delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete <device> <key>",
		Aliases: []string{"del", "rm", "remove"},
		Short:   "Delete a virtual component",
		Long: `Delete a virtual component from a Shelly Gen2+ device.

The key format is "type:id", for example "boolean:200" or "number:201".

This action cannot be undone.`,
		Example: `  # Delete a virtual component
  shelly virtual delete kitchen boolean:200

  # Skip confirmation
  shelly virtual delete kitchen boolean:200 --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddConfirmFlags(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate key format
	if _, _, err := shelly.ParseVirtualKey(opts.Key); err != nil {
		return err
	}

	// Confirm deletion
	if !opts.Yes {
		confirmed, err := ios.Confirm(fmt.Sprintf("Delete virtual component %s?", opts.Key), false)
		if err != nil {
			return err
		}
		if !confirmed {
			ios.Info("Cancelled")
			return nil
		}
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Deleting virtual component...", func(ctx context.Context) error {
		return svc.DeleteVirtualComponent(ctx, opts.Device, opts.Key)
	})
	if err != nil {
		return err
	}

	ios.Success("Deleted virtual component %s", opts.Key)

	// Invalidate cached virtual component list
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeVirtuals)
	return nil
}
