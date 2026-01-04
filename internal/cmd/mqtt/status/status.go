// Package status provides the mqtt status subcommand.
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

// NewCommand creates the mqtt status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show MQTT status",
		Long: `Show the MQTT connection status for a device.

Displays whether MQTT is enabled and if the device is connected to the broker.`,
		Example: `  # Show MQTT status
  shelly mqtt status living-room

  # Output as JSON
  shelly mqtt status living-room -o json`,
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
		cache.TypeMQTT, cache.TTLProtocols,
		"Getting MQTT status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.MQTTStatus, error) {
			return svc.GetMQTTStatus(ctx, device)
		},
		term.DisplayMQTTStatus)
}
