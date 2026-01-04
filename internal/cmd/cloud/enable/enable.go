// Package enable provides the cloud enable subcommand.
package enable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the cloud enable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "enable <device>",
		Aliases: []string{"on", "connect"},
		Short:   "Enable cloud connection",
		Long: `Enable the Shelly Cloud connection for a device.

Once enabled, the device will connect to Shelly Cloud for remote access
and monitoring.`,
		Example: `  # Enable cloud connection
  shelly cloud enable living-room`,
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
	svc := opts.Factory.ShellyService()

	err := cmdutil.RunWithSpinner(ctx, ios, "Enabling cloud connection...", func(ctx context.Context) error {
		if setErr := svc.SetCloudEnabled(ctx, opts.Device, true); setErr != nil {
			return fmt.Errorf("failed to enable cloud: %w", setErr)
		}
		ios.Success("Cloud connection enabled on %s", opts.Device)
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cached cloud status
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeCloud)
	return nil
}
