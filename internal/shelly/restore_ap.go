package shelly

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
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

	var (
		result     *backup.RestoreResult
		restoreErr error
	)
	hopErr := s.withAPHop(ctx, apSSID, apHostIP, homeWiFi, func(ctx context.Context) error {
		result, restoreErr = s.RestoreBackupGen(ctx, discovery.DefaultAPIP, generation, bkp, opts)
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
	// unconfirmed device may have failed to leave its factory AP and be stranded —
	// so confirm it actually landed before returning: poll its known static address,
	// or, on DHCP, locate it by MAC over mDNS. A confirmation failure is fatal.
	staticIP := ""
	if opts.NetworkOverride != nil {
		staticIP = opts.NetworkOverride.StaticIP
	}
	newAddr, bindIface, confErr := s.locateRejoinedDevice(ctx, generation, registryName, staticIP, bkp.Device().MAC)
	if confErr != nil {
		return result, "", fmt.Errorf(
			"restore applied at AP %q but %s could not be confirmed back on the LAN (%w); the device may "+
				"still be on its factory AP — if this host is not on the device's subnet, confirm from one that is",
			apSSID, registryName, confErr)
	}

	// On the LAN the device finally has a clock (NTP), so re-apply the config it
	// rejected at its clockless factory AP — notably Gen1 light config, which
	// returns "Timezone and time should be set" at the AP and would otherwise leave
	// colour temperature and other light settings at factory defaults. Pin the
	// re-apply to the interface that confirmed the device, so it lands even when
	// the default route to the device would be dropped by AP client isolation.
	result = s.reapplyConfigOnLAN(client.WithBindInterface(ctx, bindIface), newAddr, generation, bkp, opts, result)
	updateRegistryAddress(registryName, newAddr, bkp)

	return result, newAddr, nil
}

// lanSettleDelay gives a freshly-joined device a moment to obtain NTP time before
// the LAN config re-apply, so time-dependent settings do not hit the same clock
// error they did at the AP.
const lanSettleDelay = 8 * time.Second

// lanReapplyBudget caps the whole LAN re-apply. The device was just confirmed
// reachable, but it can drop again mid-pass (a restored setting restarts it, or
// AP-isolation flaps the path); without a bound each write would then burn the
// transport's full retry budget (30s × 3) and the pass could hang for minutes.
// The re-apply is best-effort, so exceeding this just records a warning.
const lanReapplyBudget = 90 * time.Second

// reapplyConfigOnLAN re-applies, at the device's LAN address once it has joined
// and obtained a clock, only the configuration the device rejected at its
// clockless factory AP — Gen1 light/colour-temperature config and captured light
// state. Everything else already took at the AP, so re-writing it would needlessly
// thrash the device (a redundant mode write can even restart it again); the pass
// is therefore scoped via ClockDependentOnly rather than re-running the whole
// restore. The LAN result supersedes the AP result; if the LAN pass fails, the AP
// result is returned with a warning so the caller still sees that the device landed.
func (s *Service) reapplyConfigOnLAN(
	ctx context.Context,
	addr string,
	generation int,
	bkp *backup.DeviceBackup,
	opts backup.RestoreOptions,
	apResult *backup.RestoreResult,
) *backup.RestoreResult {
	select {
	case <-ctx.Done():
		return apResult
	case <-time.After(lanSettleDelay):
	}

	lanOpts := opts
	lanOpts.SkipNetwork = true        // network is live; do not disturb the connection
	lanOpts.ClockDependentOnly = true // only the clock-gated config the AP rejected

	reapplyCtx, cancel := context.WithTimeout(ctx, lanReapplyBudget)
	defer cancel()
	lanResult, err := s.RestoreBackupGen(reapplyCtx, addr, generation, bkp, lanOpts)
	if err != nil {
		debug.TraceEvent("restore-to-ap: LAN config re-apply failed: %v", err)
		apResult.Warnings = append(apResult.Warnings,
			fmt.Sprintf("device joined %s but re-applying clock-dependent config on the LAN failed: %v", addr, err))
		return apResult
	}
	return lanResult
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

// locateRejoinedDevice confirms a device returned to the LAN after an AP hop and
// returns the address it answered on along with the host interface that reached it
// ("" = the default route). With staticIP set the address is known, so it is
// polled directly over the default route first, falling back to a same-subnet
// interface only if that fails — which happens on a network whose AP isolates
// wireless clients, where the device rejoined over the host's own wireless AP is
// reached station-to-station and dropped, while a wired interface still reaches it
// (see selectProbeBindInterfaces). On DHCP the address is unknown, so the device
// is located by MAC over mDNS, which this path cannot interface-bind, so the
// returned interface is "". The returned error is the raw cause, leaving the
// policy to the caller: restore/migrate --to-ap treat it as fatal (a restore must
// not claim a success it cannot prove), while onboard surfaces it as a non-fatal
// note. Both waits use lanRejoinTimeout so all flows confirm over the same window.
func (s *Service) locateRejoinedDevice(
	ctx context.Context,
	generation int,
	name, staticIP, mac string,
) (addr, bindIface string, err error) {
	if staticIP != "" {
		ifaces, ifErr := hostProbeIfaces()
		if ifErr != nil {
			debug.TraceEvent("restore-to-ap: host interface enumeration failed, using default route: %v", ifErr)
		}
		candidates := selectProbeBindInterfaces(net.ParseIP(staticIP), ifaces)
		win, pollErr := s.pollReachableVia(ctx, staticIP, generation, candidates,
			lanRejoinTimeout, lanRejoinPollInterval, lanRejoinProbeTimeout)
		if pollErr != nil {
			return "", "", fmt.Errorf("not confirmed at %s: %w", staticIP, pollErr)
		}
		return staticIP, win, nil
	}
	ip, dErr := s.WaitForDeviceOnNetwork(ctx, name, mac, lanRejoinTimeout)
	if dErr != nil {
		return "", "", fmt.Errorf("not found on network: %w", dErr)
	}
	return ip, "", nil
}

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

// pollReachableVia polls addr until a generation-aware probe succeeds over one of
// the candidate egress interfaces, returning the interface that reached it ("" =
// default route). Candidates are tried in order on every tick, so a working wired
// path is preferred but the device is still confirmed if only the default route
// reaches it. Each probe runs under its own probeTimeout.
//
// The poll context is marked as a polling request so its expected early failures
// — the host is still reacquiring its DHCP lease after returning from the AP and
// the device is still booting/associating — are NOT recorded against the device's
// circuit breaker. Without this the breaker opens after a few quick failures and
// every subsequent probe short-circuits with ErrCircuitOpen before the device, now
// actually up, is contacted again; the LAN reconfirm (and the clockless-config
// second pass it gates) would then never run even though the device is reachable.
func (s *Service) pollReachableVia(
	ctx context.Context,
	addr string,
	generation int,
	candidates []string,
	timeout, interval, probeTimeout time.Duration,
) (string, error) {
	ctx = ratelimit.MarkAsPolling(ctx)
	deadline := time.Now().Add(timeout)

	probe := func(bindIface string) error {
		pctx, cancel := context.WithTimeout(client.WithBindInterface(ctx, bindIface), probeTimeout)
		defer cancel()
		return s.probeReachableOnce(pctx, addr, generation)
	}

	var lastErr error
	for time.Now().Before(deadline) {
		for _, cand := range candidates {
			err := probe(cand)
			if err == nil {
				return cand, nil
			}
			lastErr = err
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("device at %s not reachable within %s", addr, timeout)
}

// pollReachable polls addr over the default route until a generation-aware probe
// succeeds or the timeout elapses. It is the single-interface form of
// pollReachableVia, kept for callers that need no egress selection.
func (s *Service) pollReachable(
	ctx context.Context,
	addr string,
	generation int,
	timeout, interval, probeTimeout time.Duration,
) error {
	_, err := s.pollReachableVia(ctx, addr, generation, []string{""}, timeout, interval, probeTimeout)
	return err
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
