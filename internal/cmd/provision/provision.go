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
	Timezone     string
	DeviceName   string
	FromDevice   string
	FromTemplate string
	Timeout      time.Duration
	BLEOnly      bool
	APOnly       bool
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
using BLE (Gen2+) and WiFi AP (Gen1). Found devices are presented for
interactive selection and provisioned with WiFi credentials automatically.

WiFi credentials are resolved in order: --from-device backup, --ssid/--password
flags, auto-detected from an existing Gen1 device, or prompted interactively.

Use --from-device to clone an existing device's full configuration (WiFi, MQTT,
cloud, light settings, schedules, etc.) onto newly provisioned devices. Use
--from-template to apply a saved device template instead.

Gen2+ devices are provisioned via BLE (parallel, no network disruption).
Gen1 devices are provisioned via their WiFi AP (sequential, requires temporary
network switch to the device's AP).

Use the subcommands for targeted provisioning of specific devices:
  wifi   - Interactive WiFi provisioning for a single device
  ble    - BLE-based provisioning for a specific device
  bulk   - Bulk provisioning from a config file

To register already-networked devices, use: shelly discover --register`,
		Example: `  # Auto-discover and provision all new devices
  shelly provision

  # Clone config from an existing device onto new devices
  shelly provision --from-device living-room --ap-only

  # Apply a saved template to new devices
  shelly provision --from-template bulb-config --ap-only -y

  # Provide WiFi credentials via flags (non-interactive)
  shelly provision --ssid MyNetwork --password secret --yes

  # Only discover via BLE (Gen2+ devices)
  shelly provision --ble-only

  # Only discover via WiFi AP (Gen1 devices)
  shelly provision --ap-only

  # Interactive WiFi provisioning for a single device
  shelly provision wifi living-room

  # BLE-based provisioning for new device
  shelly provision ble 192.168.33.1`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.SSID, "ssid", "", "WiFi SSID for provisioning")
	cmd.Flags().StringVar(&opts.Password, "password", "", "WiFi password for provisioning")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Discovery timeout")
	cmd.Flags().StringVar(&opts.DeviceName, "name", "", "Device name to assign after provisioning")
	cmd.Flags().StringVar(&opts.Timezone, "timezone", "", "Timezone to set on device")
	cmd.Flags().BoolVar(&opts.BLEOnly, "ble-only", false, "Only discover via BLE (Gen2+ devices)")
	cmd.Flags().BoolVar(&opts.APOnly, "ap-only", false, "Only discover via WiFi AP (Gen1 devices)")
	cmd.Flags().StringVar(&opts.FromDevice, "from-device", "", "Clone config from existing device")
	cmd.Flags().StringVar(&opts.FromTemplate, "from-template", "", "Apply saved template after provisioning")
	cmd.Flags().BoolVar(&opts.NoCloud, "no-cloud", false, "Disable cloud on provisioned devices")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.Yes, "all", false, "Provision all discovered devices (non-interactive)")
	cmd.MarkFlagsMutuallyExclusive("from-device", "from-template")

	cmd.AddCommand(wifi.NewCommand(f))
	cmd.AddCommand(bulk.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Load provision source (--from-device or --from-template)
	var source *shelly.ProvisionSource
	if opts.FromDevice != "" || opts.FromTemplate != "" {
		label := opts.FromDevice
		if label == "" {
			label = opts.FromTemplate
		}
		ios.StartProgress(fmt.Sprintf("Loading config from %s...", label))
		var err error
		source, err = svc.LoadProvisionSource(ctx, opts.FromDevice, opts.FromTemplate)
		ios.StopProgress()
		if err != nil {
			return err
		}
		ios.Success("Config loaded from %s", label)

		// Use WiFi creds from source if available and not overridden by flags
		if source.WiFi != nil && opts.SSID == "" {
			opts.SSID = source.WiFi.SSID
			opts.Password = source.WiFi.Password
			ios.Info("Using WiFi credentials from source: %s", source.WiFi.SSID)
		}
	}

	// Resolve WiFi credentials: flags/source → auto-detect → prompt
	if err := opts.promptWiFiCredentials(ctx); err != nil {
		return err
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
	results := opts.provisionAll(ctx, svc, selected, wifiCfg, onboardOpts, source)
	term.DisplayOnboardSummary(ios, results)

	return nil
}

// promptWiFiCredentials resolves WiFi credentials for provisioning. Tries in order:
// 1. Flags (--ssid/--password) — already set, return immediately
// 2. Auto-detect from an existing Gen1 device on the network
// 3. Interactive prompt.
func (o *Options) promptWiFiCredentials(ctx context.Context) error {
	if o.SSID != "" {
		return nil
	}

	ios := o.Factory.IOStreams()
	svc := o.Factory.ShellyService()

	// Try to auto-detect from an existing device
	ios.StartProgress("Detecting WiFi credentials from existing devices...")
	creds := svc.GetWiFiCredentials(ctx)
	ios.StopProgress()

	if creds != nil {
		ios.Success("WiFi credentials detected from network: %s", creds.SSID)
		o.SSID = creds.SSID
		o.Password = creds.Password
		return nil
	}

	// Fall back to interactive prompt
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
		Timezone:   o.Timezone,
		DeviceName: o.DeviceName,
		Timeout:    o.Timeout,
		BLEOnly:    o.BLEOnly,
		APOnly:     o.APOnly,
		NoCloud:    o.NoCloud,
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
	if !opts.APOnly {
		mw.AddLine("ble", "BLE scan")
	}
	if !opts.BLEOnly {
		mw.AddLine("wifi-ap", "WiFi AP scan")
	}

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
// If source is non-nil, applies the source config to each device after provisioning.
func (o *Options) provisionAll(
	ctx context.Context,
	svc *shelly.Service,
	selected []shelly.OnboardDevice,
	wifiCfg *shelly.OnboardWiFiConfig,
	onboardOpts *shelly.OnboardOptions,
	source *shelly.ProvisionSource,
) []*shelly.OnboardResult {
	ios := o.Factory.IOStreams()
	bleDevices, apDevices, _ := shelly.SplitBySource(selected)
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

	// Apply source config (--from-device or --from-template) to each provisioned device
	if source != nil {
		o.applySourceConfig(ctx, svc, results, source)
	}

	return results
}

// applySourceConfig applies a provision source config to all successfully provisioned devices.
func (o *Options) applySourceConfig(
	ctx context.Context,
	svc *shelly.Service,
	results []*shelly.OnboardResult,
	source *shelly.ProvisionSource,
) {
	ios := o.Factory.IOStreams()
	ios.Println()
	ios.Title("Applying source config to provisioned devices...")

	for _, r := range results {
		if r.Error != nil || r.NewAddress == "" {
			continue
		}
		ios.Printf("  %s (%s)... ", r.Device.Name, r.NewAddress)
		if err := svc.ApplyProvisionSource(ctx, r.NewAddress, source); err != nil {
			ios.Error("failed: %v", err)
		} else {
			ios.Success("done")
		}
	}
}
