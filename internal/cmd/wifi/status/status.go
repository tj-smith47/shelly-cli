// Package status provides the wifi status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the wifi status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting WiFi status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.WiFiStatus, error) {
			return svc.GetWiFiStatus(ctx, device)
		},
		displayStatus)
}

func displayStatus(ios *iostreams.IOStreams, status *shelly.WiFiStatus) {
	ios.Title("WiFi Status")
	ios.Println()

	ios.Printf("  Status:      %s\n", status.Status)
	ios.Printf("  SSID:        %s\n", valueOrEmpty(status.SSID))
	ios.Printf("  IP Address:  %s\n", valueOrEmpty(status.StaIP))
	ios.Printf("  Signal:      %d dBm\n", status.RSSI)
	if status.APCount > 0 {
		ios.Printf("  AP Clients:  %d\n", status.APCount)
	}
}

func valueOrEmpty(s string) string {
	if s == "" {
		return "<not connected>"
	}
	return s
}
