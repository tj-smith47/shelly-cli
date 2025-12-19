// Package wifi provides interactive WiFi provisioning.
package wifi

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device   string
	SSID     string
	Password string
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
You can also provide SSID and password directly via flags.`,
		Example: `  # Interactive provisioning with network scan
  shelly provision wifi living-room

  # Direct provisioning with credentials
  shelly provision wifi living-room --ssid "MyNetwork" --password "secret"

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
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi password")
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
	}

	if opts.SSID == "" {
		// Scan and select
		ssid, err := scanAndSelect(ctx, ios, svc, opts.Device)
		if err != nil {
			return err
		}
		opts.SSID = ssid
	}

	// Get password if not provided
	if opts.Password == "" {
		password, err := iostreams.Password("WiFi password:")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}
		opts.Password = password
	}

	// Apply configuration
	ios.Info("Configuring WiFi...")
	enable := true
	if err := svc.SetWiFiConfig(ctx, opts.Device, opts.SSID, opts.Password, &enable); err != nil {
		return fmt.Errorf("failed to configure WiFi: %w", err)
	}

	ios.Success("WiFi configured on %q", opts.Device)
	ios.Info("  SSID: %s", opts.SSID)
	ios.Info("The device will attempt to connect to the network.")

	return nil
}

func scanAndSelect(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string) (string, error) {
	ios.Info("Scanning for networks...")

	results, err := svc.ScanWiFi(ctx, device)
	if err != nil {
		return "", fmt.Errorf("network scan failed: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no networks found")
	}

	// Dedupe by SSID, keeping strongest signal
	seen := make(map[string]shelly.WiFiScanResult)
	for _, r := range results {
		if r.SSID == "" {
			continue
		}
		existing, exists := seen[r.SSID]
		if !exists || r.RSSI > existing.RSSI {
			seen[r.SSID] = r
		}
	}

	// Convert to slice and sort by signal strength
	networks := make([]shelly.WiFiScanResult, 0, len(seen))
	for _, n := range seen {
		networks = append(networks, n)
	}

	sort.Slice(networks, func(i, j int) bool {
		return networks[i].RSSI > networks[j].RSSI
	})

	// Build selection options
	options := make([]string, len(networks))
	for i, n := range networks {
		signal := output.FormatWiFiSignalStrength(n.RSSI)
		options[i] = fmt.Sprintf("%s (%s, ch %d)", n.SSID, signal, n.Channel)
	}

	selected, err := ios.Select("Select WiFi network:", options, 0)
	if err != nil {
		return "", fmt.Errorf("network selection failed: %w", err)
	}

	// Find the selected network by matching the option string
	for i, opt := range options {
		if opt == selected {
			return networks[i].SSID, nil
		}
	}

	return "", fmt.Errorf("selected network not found")
}
