// Package status provides the wifi status subcommand.
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

// NewCommand creates the wifi status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show WiFi status",
		Long: `Show the current WiFi status for a device.

Displays connection status, IP address, SSID, signal strength (RSSI),
and number of clients connected to the access point (if enabled).`,
		Example: `  # Show WiFi status
  shelly wifi status living-room

  # Output as JSON
  shelly wifi status living-room -o json`,
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
		cache.TypeWiFi, cache.TTLWiFi,
		"Getting WiFi status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.WiFiStatus, error) {
			return svc.GetWiFiStatus(ctx, device)
		},
		term.DisplayWiFiStatus)
}
