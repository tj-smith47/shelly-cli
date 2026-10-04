// Package provision provides device provisioning commands.
package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/ble"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/bulk"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/inspect"
	"github.com/tj-smith47/shelly-cli/internal/cmd/provision/wifi"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Discovery method labels used for progress line IDs.
const (
	methodBLE    = "BLE"
	methodWiFiAP = "WiFi AP"
)

// provisionService is the subset of *shelly.Service the provision command drives.
// Depending on this narrow interface instead of the concrete service keeps the
// orchestration testable with a stub: the concrete service reaches real devices
// over BLE and hops the host's WiFi to onboard Gen1 APs, which a unit test cannot
// exercise.
type provisionService interface {
	LoadProvisionSource(ctx context.Context, fromDevice, fromTemplate string) (*shelly.ProvisionSource, error)
	GetWiFiCredentials(ctx context.Context) *shelly.OnboardWiFiConfig
	HostWiFiCredentials(ctx context.Context) *shelly.OnboardWiFiConfig
	HostWiFiPassword(ctx context.Context, ssid string) (string, error)
	DiscoverForOnboard(ctx context.Context, opts *shelly.OnboardOptions, progress func(shelly.OnboardProgress)) ([]shelly.OnboardDevice, error)
	OnboardBLEParallel(ctx context.Context, devices []*shelly.OnboardDevice, wifiCfg *shelly.OnboardWiFiConfig, opts *shelly.OnboardOptions) []*shelly.OnboardResult
	OnboardViaAP(ctx context.Context, device *shelly.OnboardDevice, wifi *shelly.OnboardWiFiConfig, opts *shelly.OnboardOptions) *shelly.OnboardResult
	ApplyProvisionSource(ctx context.Context, deviceAddr string, source *shelly.ProvisionSource) error
}

// Options holds command options.
type Options struct {
	Factory       *cmdutil.Factory
	SSID          string
	Password      string
	PasswordStdin bool
	Open          bool
	Timezone      string
	DeviceName    string
	FromDevice    string
	FromTemplate  string
	StaticIP      string
	Gateway       string
	Netmask       string
	DNS           string
	TargetAP      string
	Timeout       time.Duration
	BLEOnly       bool
	APOnly        bool
	NoCloud       bool
	DisableAP     bool
	Yes           bool
	DiscoverOnly  bool

	// svc, when non-nil, overrides the service resolved from the Factory. It is the
	// test injection seam; production leaves it nil and uses Factory.ShellyService().
	svc provisionService

	// askPassword and askConfirm, when non-nil, replace the terminal prompts. They
	// are the test injection seams for the interactive credential flow.
	askPassword func(message string) (string, error)
	askConfirm  func(message string, defaultValue bool) (bool, error)
}

// service returns the injected provisionService when set, otherwise the concrete
// service from the Factory.
func (o *Options) service() provisionService {
	if o.svc != nil {
		return o.svc
	}
	return o.Factory.ShellyService()
}

// NewCommand creates the provision command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "provision",
		Aliases: []string{"prov"},
		Short:   "Discover and provision new Shelly devices",
		Long: `Discover and provision new Shelly devices on your network.

When run without a subcommand, provision scans for unprovisioned Shelly devices
using BLE (Gen2+) and WiFi AP (Gen1). Found devices are presented for
interactive selection and provisioned with WiFi credentials automatically.

WiFi credentials are resolved in this order:
  1. --ssid, --password and --open, which win over a --from-device source
  2. the --from-device source's network, password and open state
  3. with no network named: a registered Gen1 device's network, else the
     network this host is on, with its stored password
  4. a prompt for the network name
  5. with no password: this host's stored password for the network, else a
     prompt (without a terminal or with --yes, the stored password is looked up
     at onboarding and a network with none is refused)
A network with no password is joined only with --open, a source device whose
network has no password, or an empty password confirmed at the prompt.

Use --from-device to clone an existing device's full configuration (WiFi, MQTT,
cloud, light settings, schedules, etc.) onto newly provisioned devices. Use
--from-template to apply a saved device template instead.

Gen2+ devices are provisioned via BLE (parallel, no network disruption).
Gen1 devices are provisioned via their WiFi AP (sequential, requires temporary
network switch to the device's AP).

A Gen2+ device onboarded through its WiFi AP keeps that open access point on
after it joins the network. Use --disable-ap to turn it off once the device
answers on the LAN; a device that did not join keeps its access point, so it
can still be reached there. Gen1 devices leave access point mode by themselves
and BLE onboarding does not use the access point, so the flag changes nothing
for them.

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

  # Join a network that has no password
  shelly provision --ssid GuestWiFi --open --yes

  # List discoverable APs as JSON (for scripted before/after scan-diff)
  shelly provision --ap-only --discover-only

  # Onboard one specific AP non-interactively with a static IP
  shelly provision --ap-only --target-ap shellycolorbulb-AABBCC --name master-bath \
    --static-ip 10.23.47.227 --gateway 10.23.47.1 --netmask 255.255.254.0 --dns 10.23.47.1 --yes

  # Onboard a Gen2+ device at its AP and turn that AP off once it is on the LAN
  shelly provision --ap-only --target-ap ShellyPlus1PM-AABBCC --disable-ap --yes

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
	cmdutil.AddWiFiPasswordFlag(cmd, &opts.Password, &opts.PasswordStdin)
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", shelly.DefaultOnboardScanTimeout, "Discovery timeout")
	cmd.Flags().StringVar(&opts.DeviceName, "name", "", "Device name to assign after provisioning")
	cmd.Flags().StringVar(&opts.Timezone, "timezone", "", "Timezone to set on device")
	cmd.Flags().BoolVar(&opts.BLEOnly, "ble-only", false, "Only discover via BLE (Gen2+ devices)")
	cmd.Flags().BoolVar(&opts.APOnly, "ap-only", false, "Only discover via WiFi AP (Gen1 devices)")
	cmd.Flags().StringVar(&opts.FromDevice, "from-device", "", "Clone config from existing device")
	cmd.Flags().StringVar(&opts.FromTemplate, "from-template", "", "Apply saved template after provisioning")
	cmd.Flags().BoolVar(&opts.NoCloud, "no-cloud", false, "Disable cloud on provisioned devices")
	cmd.Flags().BoolVar(&opts.DisableAP, "disable-ap", false, "Turn off a Gen2+ device's access point once it answers on the LAN (WiFi AP onboarding)")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.Yes, "all", false, "Provision all discovered devices (non-interactive)")
	cmd.Flags().StringVar(&opts.StaticIP, "static-ip", "", "Assign a static IP to the device (requires --gateway and --netmask)")
	cmd.Flags().StringVar(&opts.Gateway, "gateway", "", "Gateway for the static IP")
	cmd.Flags().StringVar(&opts.Netmask, "netmask", "", "Netmask for the static IP (e.g. 255.255.254.0)")
	cmd.Flags().StringVar(&opts.DNS, "dns", "", "DNS server for the static IP")
	cmd.Flags().StringVar(&opts.TargetAP, "target-ap", "", "Provision only the device whose AP SSID matches (non-interactive single device)")
	cmd.Flags().BoolVar(&opts.DiscoverOnly, "discover-only", false, "List discoverable unprovisioned devices as JSON and exit (no provisioning)")
	cmdutil.AddOpenFlag(cmd, &opts.Open)
	cmd.MarkFlagsMutuallyExclusive("from-device", "from-template")
	cmd.MarkFlagsMutuallyExclusive("ble-only", "disable-ap")
	cmd.MarkFlagsRequiredTogether("static-ip", "gateway", "netmask")

	cmd.AddCommand(wifi.NewCommand(f))
	cmd.AddCommand(bulk.NewCommand(f))
	cmd.AddCommand(ble.NewCommand(f))
	cmd.AddCommand(inspect.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	if err := cmdutil.ReadWiFiPasswordStdin(ios, &opts.Password, opts.PasswordStdin); err != nil {
		return err
	}
	svc := opts.service()

	// Load provision source + resolve WiFi credentials. Skipped for
	// --discover-only, which just lists devices and needs no credentials.
	var source *shelly.ProvisionSource
	if !opts.DiscoverOnly {
		var err error
		if source, err = opts.resolveSourceAndCreds(ctx, svc); err != nil {
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

	// --discover-only: emit the list as JSON and stop (scriptable scan/diff).
	if opts.DiscoverOnly {
		return opts.outputDiscovered(unregistered)
	}

	ios.Println()
	term.DisplayOnboardDevices(ios, unregistered)
	if len(unregistered) == 0 {
		return nil
	}

	// Select devices: a single --target-ap match, or interactive selection.
	selected, err := opts.selectDevices(unregistered)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		ios.Info("No devices selected")
		return nil
	}

	// Provision and display results
	wifiCfg := opts.buildWiFiConfig()
	results := opts.provisionAll(ctx, svc, selected, wifiCfg, onboardOpts, source)
	term.DisplayOnboardSummary(ios, results)

	return nil
}

// resolveSourceAndCreds loads the optional --from-device/--from-template config
// source and resolves WiFi credentials (flags/source → auto-detect → prompt).
func (o *Options) resolveSourceAndCreds(ctx context.Context, svc provisionService) (*shelly.ProvisionSource, error) {
	ios := o.Factory.IOStreams()
	var source *shelly.ProvisionSource
	if o.FromDevice != "" || o.FromTemplate != "" {
		label := o.FromDevice
		if label == "" {
			label = o.FromTemplate
		}
		ios.StartProgress(fmt.Sprintf("Loading config from %s...", label))
		var err error
		source, err = svc.LoadProvisionSource(ctx, o.FromDevice, o.FromTemplate)
		ios.StopProgress()
		if err != nil {
			return nil, err
		}
		ios.Success("Config loaded from %s", label)

		o.adoptSourceWiFi(source.WiFi)
	}

	// Resolve WiFi credentials: flags/source → auto-detect → prompt
	if err := o.promptWiFiCredentials(ctx); err != nil {
		return nil, err
	}
	return source, nil
}

// adoptSourceWiFi uses the network a provision source records unless --ssid
// named one. A --password or --open flag keeps precedence over the source's key
// and open state.
func (o *Options) adoptSourceWiFi(src *shelly.OnboardWiFiConfig) {
	if src == nil || o.SSID != "" {
		return
	}
	o.SSID = src.SSID
	if o.Password == "" && !o.Open {
		o.Password = src.Password
		o.Open = src.Open
	}
	o.Factory.IOStreams().Info("Using WiFi credentials from source: %s", src.SSID)
}

// buildWiFiConfig assembles the WiFi config (including optional static IP) used
// to provision selected devices.
func (o *Options) buildWiFiConfig() *shelly.OnboardWiFiConfig {
	return &shelly.OnboardWiFiConfig{
		SSID:     o.SSID,
		Password: o.Password,
		Open:     o.Open,
		StaticIP: o.StaticIP,
		Gateway:  o.Gateway,
		Netmask:  o.Netmask,
		DNS:      o.DNS,
	}
}

// selectDevices resolves which discovered devices to provision. With --target-ap
// it returns the single device whose AP SSID matches (non-interactive); otherwise
// it runs the interactive multi-select (bypassed by --yes/--all).
func (o *Options) selectDevices(devices []shelly.OnboardDevice) ([]shelly.OnboardDevice, error) {
	if o.TargetAP == "" {
		return term.SelectOnboardDevices(o.Factory.IOStreams(), devices, o.Yes)
	}
	device, ok := shelly.FindByAP(devices, o.TargetAP)
	if !ok {
		return nil, fmt.Errorf("no unprovisioned device with AP SSID matching %q (%d discovered)", o.TargetAP, len(devices))
	}
	return []shelly.OnboardDevice{device}, nil
}

// outputDiscovered prints discoverable unprovisioned devices as JSON so the
// before/after scan-diff flow can identify a specific device's AP.
func (o *Options) outputDiscovered(devices []shelly.OnboardDevice) error {
	type apView struct {
		Name       string `json:"name"`
		SSID       string `json:"ssid"`
		Model      string `json:"model"`
		MAC        string `json:"mac"`
		Generation int    `json:"generation"`
		Source     string `json:"source"`
		Address    string `json:"address"`
	}
	views := make([]apView, 0, len(devices))
	for i := range devices {
		d := devices[i]
		views = append(views, apView{
			Name:       d.Name,
			SSID:       d.SSID,
			Model:      d.Model,
			MAC:        d.MACAddress,
			Generation: d.Generation,
			Source:     string(d.Source),
			Address:    d.Address,
		})
	}
	data, err := json.MarshalIndent(views, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal discovered devices: %w", err)
	}
	o.Factory.IOStreams().Println(string(data))
	return nil
}

// promptWiFiCredentials resolves WiFi credentials for provisioning. Tries in order:
// 1. Flags or source (--ssid/--password/--open) — used as given
// 2. Auto-detect from an existing Gen1 device on the network
// 3. This host's own network and stored passphrase
// 4. Interactive prompt.
//
// A known network with no password and no --open gets its password from this
// host's stored credentials or the prompt; without a terminal, or with --yes, it
// is left empty for provisioning to look up, and refused if none is found.
func (o *Options) promptWiFiCredentials(ctx context.Context) error {
	ios := o.Factory.IOStreams()
	if o.SSID == "" && !o.Open && o.detectWiFiCredentials(ctx) {
		return nil
	}

	if o.SSID == "" {
		ssid, err := ios.Input("WiFi SSID:", "")
		if err != nil {
			return fmt.Errorf("SSID input failed: %w", err)
		}
		if ssid == "" {
			return fmt.Errorf("WiFi SSID is required")
		}
		o.SSID = ssid
	}

	if o.Password == "" && !o.Open && ios.CanPrompt() && !o.Yes {
		if err := o.resolvePassword(ctx); err != nil {
			return err
		}
	}
	if o.Open {
		ios.Warning("No WiFi password provided for %q; configuring as an open network", o.SSID)
	}
	return nil
}

// detectWiFiCredentials adopts the credentials of a registered device's network,
// or of the network this host is on, and reports whether it found any.
func (o *Options) detectWiFiCredentials(ctx context.Context) bool {
	ios := o.Factory.IOStreams()
	svc := o.service()

	ios.StartProgress("Detecting WiFi credentials from existing devices...")
	creds := svc.GetWiFiCredentials(ctx)
	ios.StopProgress()

	if creds != nil {
		ios.Success("WiFi credentials detected from network: %s", creds.SSID)
		o.SSID = creds.SSID
		o.Password = creds.Password
		return true
	}

	// Recover the credentials of the network this host is already joined to. AP-hop
	// provisioning runs from a machine on the target WiFi, so the host holds both the
	// SSID and passphrase even when no Shelly device is registered yet — the same
	// recovery restore --to-ap performs.
	ios.StartProgress("Recovering WiFi credentials from this host...")
	hostCreds := svc.HostWiFiCredentials(ctx)
	ios.StopProgress()

	if hostCreds != nil {
		ios.Success("WiFi credentials recovered from host network: %s", hostCreds.SSID)
		o.SSID = hostCreds.SSID
		o.Password = hostCreds.Password
		return true
	}
	return false
}

// resolvePassword finds the password for the named network: this host's stored
// passphrase, else the prompt. An empty answer joins the network as an open one
// only when the user confirms it; declining refuses with the passphrase error.
func (o *Options) resolvePassword(ctx context.Context) error {
	pass, open, err := cmdutil.ResolveWiFiPassword(ctx, o.Factory.IOStreams(), o.service().HostWiFiPassword, o.SSID,
		cmdutil.WiFiPasswordPrompts{Password: o.askPassword, Confirm: o.askConfirm})
	if err != nil {
		return err
	}
	o.Password, o.Open = pass, open
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
		TargetAP:   o.TargetAP,
		DisableAP:  o.DisableAP,
	}
	if o.SSID != "" {
		onboardOpts.WiFi = o.buildWiFiConfig()
	}
	return onboardOpts
}

// runDiscovery runs multi-protocol device discovery with progress output.
func (o *Options) runDiscovery(ctx context.Context, svc provisionService, opts *shelly.OnboardOptions) ([]shelly.OnboardDevice, error) {
	ios := o.Factory.IOStreams()
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())
	if !opts.APOnly {
		mw.AddLine(methodBLE, "scanning...")
	}
	if !opts.BLEOnly {
		mw.AddLine(methodWiFiAP, "scanning...")
	}

	devices, err := svc.DiscoverForOnboard(ctx, opts, func(p shelly.OnboardProgress) {
		lineID := "network"
		switch p.Method {
		case methodBLE:
			lineID = methodBLE
		case methodWiFiAP:
			lineID = methodWiFiAP
		}

		switch {
		case p.Done && p.Err != nil:
			mw.UpdateLine(lineID, iostreams.StatusError, p.Err.Error())
		case p.Done:
			status := iostreams.StatusSuccess
			if p.Found == 0 {
				status = iostreams.StatusSkipped
			}
			mw.UpdateLine(lineID, status, fmt.Sprintf("%d found", p.Found))
		default:
			mw.UpdateLine(lineID, iostreams.StatusRunning, "scanning...")
		}
	})
	mw.Finalize()

	return devices, err
}

// provisionAll provisions devices grouped by their discovery source.
// If source is non-nil, applies the source config to each device after provisioning.
func (o *Options) provisionAll(
	ctx context.Context,
	svc provisionService,
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
			var r *shelly.OnboardResult
			// r carries the error too, and the results display reports it.
			if err := cmdutil.RunAtAP(ctx, ios, fmt.Sprintf("Onboarding %s at AP %s", dev.Name, dev.SSID),
				func(ctx context.Context) error {
					r = svc.OnboardViaAP(ctx, dev, wifiCfg, onboardOpts)
					return r.Error
				}); err != nil {
				ios.DebugErr("onboard "+dev.Name, err)
			}
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
	svc provisionService,
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
