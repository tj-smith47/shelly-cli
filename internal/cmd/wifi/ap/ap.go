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

var (
	ssidFlag     string
	passwordFlag string
	enableFlag   bool
	disableFlag  bool
	clientsFlag  bool
)

// NewCommand creates the wifi ap command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&ssidFlag, "ssid", "", "Access point SSID")
	cmd.Flags().StringVar(&passwordFlag, "password", "", "Access point password")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable access point")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Disable access point")
	cmd.Flags().BoolVar(&clientsFlag, "clients", false, "List connected clients")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// If --clients flag, list connected clients
	if clientsFlag {
		return cmdutil.RunList(ctx, ios, svc, device,
			"Getting connected clients...",
			"No clients connected to access point",
			func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.WiFiAPClient, error) {
				return svc.ListWiFiAPClients(ctx, device)
			},
			term.DisplayWiFiAPClients)
	}

	// Determine enable state
	var enable *bool
	if enableFlag {
		t := true
		enable = &t
	} else if disableFlag {
		f := false
		enable = &f
	}

	// Validate flags - need either enable/disable or configuration
	if enable == nil && ssidFlag == "" && passwordFlag == "" {
		return fmt.Errorf("specify --enable, --disable, or configuration options (--ssid, --password)")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring access point...", func(ctx context.Context) error {
		if err := svc.SetWiFiAPConfig(ctx, device, ssidFlag, passwordFlag, enable); err != nil {
			return fmt.Errorf("failed to configure access point: %w", err)
		}

		if disableFlag {
			ios.Success("Access point disabled on %s", device)
		} else {
			ios.Success("Access point configured on %s", device)
			if ssidFlag != "" {
				ios.Printf("  SSID: %s\n", ssidFlag)
			}
		}
		return nil
	})
}
