// Package ap provides the wifi ap subcommand.
package ap

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Clients  bool
	Device   string
	Disable  bool
	Enable   bool
	Password string
	SSID     string
}

// NewCommand creates the wifi ap command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "ap <device>",
		Aliases: []string{"accesspoint", "hotspot"},
		Short:   "Configure WiFi access point",
		Long: `Configure the WiFi access point (AP) mode for a device.

When enabled, the device creates its own WiFi network that other devices
can connect to. Use --clients to list connected clients.`,
		Example: `  # Enable access point with custom SSID
  shelly wifi ap living-room --enable --ssid "ShellyAP" --password "secret"

  # Disable access point
  shelly wifi ap living-room --disable

  # List connected clients
  shelly wifi ap living-room --clients`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "Access point SSID")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Access point password")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable access point")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable access point")
	cmd.Flags().BoolVar(&opts.Clients, "clients", false, "List connected clients")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// If --clients flag, list connected clients
	if opts.Clients {
		return cmdutil.RunList(ctx, ios, svc, opts.Device,
			"Getting connected clients...",
			"No clients connected to access point",
			func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.WiFiAPClient, error) {
				return svc.ListWiFiAPClients(ctx, device)
			},
			term.DisplayWiFiAPClients)
	}

	// Determine enable state
	var enable *bool
	if opts.Enable {
		t := true
		enable = &t
	} else if opts.Disable {
		f := false
		enable = &f
	}

	// Validate flags - need either enable/disable or configuration
	if enable == nil && opts.SSID == "" && opts.Password == "" {
		return fmt.Errorf("specify --enable, --disable, or configuration options (--ssid, --password)")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring access point...", func(ctx context.Context) error {
		if err := svc.SetWiFiAPConfig(ctx, opts.Device, opts.SSID, opts.Password, enable); err != nil {
			return fmt.Errorf("failed to configure access point: %w", err)
		}

		if opts.Disable {
			ios.Success("Access point disabled on %s", opts.Device)
		} else {
			ios.Success("Access point configured on %s", opts.Device)
			if opts.SSID != "" {
				ios.Printf("  SSID: %s\n", opts.SSID)
			}
		}
		return nil
	})
}
