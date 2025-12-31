// Package set provides the wifi set subcommand.
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
	Factory  *cmdutil.Factory
	Device   string
	Disable  bool
	DNS      string
	Enable   bool
	Gateway  string
	Netmask  string
	Password string
	SSID     string
	StaticIP string
}

// NewCommand creates the wifi set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"configure", "config"},
		Short:   "Configure WiFi connection",
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
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "WiFi network name")
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi password")
	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Static IP address (uses DHCP if not set)")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Gateway address (for static IP)")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Network mask (for static IP)")
	cmd.Flags().StringVar(&opts.DNS, "dns", "", "DNS server address (for static IP)")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable WiFi station mode")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable WiFi station mode")

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

	// Validate flags
	if opts.SSID == "" && !opts.Disable && enable == nil {
		return fmt.Errorf("--ssid is required (or use --disable to disable WiFi)")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring WiFi...", func(ctx context.Context) error {
		if err := svc.SetWiFiConfig(ctx, opts.Device, opts.SSID, opts.Password, enable); err != nil {
			return fmt.Errorf("failed to configure WiFi: %w", err)
		}

		if opts.Disable {
			ios.Success("WiFi station mode disabled on %s", opts.Device)
		} else {
			ios.Success("WiFi configured on %s", opts.Device)
			if opts.SSID != "" {
				ios.Printf("  SSID: %s\n", opts.SSID)
			}
		}
		return nil
	})
}
