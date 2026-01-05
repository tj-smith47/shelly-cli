// Package info provides the device info subcommand.
package info

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

// NewCommand creates the device info command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "info <device>",
		Aliases: []string{"details", "show"},
		Short:   "Show device information",
		Long: `Show detailed information about a device.

The device can be specified by its registered name or IP address.`,
		Example: `  # Show info for a registered device
  shelly device info living-room

  # Show info by IP address
  shelly device info 192.168.1.100

  # Output as JSON
  shelly device info living-room -o json

  # Output as YAML
  shelly device info living-room -o yaml

  # Short form
  shelly dev info office-switch`,
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
		cache.TypeDeviceInfo, cache.TTLDeviceInfo,
		"Getting device info...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.DeviceInfo, error) {
			// Use DeviceInfoAuto to support both Gen1 and Gen2 devices
			return svc.DeviceInfoAuto(ctx, device)
		},
		term.DisplayDeviceInfo)
}
