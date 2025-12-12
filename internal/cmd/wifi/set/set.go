// Package set provides the wifi set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	ssidFlag       string
	passwordFlag   string
	staticIPFlag   string
	gatewayFlag    string
	netmaskFlag    string
	nameserverFlag string
	enableFlag     bool
	disableFlag    bool
)

// NewCommand creates the wifi set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <device>",
		Short: "Configure WiFi connection",
		Long: `Configure the WiFi station (client) connection for a device.

Set the SSID and password to connect to a WiFi network. Optionally configure
static IP settings instead of using DHCP.`,
		Example: `  # Connect to a WiFi network
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Configure static IP
  shelly wifi set living-room --ssid "MyNetwork" --password "secret" \
    --static-ip "192.168.1.50" --gateway "192.168.1.1" --netmask "255.255.255.0"

  # Disable WiFi station mode
  shelly wifi set living-room --disable`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&ssidFlag, "ssid", "", "WiFi network name")
	cmd.Flags().StringVar(&passwordFlag, "password", "", "WiFi password")
	cmd.Flags().StringVar(&staticIPFlag, "static-ip", "", "Static IP address (uses DHCP if not set)")
	cmd.Flags().StringVar(&gatewayFlag, "gateway", "", "Gateway address (for static IP)")
	cmd.Flags().StringVar(&netmaskFlag, "netmask", "", "Network mask (for static IP)")
	cmd.Flags().StringVar(&nameserverFlag, "dns", "", "DNS server address (for static IP)")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable WiFi station mode")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Disable WiFi station mode")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Determine enable state
	var enable *bool
	if enableFlag {
		t := true
		enable = &t
	} else if disableFlag {
		f := false
		enable = &f
	}

	// Validate flags
	if ssidFlag == "" && !disableFlag && enable == nil {
		return fmt.Errorf("--ssid is required (or use --disable to disable WiFi)")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring WiFi...", func(ctx context.Context) error {
		if err := svc.SetWiFiConfig(ctx, device, ssidFlag, passwordFlag, enable); err != nil {
			return fmt.Errorf("failed to configure WiFi: %w", err)
		}

		if disableFlag {
			ios.Success("WiFi station mode disabled on %s", device)
		} else {
			ios.Success("WiFi configured on %s", device)
			if ssidFlag != "" {
				ios.Printf("  SSID: %s\n", ssidFlag)
			}
		}
		return nil
	})
}
