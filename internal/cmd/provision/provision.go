// Package provision provides device provisioning commands.
package provision

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/bulk"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/wifi"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory      *cmdutil.Factory
	SSID         string
	Password     string
	Subnet       string
	Timezone     string
	DeviceName   string
	Timeout      time.Duration
	BLEOnly      bool
	APOnly       bool
	NetworkOnly  bool
	RegisterOnly bool
	NoCloud      bool
	Yes          bool
}

// NewCommand creates the provision command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "provision",
		Aliases: []string{"prov", "setup"},
		Short:   "Discover and provision new Shelly devices",
		Long: `Discover and provision new Shelly devices on your network.

When run without a subcommand, provision scans for unprovisioned Shelly devices
using BLE (Gen2+), WiFi AP (Gen1), and network discovery (mDNS/CoIoT). Found
devices are presented for interactive selection and provisioned with WiFi
credentials automatically.

Gen2+ devices are provisioned via BLE (parallel, no network disruption).
Gen1 devices are provisioned via their WiFi AP (sequential, requires temporary
network switch to the device's AP).

Already-networked but unregistered devices are simply registered in the config.

Use the subcommands for targeted provisioning of specific devices:
  wifi   - Interactive WiFi provisioning for a single device
  ble    - BLE-based provisioning for a specific device
  bulk   - Bulk provisioning from a config file`,
		Example: `  # Auto-discover and provision all new devices
  shelly provision

  # Provide WiFi credentials via flags (non-interactive)
  shelly provision --ssid MyNetwork --password secret --yes

  # Only discover via BLE (Gen2+ devices)
  shelly provision --ble-only

  # Only discover via WiFi AP (Gen1 devices)
  shelly provision --ap-only

  # Only register already-networked devices (no provisioning)
  shelly provision --register-only

  # Scan a specific subnet for devices
  shelly provision --subnet 192.168.1.0/24

  # Interactive WiFi provisioning for a single device
  shelly provision wifi living-room

  # Bulk provision from config file
  shelly provision bulk devices.yaml

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "WiFi SSID for provisioning")
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi password for provisioning")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Discovery timeout")
	cmd.Flags().StringVar(&opts.Subnet, "subnet", "", "Subnet to scan (e.g., 192.168.1.0/24)")
	cmd.Flags().StringVar(&opts.DeviceName, "name", "", "Device name to assign after provisioning")
	cmd.Flags().StringVar(&opts.Timezone, "timezone", "", "Timezone to set on device")
	cmd.Flags().BoolVar(&opts.BLEOnly, "ble-only", false, "Only discover via BLE (Gen2+ devices)")
	cmd.Flags().BoolVar(&opts.APOnly, "ap-only", false, "Only discover via WiFi AP (Gen1 devices)")
	cmd.Flags().BoolVar(&opts.NetworkOnly, "network-only", false, "Only discover already-networked devices")
	cmd.Flags().BoolVar(&opts.RegisterOnly, "register-only", false, "Only register devices (skip provisioning)")
	cmd.Flags().BoolVar(&opts.NoCloud, "no-cloud", false, "Disable cloud on provisioned devices")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompts")

	cmd.AddCommand(wifi.NewCommand(f))
	cmd.AddCommand(bulk.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Prompt for WiFi credentials if not provided via flags (unless register-only)
	if !opts.RegisterOnly {
		if err := opts.promptWiFiCredentials(); err != nil {
			return err
		}
	}

	// Discovery phase
	onboardOpts := opts.buildOnboardOptions()
	devices, err := opts.runDiscovery(ctx, svc, onboardOpts)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	// Filter to unregistered devices
	unregistered := shelly.FilterUnregistered(devices)
	ios.Println()
	term.DisplayOnboardDevices(ios, unregistered)
	if len(unregistered) == 0 {
		return nil
	}

	// Interactive selection
	selected, err := term.SelectOnboardDevices(ios, unregistered, opts.Yes)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		ios.Info("No devices selected")
		return nil
	}

	// Provision and display results
	wifiCfg := &shelly.OnboardWiFiConfig{SSID: opts.SSID, Password: opts.Password}
	results := opts.provisionAll(ctx, svc, selected, wifiCfg, onboardOpts)
	term.DisplayOnboardSummary(ios, results)

	return nil
}

// promptWiFiCredentials prompts for SSID and password if not provided via flags.
func (o *Options) promptWiFiCredentials() error {
	if o.SSID != "" {
		return nil
	}

	ios := o.Factory.IOStreams()
	ssid, err := ios.Input("WiFi SSID:", "")
	if err != nil {
		return fmt.Errorf("SSID input failed: %w", err)
	}
	if ssid == "" {
		return fmt.Errorf("WiFi SSID is required")
	}
	o.SSID = ssid

	if o.Password == "" {
		pass, passErr := iostreams.Password("WiFi password:")
		if passErr != nil {
			return fmt.Errorf("password input failed: %w", passErr)
		}
		o.Password = pass
	}

	return nil
}

// buildOnboardOptions converts command Options to service OnboardOptions.
func (o *Options) buildOnboardOptions() *shelly.OnboardOptions {
	onboardOpts := &shelly.OnboardOptions{
		Subnet:       o.Subnet,
		Timezone:     o.Timezone,
		DeviceName:   o.DeviceName,
		Timeout:      o.Timeout,
		BLEOnly:      o.BLEOnly,
		APOnly:       o.APOnly,
		NetworkOnly:  o.NetworkOnly,
		RegisterOnly: o.RegisterOnly,
		NoCloud:      o.NoCloud,
	}
	if o.SSID != "" {
		onboardOpts.WiFi = &shelly.OnboardWiFiConfig{
			SSID:     o.SSID,
			Password: o.Password,
		}
	}
	return onboardOpts
}

// runDiscovery runs multi-protocol device discovery with progress output.
func (o *Options) runDiscovery(ctx context.Context, svc *shelly.Service, opts *shelly.OnboardOptions) ([]shelly.OnboardDevice, error) {
	ios := o.Factory.IOStreams()
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	mw.AddLine("ble", "BLE scan")
	mw.AddLine("wifi-ap", "WiFi AP scan")
	mw.AddLine("network", "Network scan")

	devices, err := svc.DiscoverForOnboard(ctx, opts, func(p shelly.OnboardProgress) {
		lineID := "network"
		switch p.Method {
		case "BLE":
			lineID = "ble"
		case "WiFi AP":
			lineID = "wifi-ap"
		}

		switch {
		case p.Done && p.Err != nil:
			mw.UpdateLine(lineID, iostreams.StatusError, fmt.Sprintf("%s: %v", p.Method, p.Err))
		case p.Done:
			status := iostreams.StatusSuccess
			if p.Found == 0 {
				status = iostreams.StatusSkipped
			}
			mw.UpdateLine(lineID, status, fmt.Sprintf("%s: %d found", p.Method, p.Found))
		default:
			mw.UpdateLine(lineID, iostreams.StatusRunning, fmt.Sprintf("%s: scanning...", p.Method))
		}
	})
	mw.Finalize()

	return devices, err
}

// provisionAll provisions devices grouped by their discovery source.
func (o *Options) provisionAll(
	ctx context.Context,
	svc *shelly.Service,
	selected []shelly.OnboardDevice,
	wifiCfg *shelly.OnboardWiFiConfig,
	onboardOpts *shelly.OnboardOptions,
) []*shelly.OnboardResult {
	ios := o.Factory.IOStreams()
	bleDevices, apDevices, networkDevices := shelly.SplitBySource(selected)
	var results []*shelly.OnboardResult

	// BLE devices in parallel
	if len(bleDevices) > 0 {
		ios.Println()
		ios.Title("Provisioning %d BLE device(s)...", len(bleDevices))
		bleResults := svc.OnboardBLEParallel(ctx, bleDevices, wifiCfg, onboardOpts)
		results = append(results, bleResults...)
		term.DisplayOnboardResults(ios, bleResults)
	}

	// AP devices sequentially (requires network switching)
	if len(apDevices) > 0 {
		ios.Println()
		ios.Title("Provisioning %d WiFi AP device(s)...", len(apDevices))
		for _, dev := range apDevices {
			ios.Printf("  Connecting to %s...\n", dev.SSID)
			r := svc.OnboardViaAP(ctx, dev, wifiCfg, onboardOpts)
			results = append(results, r)
			term.DisplayOnboardResults(ios, []*shelly.OnboardResult{r})
		}
	}

	// Already-networked devices: just register
	if len(networkDevices) > 0 {
		ios.Println()
		ios.Title("Registering %d networked device(s)...", len(networkDevices))
		netResults := shelly.RegisterNetworkDevices(networkDevices)
		results = append(results, netResults...)
		term.DisplayOnboardResults(ios, netResults)
	}

	return results
}
