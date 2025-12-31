// Package disable provides the cloud disable subcommand.
package disable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the cloud disable command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "disable <device>",
		Aliases: []string{"off", "disconnect"},
		Short:   "Disable cloud connection",
		Long: `Disable the Shelly Cloud connection for a device.

Once disabled, the device will only be accessible via local network.`,
		Example: `  # Disable cloud connection
  shelly cloud disable living-room`,
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

	return cmdutil.RunWithSpinner(ctx, ios, "Disabling cloud connection...", func(ctx context.Context) error {
		if err := svc.SetCloudEnabled(ctx, opts.Device, false); err != nil {
			return fmt.Errorf("failed to disable cloud: %w", err)
		}
		ios.Success("Cloud connection disabled on %s", opts.Device)
		return nil
	})
}
