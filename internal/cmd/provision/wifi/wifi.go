// Package wifi provides interactive WiFi provisioning.
package wifi

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

// Options holds command options.
type Options struct {
	Device   string
	SSID     string
	Password string
	Open     bool
	NoScan   bool
	Factory  *cmdutil.Factory
}

// NewCommand creates the provision wifi command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "wifi <device>",
		Aliases: []string{"network", "wlan"},
		Short:   "Interactive WiFi provisioning",
		Long: `Provision WiFi settings interactively for a device.

By default, this command scans for available networks and prompts you to select one.
You can also provide SSID and password directly via flags.

When no password is given and none is typed at the prompt, a device that stays
on the same network keeps the password it has, and a different network takes
the password this host has stored for it; with none, the command is refused.
Use --open for a network that has no password.`,
		Example: `  # Interactive provisioning with network scan
  shelly provision wifi living-room

  # Direct provisioning with credentials
  shelly provision wifi living-room --ssid "MyNetwork" --password "secret"

  # Join a network that has no password
  shelly provision wifi living-room --ssid "GuestNet" --open

  # Skip scan and prompt for SSID
  shelly provision wifi living-room --no-scan`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "WiFi network name (skip selection)")
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password)
	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmd.Flags().BoolVar(&opts.NoScan, "no-scan", false, "Skip network scan, prompt for SSID")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Get SSID if not provided
	if opts.SSID == "" && opts.NoScan {
		// Prompt for SSID
		ssid, err := ios.Input("WiFi network name (SSID):", "")
		if err != nil {
			return fmt.Errorf("failed to get SSID: %w", err)
		}
		opts.SSID = ssid
		if opts.SSID == "" {
			return fmt.Errorf("WiFi SSID is required with --no-scan")
		}
	}

	if opts.SSID == "" && !opts.NoScan {
		ssid, err := cmdutil.SelectWiFiNetwork(ctx, ios, svc, opts.Device)
		if err != nil {
			return err
		}
		opts.SSID = ssid
	}

	// A blank answer means "not known", so the service decides from the
	// device's current network and this host's stored credentials.
	if opts.Password == "" && !opts.Open && ios.CanPrompt() {
		password, err := iostreams.Password("WiFi password (blank to keep or look up):")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}
		opts.Password = password
	}

	// Apply configuration
	ios.Info("Configuring WiFi...")
	write := network.StationWrite{SSID: opts.SSID, Password: opts.Password, Open: opts.Open, Enable: new(true)}
	warnings, err := svc.SetWiFiConfig(ctx, opts.Device, write)
	if err != nil {
		return fmt.Errorf("failed to configure WiFi: %w", err)
	}
	for _, w := range warnings {
		ios.Warning("%s", w)
	}

	ios.Success("WiFi configured on %q", opts.Device)
	ios.Info("  SSID: %s", opts.SSID)
	ios.Info("The device will attempt to connect to the network.")

	return nil
}
