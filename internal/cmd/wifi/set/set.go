// Package set provides the wifi set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
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
	Open     bool
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
static IP settings instead of using DHCP.

Without --password, a device that stays on the same network keeps the
password it has. A different network takes the password this host has stored
for it; with none, the command is refused. Use --open for a network that has
no password.`,
		Example: `  # Connect to a WiFi network
  shelly wifi set living-room --ssid "MyNetwork" --password "secret"

  # Join a network this host knows, using its stored password
  shelly wifi set living-room --ssid "MyNetwork"

  # Join a network that has no password
  shelly wifi set living-room --ssid "GuestNet" --open

  # Configure static IP
  shelly wifi set living-room --ssid "MyNetwork" --password "secret" \
    --static-ip "192.168.1.50" --gateway "192.168.1.1" --netmask "255.255.255.0"

  # Change only the address; the gateway and netmask stay the device's
  shelly wifi set living-room --ssid "MyNetwork" --static-ip "192.168.1.51"

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
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password)
	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmdutil.AddStaticIPFlags(cmd, &opts.StaticIP, &opts.Gateway, &opts.Netmask, &opts.DNS,
		"Static IPv4 address (DHCP when not set; --gateway, --netmask and --dns default to the device's current ones)",
		"the device's current one")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable WiFi station mode")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable WiFi station mode")
	cmd.MarkFlagsMutuallyExclusive("enable", "disable")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Configuring or enabling WiFi station mode needs an SSID; only --disable
	// may omit it, so --enable alone never pushes an empty SSID.
	if opts.SSID == "" && !opts.Disable {
		return fmt.Errorf("--ssid is required (or use --disable to disable WiFi)")
	}
	flags := cmdutil.NetworkFlags{
		SSID: opts.SSID, Password: opts.Password, Open: opts.Open,
		StaticIP: opts.StaticIP, Gateway: opts.Gateway, Netmask: opts.Netmask, DNS: opts.DNS,
	}
	if err := flags.Validate(false); err != nil {
		return err
	}

	write := network.StationWrite{
		SSID:     opts.SSID,
		Password: opts.Password,
		Open:     opts.Open,
		StaticIP: opts.StaticIP,
		Gateway:  opts.Gateway,
		Netmask:  opts.Netmask,
		DNS:      opts.DNS,
	}
	switch {
	case opts.Enable:
		write.Enable = new(true)
	case opts.Disable:
		write.Enable = new(false)
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring WiFi...", func(ctx context.Context) error {
		warnings, setErr := svc.SetWiFiConfig(ctx, opts.Device, write)
		if setErr != nil {
			return fmt.Errorf("failed to configure WiFi: %w", setErr)
		}
		for _, w := range warnings {
			ios.Warning("%s", w)
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
