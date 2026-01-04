// Package status provides the cloud status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
}

// NewCommand creates the cloud status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show cloud connection status",
		Long: `Show the Shelly Cloud connection status for a device.

Displays whether the device is currently connected to Shelly Cloud.`,
		Example: `  # Show cloud status
  shelly cloud status living-room

  # Output as JSON
  shelly cloud status living-room -o json`,
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

	return cmdutil.RunCachedDeviceStatus(ctx, opts.Factory, opts.Device,
		cache.TypeCloud, cache.TTLCloud,
		"Getting cloud status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.CloudStatus, error) {
			return svc.GetCloudStatus(ctx, device)
		},
		term.DisplayCloudConnectionStatus)
}
