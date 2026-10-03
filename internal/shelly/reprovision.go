package shelly

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/reprovision"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/netguard"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// APInspection is the state read from a device at its factory WiFi AP: its
// identity and, for Gen1, the WiFi station settings it would use to join a
// network.
type APInspection = reprovision.Inspection

// apFlows holds the factory access point operations the adapters below call.
type apFlows struct {
	restore func(context.Context, *reprovision.RestoreOptions) (*reprovision.RestoreResult, error)
	onboard func(context.Context, *reprovision.OnboardOptions) (*reprovision.OnboardResult, error)
	inspect func(context.Context, *reprovision.InspectOptions) (*reprovision.Inspection, error)
}

// defaultAPFlows runs the operations of the reprovision package.
var defaultAPFlows = &apFlows{
	restore: refusedUnderTest("reprovision.Restore", reprovision.Restore),
	onboard: refusedUnderTest("reprovision.Onboard", reprovision.Onboard),
	inspect: refusedUnderTest("reprovision.Inspect", reprovision.Inspect),
}

// refusedUnderTest wraps an access point flow so that under `go test` it
// returns netguard.ErrBlocked: the flow joins WiFi networks and dials through
// sockets the SDK builds, which no dialer guard sees.
func refusedUnderTest[O, R any](name string, flow func(context.Context, O) (R, error)) func(context.Context, O) (R, error) {
	return func(ctx context.Context, opts O) (R, error) {
		if err := netguard.Refuse(name); err != nil {
			var zero R
			return zero, err
		}
		return flow(ctx, opts)
	}
}

// flows returns the factory access point operations the service runs.
func (s *Service) flows() *apFlows {
	if s.ap != nil {
		return s.ap
	}
	return defaultAPFlows
}

// WithWiFiScanner sets the WiFi backend that restore --to-ap, migrate --to-ap,
// AP onboarding and AP inspection use to join and leave networks. Without it
// they use the platform scanner.
func WithWiFiScanner(scanner discovery.WiFiScanner) ServiceOption {
	return func(s *Service) {
		s.wifiScanner = scanner
	}
}

// stepReporterKey is the context key carrying a WithStepReporter callback.
type stepReporterKey struct{}

// WithStepReporter returns a context under which the factory access point
// operations call fn with a short description of each stage as it starts, such
// as "joining access point ShellyBulbDuo-D12965".
func WithStepReporter(ctx context.Context, fn func(step string)) context.Context {
	return context.WithValue(ctx, stepReporterKey{}, fn)
}

// StepReporter returns the WithStepReporter callback carried by ctx, or nil.
func StepReporter(ctx context.Context) func(string) {
	fn, ok := ctx.Value(stepReporterKey{}).(func(string))
	if !ok {
		return nil
	}
	return fn
}

// traceWriter writes each line it receives to the debug trace log.
type traceWriter struct{}

// Write records p as one debug trace event.
func (traceWriter) Write(p []byte) (int, error) {
	debug.TraceEvent("reprovision: %s", strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

// traceOptions keeps debug records and drops slog's time attribute, since the
// trace log stamps each event itself.
var traceOptions = &slog.HandlerOptions{
	Level: slog.LevelDebug,
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) == 0 && a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	},
}

// traceLogger sends the reprovision package's debug logs to the debug trace log.
var traceLogger = slog.New(slog.NewTextHandler(traceWriter{}, traceOptions))

// RestoreToAP restores a backup onto a device sitting at its factory WiFi AP. It
// hops the host onto the AP, writes the WiFi settings there, returns the host to
// its home network, waits for the device on the LAN and applies the full
// configuration at its LAN address. The device's registry entry registryName is
// pointed at that address.
//
// It returns the restore result and the device's LAN address. The address is
// also returned, and recorded, when the device joined the LAN but the full
// configuration restore there failed, so the restore can be finished by hand.
// A device that announced itself on the LAN but cannot be reached from this host
// is not recorded.
func (s *Service) RestoreToAP(
	ctx context.Context,
	apSSID string,
	apHostIP string,
	registryName string,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) (*backup.RestoreResult, string, error) {
	res, err := s.flows().restore(ctx, s.restoreOptions(ctx, apSSID, apHostIP, bkp, &opts))

	var result *backup.RestoreResult
	if res != nil && res.Restore != nil {
		result = backup.FromLibraryResult(res.Restore, bkp.Device().Generation, bkp.Backup, opts.SkipWebhooks)
	}
	addr := ""
	if res != nil && res.Reachable && res.Address != "" {
		addr = res.Address
		updateRegistryAddress(registryName, addr, bkp)
	}
	if err != nil {
		return result, addr, restoreAPError(err, restoreTarget{
			apSSID:   apSSID,
			name:     registryName,
			override: opts.NetworkOverride,
		}, res)
	}
	return result, addr, nil
}

// restoreOptions maps a CLI restore onto the reprovision options.
func (s *Service) restoreOptions(
	ctx context.Context,
	apSSID, apHostIP string,
	bkp *backup.DeviceBackup,
	opts *backup.RestoreOptions,
) *reprovision.RestoreOptions {
	ro := &reprovision.RestoreOptions{
		Scanner:                s.scanner(),
		StepTrace:              opts.StepTrace,
		Backup:                 bkp.Backup,
		Logger:                 traceLogger,
		OnStep:                 StepReporter(ctx),
		APSSID:                 apSSID,
		Name:                   backup.ResolveName(opts.Name, opts.AliasName, bkp.Device().MAC, shellyDeviceIDSuffix(apSSID)),
		APHostIP:               apHostIP,
		FirmwareURL:            opts.FirmwareURL,
		AllowFirmwareDowngrade: opts.AllowFirmwareDowngrade,
		// Both commands that restore at an access point may write a backup onto
		// a device other than the one it was taken from: backup restore of a
		// template, and migrate onto a new device.
		AllowForeignBackup: true,
		SkipAuth:           opts.SkipAuth,
		SkipScripts:        opts.SkipScripts,
		SkipSchedules:      opts.SkipSchedules,
		SkipKVS:            opts.SkipKVS,
		SkipWebhooks:       opts.SkipWebhooks,
		SkipState:          opts.SkipState,
		SkipMeters:         opts.SkipMeters,
	}
	ro.Network = opts.NetworkOverride.Network()
	return ro
}

// PlanAPStation describes the station a --to-ap restore of bkp writes, before
// any hop, on a dry run and a real one. A key from the flags or the backup is
// taken as is. Otherwise the passphrase comes from this host's stored
// credentials, for the network named by the flags or the backup or, when
// neither names one, for the network this host is on; with none stored it
// refuses with the PassphraseError the hop would return. It reads only the
// host's current network and stored passphrases: it never scans, joins or
// leaves a network, and never reads the device. --open, --password or
// --static-ip with no network named by --ssid or the backup is refused, as
// PlanLANStation refuses it.
func (s *Service) PlanAPStation(
	ctx context.Context, bkp *backup.DeviceBackup, ov *backup.NetworkOverride,
) (LANStationPlan, error) {
	fromBackup := reprovision.NetworkFromBackup(bkp.Backup)
	given := ov.Network()
	joined, err := reprovision.MergeNetwork(&fromBackup, &given)
	if err != nil {
		return LANStationPlan{}, backup.StaticNetworkError(err, ov)
	}
	plan := LANStationPlan{SSID: joined.SSID, Address: primaryAddress(bkp, given, joined)}
	if joined.SSID == "" {
		// The hop would join this host's current network with these flags,
		// which a LAN restore refuses; both paths refuse alike.
		if namesNetwork(ov) {
			return plan, errNoNetworkNamed()
		}
		// The hop joins the network this host is on; with none it fails on
		// the missing passphrase for an unnamed network.
		plan.SSID = s.hostCurrentSSID(ctx)
		if plan.SSID == "" {
			return plan, PassphraseError("")
		}
	} else if key, ok := stationKeyFromMerge(&given, &joined); ok {
		plan.Key = key
		return plan, nil
	}
	if _, err := s.hostPassphrase(ctx, plan.SSID); err != nil {
		return plan, err
	}
	plan.Key = StationKeyHost
	return plan, nil
}

// DescribeAPRestoreName returns the dry-run line for the name a --to-ap
// restore of bkp onto the device at access point apSSID writes. The device is
// identified by the MAC suffix in the access point name, the rule RestoreToAP
// applies.
func DescribeAPRestoreName(apSSID string, bkp *backup.DeviceBackup, opts backup.RestoreOptions) string {
	return backup.DescribeName(opts.Name, opts.AliasName, bkp.RecordedName(), bkp.Device().MAC,
		shellyDeviceIDSuffix(apSSID))
}

// PassphraseError reports a network to join for which no passphrase was given
// and none was found in this host's stored credentials, worded with the CLI's
// flags. errors.As reaches a *reprovision.NoPassphraseError naming ssid.
func PassphraseError(ssid string) error {
	return passphraseError(&reprovision.NoPassphraseError{SSID: ssid})
}

// passphraseError words a reprovision.ErrNoPassphrase error for the CLI, naming
// the network the SDK resolved.
func passphraseError(err error) error {
	var pwErr *reprovision.NoPassphraseError
	if !errors.As(err, &pwErr) || pwErr.SSID == "" {
		return &backup.CLIError{
			Msg: "no WiFi passphrase: no network was named and this host is not on a WiFi network — pass --ssid " +
				"with --password, or with --open for a network that has no password",
			Err: err,
		}
	}
	return &backup.CLIError{
		Msg: fmt.Sprintf("no WiFi passphrase for %q: Shelly devices return no station key and none was found in "+
			"this host's stored credentials — pass --password, or --open for a network that has no password",
			pwErr.SSID),
		Err: err,
	}
}

// restoreTarget describes what a --to-ap restore was aimed at, for its errors.
type restoreTarget struct {
	apSSID   string                  // the device's factory AP
	name     string                  // the target's registry name
	override *backup.NetworkOverride // the --ssid/--password/--open/--static-ip flags, when given
}

// restoreAPError words a reprovision.Restore error for the CLI. Errors whose SDK
// text names an SDK option, or that concern the target by name, are reworded;
// the rest carry the SDK text unchanged.
func restoreAPError(err error, t restoreTarget, res *reprovision.RestoreResult) error {
	apSSID, name := t.apSSID, t.name
	var fwErr *reprovision.FirmwareUnavailableError
	switch {
	case errors.Is(err, reprovision.ErrNoPassphrase):
		return passphraseError(err)
	case errors.Is(err, reprovision.ErrIncompleteStaticNetwork):
		return backup.StaticNetworkError(err, t.override)
	case errors.Is(err, reprovision.ErrNoRoute) && res != nil:
		return &backup.CLIError{
			Msg: fmt.Sprintf("restore applied at AP %q and %s rejoined the LAN at %s (seen via %s), but this host "+
				"has no route to it to write the full configuration — run the restore from a host on the device's "+
				"subnet", apSSID, name, res.Address, res.SeenVia),
			Err: err,
		}
	case errors.Is(err, reprovision.ErrNotRejoined):
		return &backup.CLIError{
			Msg: fmt.Sprintf("restore applied at AP %q but %s was not seen back on the LAN (%v); the device may "+
				"still be on its factory AP — if this host is not on the device's subnet, restore from one that is",
				apSSID, name, causeOf(err)),
			Err: err,
		}
	case res != nil && res.Reachable:
		return &backup.CLIError{
			Msg: fmt.Sprintf("%s joined the LAN at %s but the full configuration restore failed: %s",
				name, res.Address, restoreCause(causeOf(err), res)),
			Err: err,
		}
	case errors.As(err, &fwErr):
		return &backup.CLIError{
			Msg: fmt.Sprintf("AP hop for %q failed: device on firmware %q needs an update to the backup's %q "+
				"before restore, but no firmware image is available (the factory AP has no internet, so the image "+
				"is prefetched before the hop — its URL was underivable or the download failed); retry with "+
				"connectivity, pass --firmware-url, or --allow-firmware-downgrade to force the downgrade and "+
				"accept the reboot-loop risk", apSSID, fwErr.Current, fwErr.Required),
			Err: err,
		}
	case errors.Is(err, reprovision.ErrUnstable) && destabilizedStep(res) != "":
		return &backup.CLIError{
			Msg: fmt.Sprintf("restore at AP %q failed: %s", apSSID, restoreCause(err, res)),
			Err: err,
		}
	}
	return err
}

// causeOf returns the error err wraps, or err itself when it wraps none.
func causeOf(err error) error {
	if cause := errors.Unwrap(err); cause != nil {
		return cause
	}
	return err
}

// restoreCause words the cause of a failed restore pass. A step that drove the
// device into a reboot loop names --trace-file, where the SDK names its option.
func restoreCause(cause error, res *reprovision.RestoreResult) string {
	if step := destabilizedStep(res); step != "" && errors.Is(cause, reprovision.ErrUnstable) {
		return fmt.Sprintf("restore halted: device became unstable after the %q step — a write drove it "+
			"into a reboot loop; capture the per-step trace with --trace-file to confirm", step)
	}
	return cause.Error()
}

// destabilizedStep returns the restore step that drove the device into a reboot
// loop, or "".
func destabilizedStep(res *reprovision.RestoreResult) string {
	if res == nil || res.Restore == nil {
		return ""
	}
	return res.Restore.DestabilizedStep
}

// updateRegistryAddress points the named registry entry at the device's new LAN
// address, registering a fresh entry from the backup's device info when the name
// is not yet known. A name that already resolves to the new address is left
// untouched.
func updateRegistryAddress(name, addr string, bkp *backup.DeviceBackup) {
	if dev, ok := config.GetDevice(name); ok {
		if dev.Address == addr {
			return
		}
		if err := config.UpdateDeviceAddress(name, addr); err != nil {
			debug.TraceEvent("restore-to-ap: update address for %s: %v", name, err)
		}
		return
	}

	info := bkp.Device()
	if err := utils.RegisterDeviceFromModelCode(name, addr, info.Generation, info.Model, nil); err != nil {
		debug.TraceEvent("restore-to-ap: register %s: %v", name, err)
	}
}

// OnboardViaAP joins a device at its factory WiFi AP to the network in wifi: it
// hops the host onto the AP, writes the station settings, returns the host to
// its home network and looks for the device on the LAN. A device found there is
// added to the registry.
func (s *Service) OnboardViaAP(
	ctx context.Context,
	device *OnboardDevice,
	wifi *OnboardWiFiConfig,
	opts *OnboardOptions,
) *OnboardResult {
	result := &OnboardResult{Device: device, Method: string(OnboardSourceWiFiAP)}

	res, err := s.flows().onboard(ctx, &reprovision.OnboardOptions{
		Scanner: s.scanner(),
		Logger:  traceLogger,
		OnStep:  StepReporter(ctx),
		APSSID:  device.SSID,
		Network: reprovision.Network{
			SSID:     wifi.SSID,
			Password: wifi.Password,
			StaticIP: wifi.StaticIP,
			Gateway:  wifi.Gateway,
			Netmask:  wifi.Netmask,
			DNS:      wifi.DNS,
			Open:     wifi.Open,
		},
		Generation: device.Generation,
		DisableAP:  opts != nil && opts.DisableAP,
	})
	if err != nil {
		if errors.Is(err, reprovision.ErrNoPassphrase) {
			err = passphraseError(err)
		}
		result.Error = err
		return result
	}

	result.NewAddress = res.Address
	result.Note = res.Note
	result.APDisabled = res.APDisabled
	if res.Address == "" {
		return result
	}
	// The device's own answer at the AP beats the generation the scan guessed.
	registered := *device
	if res.Generation != 0 {
		registered.Generation = res.Generation
	}
	if regErr := RegisterOnboardedDevice(&registered, res.Address); regErr != nil {
		debug.TraceEvent("onboard register %s: %v", device.Name, regErr)
	} else {
		result.Registered = true
	}
	// The AP write carries only the station settings, so the name, timezone and
	// cloud choice reach the device over the LAN once it is there.
	if !res.Reachable {
		result.Note = joinNotes(result.Note, opts.describeUnapplied("the device announced itself but this host has no route to it"))
		return result
	}
	result.Note = joinNotes(result.Note, s.applyOnboardOptions(ctx, res.Address, opts))
	return result
}

// describeUnapplied names the device settings in o that an onboard did not
// apply, with the reason, or returns "" when o asks for none.
func (o *OnboardOptions) describeUnapplied(reason string) string {
	if o == nil {
		return ""
	}
	var asked []string
	if o.DeviceName != "" {
		asked = append(asked, "name")
	}
	if o.Timezone != "" {
		asked = append(asked, "timezone")
	}
	if o.NoCloud {
		asked = append(asked, "cloud")
	}
	if len(asked) == 0 {
		return ""
	}
	return fmt.Sprintf("%s not applied: %s", strings.Join(asked, ", "), reason)
}

// applyOnboardOptions writes the name, timezone and cloud choice in opts to the
// device at identifier and returns a note naming whatever the device refused,
// or "" when everything asked for was applied.
func (s *Service) applyOnboardOptions(ctx context.Context, identifier string, opts *OnboardOptions) string {
	if opts == nil || (opts.DeviceName == "" && opts.Timezone == "" && !opts.NoCloud) {
		return ""
	}
	var refused []string
	err := s.WithDevice(ctx, identifier, func(dev *DeviceClient) error {
		apply := func(what string, fn func() error) {
			if ferr := fn(); ferr != nil {
				refused = append(refused, fmt.Sprintf("%s: %v", what, ferr))
			}
		}
		if dev.IsGen1() {
			g1 := dev.Gen1().Device()
			if opts.DeviceName != "" {
				apply("name", func() error { return g1.SetName(ctx, opts.DeviceName) })
			}
			if opts.Timezone != "" {
				apply("timezone", func() error { return g1.SetTimezone(ctx, opts.Timezone) })
			}
			if opts.NoCloud {
				apply("cloud", func() error { return g1.SetCloud(ctx, false) })
			}
			return nil
		}
		rpcClient := dev.Gen2().RPCClient()
		if opts.DeviceName != "" {
			name := opts.DeviceName
			apply("name", func() error {
				return components.NewSys(rpcClient).SetConfig(ctx, &components.SysConfig{
					Device: &components.SysDeviceConfig{Name: &name},
				})
			})
		}
		if opts.Timezone != "" {
			tz := opts.Timezone
			apply("timezone", func() error {
				return components.NewSys(rpcClient).SetConfig(ctx, &components.SysConfig{
					Location: &components.SysLocationConfig{TZ: &tz},
				})
			})
		}
		if opts.NoCloud {
			enable := false
			apply("cloud", func() error {
				return components.NewCloud(rpcClient).SetConfig(ctx, &components.CloudConfig{Enable: &enable})
			})
		}
		return nil
	})
	if err != nil {
		return opts.describeUnapplied(err.Error())
	}
	if len(refused) == 0 {
		return ""
	}
	return "not applied: " + strings.Join(refused, "; ")
}

// joinNotes joins two result notes, skipping empty ones.
func joinNotes(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "; " + b
}

// InspectAtAP hops the host onto a device's factory WiFi AP, reads the device's
// identity and stored WiFi station settings, and returns the host to its home
// network. Use it to verify what a --to-ap restore or onboard actually wrote to a
// device that isn't reaching the LAN.
func (s *Service) InspectAtAP(ctx context.Context, apSSID, apHostIP string) (*APInspection, error) {
	insp, err := s.flows().inspect(ctx, &reprovision.InspectOptions{
		Scanner:  s.scanner(),
		Logger:   traceLogger,
		OnStep:   StepReporter(ctx),
		APSSID:   apSSID,
		APHostIP: apHostIP,
	})
	if err != nil {
		return nil, err
	}
	return insp, nil
}
