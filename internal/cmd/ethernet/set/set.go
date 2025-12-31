// Package set provides the ethernet set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	Device     string
	Disable    bool
	Enable     bool
	Gateway    string
	Nameserver string
	Netmask    string
	StaticIP   string
}

// NewCommand creates the ethernet set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"configure", "config"},
		Short:   "Configure Ethernet connection",
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
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Static IP address (uses DHCP if not set)")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Gateway address (for static IP)")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Network mask (for static IP)")
	cmd.Flags().StringVar(&opts.Nameserver, "dns", "", "DNS server address (for static IP)")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable Ethernet")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable Ethernet")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Determine enable state
	var enable *bool
	if opts.Enable {
		t := true
		enable = &t
	} else if opts.Disable {
		f := false
		enable = &f
	}

	// Determine IPv4 mode
	ipv4Mode := ""
	if opts.StaticIP != "" {
		ipv4Mode = "static"
	} else if opts.Enable {
		ipv4Mode = "dhcp"
	}

	// Validate flags
	if enable == nil && opts.StaticIP == "" {
		return fmt.Errorf("specify --enable, --disable, or configuration options")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring Ethernet...", func(ctx context.Context) error {
		if err := svc.SetEthernetConfig(ctx, opts.Device, enable, ipv4Mode, opts.StaticIP, opts.Netmask, opts.Gateway, opts.Nameserver); err != nil {
			return fmt.Errorf("failed to configure Ethernet: %w", err)
		}

		switch {
		case opts.Disable:
			ios.Success("Ethernet disabled on %s", opts.Device)
		case opts.StaticIP != "":
			ios.Success("Ethernet configured with static IP on %s", opts.Device)
			ios.Printf("  IP: %s\n", opts.StaticIP)
		default:
			ios.Success("Ethernet enabled with DHCP on %s", opts.Device)
		}
		return nil
	})
}
