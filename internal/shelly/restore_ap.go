package shelly

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// RestoreToAP restores a backup onto a device sitting at its factory WiFi AP, in
// a single operation: it hops the host onto the device's open AP, applies the
// backup (with any network + name overrides) at discovery.DefaultAPIP, then
// returns the host to the home network. The restored WiFi station config — the
// backup's own credentials plus the network override's static IP — moves the
// device onto the LAN. With a static IP the device lands at a known address,
// which is polled for reachability and written into the registry entry named
// registryName. The backup's WiFi credentials double as the host's reconnect
// credentials, since the source device and the host share the home network.
//
// Returns the restore result and the device's confirmed LAN address. A --to-ap
// restore that cannot confirm the device rejoined the LAN returns an error rather
// than a blind success, since an unconfirmed device may be stranded on its AP.
func (s *Service) RestoreToAP(
	ctx context.Context,
	apSSID string,
	apHostIP string,
	registryName string,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) (*backup.RestoreResult, string, error) {
	// The source backup carries the home-network SSID (the same network the host
	// is joined to). Its key is usually masked, so the passphrase is resolved
	// separately below; the SSID alone still seeds the host's reconnect creds.
	homeWiFi := extractWiFiFromBackup(bkp)
	if homeWiFi == nil {
		homeWiFi = &OnboardWiFiConfig{}
	}

	// Resolve the network the device joins and the passphrase it needs (no Shelly
	// device returns its station key, so this falls back to the host's creds).
	joinSSID, joinPass, err := s.resolveJoinNetwork(ctx, homeWiFi, opts.NetworkOverride)
	if err != nil {
		return nil, "", err
	}

	// Apply the resolved credentials to the device's station config (written by
	// the restore via the override) and to the host's own reconnect creds.
	if opts.NetworkOverride == nil {
		opts.NetworkOverride = &backup.NetworkOverride{}
	}
	opts.NetworkOverride.SSID = joinSSID
	opts.NetworkOverride.Password = joinPass
	homeWiFi.SSID = joinSSID
	homeWiFi.Password = joinPass

	// At the AP the device carries no generation hint on its bare IP, and a Gen1
	// device does not answer the Gen2 RPC probe — so route by the backup's known
	// generation (the target is the same model as the source) instead of probing.
	generation := bkp.Device().Generation

	// The AP pass writes ONLY the station config (NetworkOnly) — just enough to move
	// the device off its factory AP and onto the LAN — and the FULL configuration is
	// then applied on the LAN. This split is non-negotiable for every --to-ap restore,
	// firmware update or not: a factory AP has no clock (the device rejects clock-gated
	// config like astronomical schedule rules) and, the instant the station config is
	// written, the device starts leaving the AP for the LAN — so any further write at
	// the AP races a device that is dropping the connection and is misread as a reboot
	// loop (the restore halts at, e.g., the "mqtt" step). On the LAN the device is
	// stable, has NTP time, and a reboot-to-join is not mistaken for a crash.
	//
	// A firmware update happens automatically AT THE AP when the device's build is
	// older than the backup's: a device on corrupt/older firmware reboot-loops the
	// instant station mode is active, so it can never stay up to OTA on the LAN. The
	// image is prefetched now (while the host still has internet — the factory AP has
	// none) and re-served from the host's AP-subnet address during the hop, before the
	// station write; the at-AP check then decides whether to actually flash.
	apOpts := opts
	apOpts.NetworkOnly = true
	// NetworkOnly already bypasses the clock-gated writes, but be explicit: the AP is
	// clockless, so the pre-schedule-write clock wait could never succeed there.
	apOpts.SkipClockWait = true
	fwPath := prefetchAPFirmware(ctx, generation, bkp, opts)
	if fwPath != "" {
		defer removeFirmwareTemp(fwPath)
	}

	var (
		result     *backup.RestoreResult
		restoreErr error
	)
	hopErr := s.withAPHop(ctx, apSSID, apHostIP, homeWiFi, func(ctx context.Context) error {
		// Before ANY write, prove the device answering at the AP is the one whose AP
		// the host joined. A --to-ap pass that wrote to a device it was not aimed at
		// would strand a bystander (the documented sl-bulb corruption), so abort
		// before the firmware/station writes if the identity does not match.
		if idErr := s.confirmAPDeviceIdentity(ctx, generation, apSSID); idErr != nil {
			restoreErr = idErr
			return idErr
		}
		// Flash the device (if its build is older than the backup's) while it is stable
		// at its AP, before the station config that would otherwise reboot-loop it onto
		// a dead LAN. A no-op when the firmware already matches.
		if generation == 1 {
			if fwErr := s.ensureGen1FirmwareAtAP(ctx, apFirmwareBindIP(apHostIP), fwPath, bkp.DeviceInfo.Version, opts.AllowFirmwareDowngrade); fwErr != nil {
				return fwErr
			}
			// Gate the station write — the one write whose reboot can strand the device —
			// on confirmed stability, while the device is still on its recoverable AP. A
			// device not holding a stable uptime here would be lost by the reboot, so abort
			// before the write rather than brick it. This runs on every Gen1 --to-ap restore,
			// flash or not, so the no-flash path (firmware already matched) is gated too.
			if stErr := s.confirmGen1StableAtAP(ctx); stErr != nil {
				return stErr
			}
		}
		result, restoreErr = s.RestoreBackupGen(ctx, discovery.DefaultAPIP, generation, bkp, apOpts)
		if restoreErr != nil {
			return restoreErr
		}
		// A Gen1 device persists its new station config but keeps serving its AP
		// until it reboots; reboot it now, while the host is still on the AP, so it
		// drops the AP and joins the LAN with the restored credentials. The reboot
		// call itself usually errors (the device tears down the connection
		// mid-response) — that is expected and non-fatal.
		s.rebootAtAP(ctx, generation)
		return nil
	})
	if restoreErr != nil {
		return nil, "", fmt.Errorf("restore at AP %q failed: %w", apSSID, restoreErr)
	}
	if hopErr != nil {
		// No restoreErr means the host never reached the AP.
		return nil, "", fmt.Errorf("AP hop for %q failed: %w", apSSID, hopErr)
	}

	// The device is now rebooting and joining the LAN with the restored station
	// config. A --to-ap restore must not report a success it cannot prove — an
	// unconfirmed device may have failed to leave its factory AP and be stranded — so
	// confirm it actually landed before returning, racing route-independent presence
	// signals (mDNS, plus CoIoT on Gen1) against unicast probes across every host
	// interface (see confirmRejoin). A confirmation failure is fatal.
	staticIP := ""
	if opts.NetworkOverride != nil {
		staticIP = opts.NetworkOverride.StaticIP
	}
	conf, confErr := s.confirmRejoin(ctx, generation, staticIP, bkp.Device().MAC)
	if confErr != nil {
		return result, "", fmt.Errorf(
			"restore applied at AP %q but %s was not seen back on the LAN (%w); the device may "+
				"still be on its factory AP — if this host is not on the device's subnet, restore from one that is",
			apSSID, registryName, confErr)
	}
	if !conf.writeable {
		// The device announced itself over multicast — it provably rejoined the LAN —
		// but no unicast route from this host reached it, so the full-configuration LAN
		// pass cannot be written from here. This is the dual-homed / AP-isolated case:
		// fail loud with an accurate cause rather than claim a restore that did not land.
		return result, "", fmt.Errorf(
			"restore applied at AP %q and %s rejoined the LAN at %s (seen via %s), but this host has no route to it "+
				"to write the full configuration — run the restore from a host on the device's subnet",
			apSSID, registryName, conf.addr, conf.via)
	}

	// Complete the configuration on the LAN. Pin the pass to the interface that
	// confirmed the device, so it lands even when the default route would be dropped
	// by AP client isolation.
	bindCtx := client.WithBindInterface(ctx, conf.bindIface)
	return s.completeRestoreOnLAN(bindCtx, conf.addr, registryName, generation, bkp, opts)
}

// completeRestoreOnLAN applies the full configuration once the device has rejoined
// the LAN at newAddr, where it is stable and has an NTP clock. The AP pass wrote
// only the station config (and flashed firmware, if requested), so this pass IS the
// restore — every setting group lands here. Its failure is therefore fatal, but the
// device has joined the LAN, so its address is recorded regardless.
func (s *Service) completeRestoreOnLAN(
	ctx context.Context,
	newAddr, registryName string,
	generation int,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) (*backup.RestoreResult, string, error) {
	lanResult, lanErr := s.fullRestoreOnLAN(ctx, newAddr, generation, bkp, opts)
	updateRegistryAddress(registryName, newAddr, bkp)
	if lanErr != nil {
		return lanResult, newAddr, fmt.Errorf(
			"%s joined the LAN at %s but the full configuration restore failed: %w",
			registryName, newAddr, lanErr)
	}
	return lanResult, newAddr, nil
}

// lanSettleDelay gives a freshly-joined device a moment to obtain NTP time before
// the LAN config re-apply, so time-dependent settings do not hit the same clock
// error they did at the AP.
const lanSettleDelay = 8 * time.Second

// lanFullRestoreBudget caps the LAN pass that applies the full configuration after
// the AP pass moved the device onto the network: it writes every setting group and
// the device can restart mid-pass (a restored mode write, or the reboot to join the
// network), so it is generous. Any firmware update already happened at the AP, so
// this is config-only — minutes, not the OTA's tens of minutes.
const lanFullRestoreBudget = 4 * time.Minute

// fullRestoreOnLAN applies the complete configuration at the device's LAN address
// after the AP pass flashed the firmware and wrote only the station config. Here —
// on matched firmware, with an NTP clock, and stable — RestoreBackupGen writes the
// whole backup in one pass. SkipNetwork is set because the station config already
// took at the AP and must not be disturbed; AllowFirmwareDowngrade is forced on so
// the shelly-go restore does not re-attempt a firmware update on the LAN (the flash
// already happened at the AP, and a LAN OTA would reboot-loop the device anyway).
// Unlike the clock-dependent re-apply, this pass IS the restore, so its failure is
// returned to the caller as fatal rather than downgraded to a warning.
func (s *Service) fullRestoreOnLAN(
	ctx context.Context,
	addr string,
	generation int,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
) (*backup.RestoreResult, error) {
	lanOpts := opts
	lanOpts.SkipNetwork = true            // station config already applied at the AP
	lanOpts.NetworkOnly = false           // the LAN pass applies the full configuration
	lanOpts.ClockDependentOnly = false    // not a re-apply: everything is written here
	lanOpts.AllowFirmwareDowngrade = true // firmware already handled at the AP; never re-OTA on the LAN
	return s.restoreOnLAN(ctx, addr, generation, bkp, lanOpts, lanFullRestoreBudget)
}

// restoreOnLAN runs a bounded LAN restore pass after a settle delay, returning the
// raw result and error. It is the shared core of the best-effort clock-dependent
// re-apply and the fatal post-firmware full restore; each caller sets its own
// failure policy.
func (s *Service) restoreOnLAN(
	ctx context.Context,
	addr string,
	generation int,
	bkp *backup.DeviceBackup,
	lanOpts backup.RestoreOptions,
	budget time.Duration,
) (*backup.RestoreResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(lanSettleDelay):
	}

	lanCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	return s.RestoreBackupGen(lanCtx, addr, generation, bkp, lanOpts)
}

// rebootAtAP reboots the device at its AP address so it applies the restored
// station config and joins the LAN. The call is expected to fail as the device
// tears down the connection while rebooting, so the error is logged, not raised.
func (s *Service) rebootAtAP(ctx context.Context, generation int) {
	var err error
	if generation == 1 {
		err = s.WithGen1Connection(ctx, discovery.DefaultAPIP, func(conn *client.Gen1Client) error {
			return conn.Reboot(ctx)
		})
	} else {
		err = s.WithConnection(ctx, discovery.DefaultAPIP, func(conn *client.Client) error {
			_, callErr := conn.Call(ctx, "Shelly.Reboot", nil)
			return callErr
		})
	}
	if err != nil {
		debug.TraceEvent("restore-to-ap: reboot at AP (expected as device drops connection): %v", err)
	}
}

// confirmAPDeviceIdentity verifies the device answering at the factory AP is the
// one whose AP the host was asked to join, BEFORE any write. A Shelly factory AP
// SSID embeds the device's MAC suffix (e.g. "ShellyBulbDuo-6645B6"); the device
// reports its full MAC. If the host associated with a different device's AP — a
// duplicate/rogue SSID, or a stale association left over from a fleet hop — the
// reported MAC will not end with the SSID's suffix, and the write is aborted so it
// cannot overwrite a device the operation was never aimed at. When the SSID carries
// no recognizable MAC suffix (a user-renamed AP), identity cannot be derived from
// the name and the guard is skipped rather than blocking a legitimate restore.
func (s *Service) confirmAPDeviceIdentity(ctx context.Context, generation int, apSSID string) error {
	// Skip the device read entirely when the SSID yields no MAC to check against —
	// a write to a user-renamed AP cannot be verified by name, so do not pay for an
	// extra round-trip that could not change the outcome.
	if macSuffixFromAPSSID(apSSID) == "" {
		debug.TraceEvent("restore-to-ap: AP %q carries no MAC suffix; skipping device-identity guard", apSSID)
		return nil
	}
	actual, err := s.readDeviceMACAtAP(ctx, generation)
	if err != nil {
		return fmt.Errorf("could not read device identity at AP %q to confirm the target before writing: %w", apSSID, err)
	}
	matched, skip, want, got := evaluateAPIdentity(actual, apSSID)
	if skip {
		debug.TraceEvent("restore-to-ap: device at AP %q returned unparseable MAC %q; skipping identity guard", apSSID, actual)
		return nil
	}
	if !matched {
		return fmt.Errorf(
			"AP %q is serving device %s, whose MAC does not match the AP name's suffix %q: refusing to write to "+
				"avoid corrupting a device this operation was not aimed at — confirm the AP name, and that the "+
				"intended device is the one currently in AP mode, then retry",
			apSSID, got, want)
	}
	return nil
}

// evaluateAPIdentity decides whether a device reporting actualMAC is the one a
// Shelly factory AP named apSSID belongs to. matched is the verdict; skip is true
// when identity cannot be derived (the SSID carries no MAC suffix, or the reported
// MAC is unparseable), in which case the caller bypasses the guard rather than
// failing a legitimate operation. want and got are the normalized SSID suffix and
// full device MAC, for the caller's diagnostics. It performs no IO so the whole
// decision table is unit-testable without a device.
func evaluateAPIdentity(actualMAC, apSSID string) (matched, skip bool, want, got string) {
	want = macSuffixFromAPSSID(apSSID)
	if want == "" {
		return false, true, "", ""
	}
	got = normalizeMAC(actualMAC)
	if got == "" {
		return false, true, want, ""
	}
	return strings.HasSuffix(got, want), false, want, got
}

// macSuffixFromAPSSID extracts the trailing MAC hex from a Shelly factory AP SSID
// (the token after the final '-', e.g. "ShellyBulbDuo-6645B6" -> "6645B6"),
// upper-cased. It returns "" when the trailing token is absent or not pure hex, so
// a user-renamed AP (no derivable MAC) disables the identity guard rather than
// failing it.
func macSuffixFromAPSSID(ssid string) string {
	idx := strings.LastIndex(ssid, "-")
	if idx < 0 || idx == len(ssid)-1 {
		return ""
	}
	suffix := strings.ToUpper(ssid[idx+1:])
	for _, c := range suffix {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return ""
		}
	}
	return suffix
}

// readDeviceMACAtAP reads the MAC of the device answering at the factory AP
// address. Both generations report it through the device-identity read the client
// performs when the connection is established (Gen1 /shelly, Gen2+ Shelly.GetDeviceInfo),
// so the MAC is already cached on the connection and needs no extra round-trip.
func (s *Service) readDeviceMACAtAP(ctx context.Context, generation int) (string, error) {
	var mac string
	if generation == 1 {
		err := s.WithGen1Connection(ctx, discovery.DefaultAPIP, func(conn *client.Gen1Client) error {
			mac = conn.Info().MAC
			return nil
		})
		return mac, err
	}
	err := s.WithConnection(ctx, discovery.DefaultAPIP, func(conn *client.Client) error {
		mac = conn.Info().MAC
		return nil
	})
	return mac, err
}

// resolveJoinNetwork determines the SSID and passphrase the target device will
// use to join the LAN. No Shelly device returns its station key (Gen1 masks it,
// Gen2+ makes it write-only), so the passphrase comes by precedence: an explicit
// override (--password), then the backup's own key (unmasked Gen1 only), then the
// host's stored credentials for that network — the host running the CLI is
// already joined to it, so it already holds the passphrase. Returns an error when
// none of these yields a passphrase.
func (s *Service) resolveJoinNetwork(
	ctx context.Context,
	homeWiFi *OnboardWiFiConfig,
	override *backup.NetworkOverride,
) (ssid, pass string, err error) {
	ssid = homeWiFi.SSID
	if override != nil && override.SSID != "" {
		ssid = override.SSID
	}
	if override != nil {
		pass = override.Password
	}
	if pass == "" {
		pass = homeWiFi.Password
	}
	if pass == "" && ssid != "" {
		if pw, lookupErr := s.hostWiFiPassword(ctx, ssid); lookupErr != nil {
			debug.TraceEvent("restore-to-ap: host passphrase for %q not recovered: %v", ssid, lookupErr)
		} else {
			debug.TraceEvent("restore-to-ap: recovered passphrase for %q from host credentials", ssid)
			pass = pw
		}
	}
	if pass == "" {
		return "", "", fmt.Errorf(
			"no WiFi passphrase for %q: Shelly devices return no station key and none was "+
				"found in this host's stored credentials — pass --password", ssid)
	}
	return ssid, pass, nil
}

// hostWiFiPassword recovers the passphrase the host has stored for ssid from the
// OS credential store, when the platform's WiFi scanner supports it. It lets a
// device join the network the host is already on without re-supplying a
// passphrase the host already holds; no Shelly device returns its station key.
func (s *Service) hostWiFiPassword(ctx context.Context, ssid string) (string, error) {
	scanner := discovery.NewWiFiDiscoverer().Scanner
	if scanner == nil {
		return "", fmt.Errorf("WiFi not supported on this platform")
	}
	provider, ok := scanner.(discovery.HostNetworkPasswordProvider)
	if !ok {
		return "", fmt.Errorf("host passphrase recovery not supported on this platform")
	}
	return provider.HostNetworkPassword(ctx, ssid)
}

// lanRejoinTimeout bounds how long a --to-ap flow waits for a device to reappear
// on the LAN at its known static address after the host returns from the AP. It
// must absorb the device's reboot and WiFi association AND the host reacquiring
// its own DHCP lease after hopping back, so it is generous.
const lanRejoinTimeout = 120 * time.Second

// lanRejoinPollInterval is the delay between reachability probes, and
// lanRejoinProbeTimeout bounds a single probe so one slow attempt (the SDK
// retries network errors with exponential backoff) cannot blow past the poll
// cadence and starve the remaining probes of the time budget.
const (
	lanRejoinPollInterval = 3 * time.Second
	lanRejoinProbeTimeout = 8 * time.Second
)

// probeReachableOnce runs a single generation-aware reachability probe against
// addr over whatever egress interface the context pins (see
// client.WithBindInterface). The probe is generation-aware (Gen1 /settings vs a
// Gen2+ RPC) because the freshly-provisioned target keeps the source's generation
// and a Gen2 device does not serve the Gen1 /settings endpoint.
func (s *Service) probeReachableOnce(ctx context.Context, addr string, generation int) error {
	if generation == 1 {
		return s.WithGen1Connection(ctx, addr, func(conn *client.Gen1Client) error {
			_, settingsErr := conn.GetSettings(ctx)
			return settingsErr
		})
	}
	return s.WithConnection(ctx, addr, func(conn *client.Client) error {
		_, callErr := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		return callErr
	})
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
