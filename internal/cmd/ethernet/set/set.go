// Package set provides the ethernet set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	staticIPFlag   string
	gatewayFlag    string
	netmaskFlag    string
	nameserverFlag string
	enableFlag     bool
	disableFlag    bool
)

// NewCommand creates the ethernet set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <device>",
		Short: "Configure Ethernet connection",
		Long: `Configure the Ethernet connection for a device.

By default, Ethernet uses DHCP for automatic IP configuration. Use static IP
options to configure a fixed IP address.

Only available on Shelly Pro devices with an Ethernet port.`,
		Example: `  # Enable Ethernet with DHCP
  shelly ethernet set living-room-pro --enable

  # Configure static IP
  shelly ethernet set living-room-pro --enable \
    --static-ip "192.168.1.50" --gateway "192.168.1.1" \
    --netmask "255.255.255.0" --dns "8.8.8.8"

  # Disable Ethernet
  shelly ethernet set living-room-pro --disable`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&staticIPFlag, "static-ip", "", "Static IP address (uses DHCP if not set)")
	cmd.Flags().StringVar(&gatewayFlag, "gateway", "", "Gateway address (for static IP)")
	cmd.Flags().StringVar(&netmaskFlag, "netmask", "", "Network mask (for static IP)")
	cmd.Flags().StringVar(&nameserverFlag, "dns", "", "DNS server address (for static IP)")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable Ethernet")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Disable Ethernet")

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

	// Determine IPv4 mode
	ipv4Mode := ""
	if staticIPFlag != "" {
		ipv4Mode = "static"
	} else if enableFlag {
		ipv4Mode = "dhcp"
	}

	// Validate flags
	if enable == nil && staticIPFlag == "" {
		return fmt.Errorf("specify --enable, --disable, or configuration options")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring Ethernet...", func(ctx context.Context) error {
		if err := svc.SetEthernetConfig(ctx, device, enable, ipv4Mode, staticIPFlag, netmaskFlag, gatewayFlag, nameserverFlag); err != nil {
			return fmt.Errorf("failed to configure Ethernet: %w", err)
		}

		switch {
		case disableFlag:
			ios.Success("Ethernet disabled on %s", device)
		case staticIPFlag != "":
			ios.Success("Ethernet configured with static IP on %s", device)
			ios.Printf("  IP: %s\n", staticIPFlag)
		default:
			ios.Success("Ethernet enabled with DHCP on %s", device)
		}
		return nil
	})
}
