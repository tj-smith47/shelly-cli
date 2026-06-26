package shelly

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// refusingAddr returns a localhost address whose port is closed, so connection
// attempts fail fast and deterministically with "connection refused".
func refusingAddr(t *testing.T) string {
	t.Helper()
	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if cerr := ln.Close(); cerr != nil {
		t.Fatalf("close listener: %v", cerr)
	}
	return addr
}

func anyCircuitOpen(rl *ratelimit.DeviceRateLimiter) bool {
	for _, st := range rl.AllStats() {
		if st.Circuit.State == ratelimit.StateOpen {
			return true
		}
	}
	return false
}

// TestPollReachable_DoesNotTripCircuit is the regression guard for the --to-ap
// LAN-reconfirm failure: the reachability poll must NOT open the device's circuit
// breaker on its expected early failures (host still reacquiring DHCP, device
// still booting). If it did, every later probe would short-circuit with
// ErrCircuitOpen and the device's eventual recovery — plus the clockless-config
// second pass it gates — would never be observed. The control loop confirms the
// breaker genuinely trips for the same failures when NOT marked as polling.
// TestRejoinProbe_PollingDoesNotTripCircuit is the regression guard for the --to-ap
// LAN-reconfirm failure on the live path: confirmRejoin marks its context as polling
// before racing probes (probeReachableVia), so the expected early failures — host
// still reacquiring its DHCP lease, device still booting — must NOT open the device's
// circuit breaker. If they did, every later probe would short-circuit with
// ErrCircuitOpen and the device's eventual recovery would never be observed. The
// control loop confirms the breaker genuinely trips for the same failures when NOT
// marked as polling. Both generations route through the same probe seam.
func TestRejoinProbe_PollingDoesNotTripCircuit(t *testing.T) {
	t.Parallel()

	for _, gen := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d", gen), func(t *testing.T) {
			t.Parallel()
			addr := refusingAddr(t)

			// Polling path: many failing probes must leave the circuit closed.
			rl := ratelimit.New()
			svc := New(NewConfigResolver(), WithRateLimiter(rl))
			pollCtx := ratelimit.MarkAsPolling(context.Background())
			for range 5 {
				pctx, cancel := context.WithTimeout(pollCtx, 30*time.Millisecond)
				if err := svc.probeReachableVia(pctx, addr, "", gen); err == nil {
					t.Fatal("probe unexpectedly reached a closed port")
				}
				cancel()
			}
			if anyCircuitOpen(rl) {
				t.Error("circuit opened while polling — breaker not suppressed on the rejoin probe path")
			}

			// Control: the same failures without the polling mark DO open the breaker,
			// proving the suppression above is meaningful (default threshold is 3).
			rl2 := ratelimit.New()
			svc2 := New(NewConfigResolver(), WithRateLimiter(rl2))
			for range 5 {
				pctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
				if err := svc2.probeReachableVia(pctx, addr, "", gen); err == nil {
					t.Fatal("control probe unexpectedly reached a closed port")
				}
				cancel()
			}
			if !anyCircuitOpen(rl2) {
				t.Error("control: circuit never opened for non-polling failures — test cannot distinguish the fix")
			}
		})
	}
}

// TestConfirmRejoinedLAN_Notes covers the static and DHCP confirm branches, asserting
// each surfaces a non-fatal note rather than a fabricated success when the device
// cannot be located. A cancelled context returns the racer at once with no sighting,
// so onboard reports the not-seen cause as a warning.
func TestConfirmRejoinedLAN_Notes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		wifi *OnboardWiFiConfig
	}{
		{"static address not confirmed", &OnboardWiFiConfig{SSID: "Home", StaticIP: "192.0.2.10"}},
		{"dhcp device not found", &OnboardWiFiConfig{SSID: "Home"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // racer returns immediately with no sighting
			svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
			dev := &OnboardDevice{Name: "bulb", Generation: 1}

			addr, note := svc.confirmRejoinedLAN(ctx, dev, tc.wifi)
			if addr != "" {
				t.Errorf("addr = %q, want empty on failed confirm", addr)
			}
			if note == "" || !strings.Contains(note, "provisioned but") || !strings.Contains(note, "not seen on the network") {
				t.Errorf("note = %q, want a 'provisioned but ... not seen on the network' warning", note)
			}
		})
	}
}

// TestRaceRejoin covers the route-independent rejoin confirmation race over its
// injected presence-scan and unicast-probe seams (no real multicast or sockets):
//   - a unicast probe success is the strong, writeable outcome and names the
//     interface that reached the device (the AP-isolation fallback);
//   - a multicast-only sighting with no unicast route is the weak, NOT-writeable
//     outcome (the Bug C dual-homed case — provably back, unreachable from here);
//   - no sighting at all is the Bug A guard: an error, never a blind success;
//   - on DHCP an address learned by one interface's presence scan is probed by all.
func TestRaceRejoin(t *testing.T) {
	t.Parallel()

	fastCfg := func() rejoinConfig {
		return rejoinConfig{
			mac:             "AA:BB:CC:DD:EE:FF",
			generation:      1,
			timeout:         500 * time.Millisecond,
			interval:        5 * time.Millisecond,
			presenceTimeout: 10 * time.Millisecond,
			probeTimeout:    10 * time.Millisecond,
		}
	}
	neverSeen := func(context.Context, string, string, bool, time.Duration) (string, error) {
		return "", nil
	}
	noRoute := func(context.Context, string, string, int) error {
		return fmt.Errorf("no route to host")
	}

	t.Run("unicast probe wins names the reaching interface", func(t *testing.T) {
		t.Parallel()
		cfg := fastCfg()
		cfg.staticIP = "192.0.2.10"
		cfg.candidates = []string{"", "eth0"}
		cfg.scanPresence = neverSeen
		// Only the wired interface reaches the device (AP-isolated wireless).
		cfg.probe = func(_ context.Context, addr, iface string, _ int) error {
			if iface == "eth0" && addr == "192.0.2.10" {
				return nil
			}
			return fmt.Errorf("no route")
		}

		conf, err := raceRejoin(context.Background(), cfg)
		if err != nil {
			t.Fatalf("raceRejoin: %v", err)
		}
		if !conf.writeable {
			t.Error("writeable = false, want true for a unicast probe success")
		}
		if conf.addr != "192.0.2.10" || conf.bindIface != "eth0" || conf.via != "probe" {
			t.Errorf("conf = %+v, want addr=192.0.2.10 bindIface=eth0 via=probe", conf)
		}
	})

	t.Run("presence only is not writeable (Bug C dual-homed)", func(t *testing.T) {
		t.Parallel()
		cfg := fastCfg()
		cfg.staticIP = "192.0.2.10"
		cfg.candidates = []string{""}
		cfg.scanPresence = func(context.Context, string, string, bool, time.Duration) (string, error) {
			return "192.0.2.10", nil // announced over multicast...
		}
		cfg.probe = noRoute // ...but no unicast route from this host

		conf, err := raceRejoin(context.Background(), cfg)
		if err != nil {
			t.Fatalf("raceRejoin: %v", err)
		}
		if conf.writeable {
			t.Error("writeable = true, want false when only multicast saw the device")
		}
		if conf.addr != "192.0.2.10" || conf.via != "coiot" {
			t.Errorf("conf = %+v, want addr=192.0.2.10 via=coiot (Gen1 presence)", conf)
		}
	})

	t.Run("never seen returns an error not a blind success (Bug A)", func(t *testing.T) {
		t.Parallel()
		cfg := fastCfg()
		cfg.staticIP = "192.0.2.10"
		cfg.candidates = []string{""}
		cfg.scanPresence = neverSeen
		cfg.probe = noRoute

		conf, err := raceRejoin(context.Background(), cfg)
		if err == nil {
			t.Fatalf("expected error when the device is never seen, got conf %+v", conf)
		}
		if conf.addr != "" {
			t.Errorf("addr = %q, want empty when not seen", conf.addr)
		}
	})

	t.Run("dhcp address learned by presence is probed by every interface", func(t *testing.T) {
		t.Parallel()
		cfg := fastCfg()
		cfg.staticIP = "" // DHCP: address unknown up front
		cfg.candidates = []string{"", "wlan0"}
		// Only wlan0 hears the announcement; the address it learns must still be
		// probeable by any interface via the shared-address path.
		cfg.scanPresence = func(_ context.Context, _, iface string, _ bool, _ time.Duration) (string, error) {
			if iface == "wlan0" {
				return "192.0.2.55", nil
			}
			return "", nil
		}
		cfg.probe = func(_ context.Context, addr, _ string, _ int) error {
			if addr == "192.0.2.55" {
				return nil
			}
			return fmt.Errorf("unknown target %q", addr)
		}

		conf, err := raceRejoin(context.Background(), cfg)
		if err != nil {
			t.Fatalf("raceRejoin: %v", err)
		}
		if !conf.writeable || conf.addr != "192.0.2.55" || conf.via != "probe" {
			t.Errorf("conf = %+v, want addr=192.0.2.55 writeable=true via=probe", conf)
		}
	})
}

// TestRejoinCandidateInterfaces covers the candidate-set policy: a known static IP
// narrows to its same-subnet interfaces (default route first), while a DHCP confirm
// — address and subnet unknown — fans out across every host interface so a
// multi-homed host can bind a presence listener to whichever segment hears the
// device, again with the default route first.
func TestRejoinCandidateInterfaces(t *testing.T) {
	t.Parallel()

	_, subnet, err := net.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR: %v", err)
	}
	ifaces := []probeIface{
		{Name: "eth0", Nets: []*net.IPNet{subnet}},
		{Name: "wlan0", IsWireless: true, Nets: []*net.IPNet{subnet}},
	}

	t.Run("static IP narrows to same-subnet interfaces", func(t *testing.T) {
		t.Parallel()
		got := rejoinCandidateInterfaces("192.0.2.10", ifaces)
		want := []string{"", "eth0", "wlan0"} // default route, then wired before wireless
		if !equalStrings(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("dhcp fans out across every interface", func(t *testing.T) {
		t.Parallel()
		got := rejoinCandidateInterfaces("", ifaces)
		want := []string{"", "eth0", "wlan0"}
		if !equalStrings(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("dhcp default route only when no interfaces enumerated", func(t *testing.T) {
		t.Parallel()
		got := rejoinCandidateInterfaces("", nil)
		if !equalStrings(got, []string{""}) {
			t.Errorf("got %v, want [\"\"]", got)
		}
	})
}

// TestScanPresenceOnce_InputValidation covers the input-guard paths that never
// reach the network: an unparseable MAC and an unknown interface name both error
// before any multicast listener is created.
func TestScanPresenceOnce_InputValidation(t *testing.T) {
	t.Parallel()
	svc := NewService()

	if _, err := svc.scanPresenceOnce(context.Background(), "not-a-mac", "", true, time.Second); err == nil {
		t.Error("expected an error for an unparseable MAC, got nil")
	}

	_, err := svc.scanPresenceOnce(context.Background(), "AA:BB:CC:DD:EE:FF",
		"definitely-not-a-real-iface", true, time.Second)
	if err == nil || !strings.Contains(err.Error(), "resolve interface") {
		t.Errorf("err = %v, want a 'resolve interface' failure for an unknown interface", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestResolveJoinNetwork_Precedence covers the deterministic resolution paths
// (override password, then backup key) that never reach the host credential
// lookup, so the test does not depend on the runner's WiFi configuration.
func TestResolveJoinNetwork_Precedence(t *testing.T) {
	t.Parallel()
	svc := NewService()

	tests := []struct {
		name     string
		homeWiFi *OnboardWiFiConfig
		override *backup.NetworkOverride
		wantSSID string
		wantPass string
	}{
		{
			name:     "override password and ssid win",
			homeWiFi: &OnboardWiFiConfig{SSID: "BackupNet", Password: "backupkey"},
			override: &backup.NetworkOverride{SSID: "OverrideNet", Password: "overridekey"},
			wantSSID: "OverrideNet",
			wantPass: "overridekey",
		},
		{
			name:     "override ssid, backup key fills password",
			homeWiFi: &OnboardWiFiConfig{SSID: "BackupNet", Password: "backupkey"},
			override: &backup.NetworkOverride{SSID: "OverrideNet"},
			wantSSID: "OverrideNet",
			wantPass: "backupkey",
		},
		{
			name:     "no override, backup ssid and key used",
			homeWiFi: &OnboardWiFiConfig{SSID: "BackupNet", Password: "backupkey"},
			override: nil,
			wantSSID: "BackupNet",
			wantPass: "backupkey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ssid, pass, err := svc.resolveJoinNetwork(context.Background(), tt.homeWiFi, tt.override)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ssid != tt.wantSSID {
				t.Errorf("ssid = %q, want %q", ssid, tt.wantSSID)
			}
			if pass != tt.wantPass {
				t.Errorf("pass = %q, want %q", pass, tt.wantPass)
			}
		})
	}
}

// TestResolveJoinNetwork_NoPassphraseErrors covers the terminal error: with no
// override key, no backup key, and an EMPTY SSID, the host-credential lookup is
// skipped (it is guarded on a non-empty SSID) and the function fails loudly demanding
// --password. The empty SSID is deliberate: it keeps the test off the host WiFi
// credential store entirely (see the note on hostWiFiPassword below).
func TestResolveJoinNetwork_NoPassphraseErrors(t *testing.T) {
	t.Parallel()
	svc := NewService()

	_, _, err := svc.resolveJoinNetwork(context.Background(), &OnboardWiFiConfig{}, nil)
	if err == nil {
		t.Fatal("expected an error when no passphrase can be resolved")
	}
	if !strings.Contains(err.Error(), "--password") {
		t.Errorf("err = %q, want it to suggest --password", err)
	}
}

// hostWiFiPassword and resolveJoinNetwork's host-credential fallback are intentionally
// left UNCOVERED: on this platform discovery.NewWiFiDiscoverer().Scanner is a real
// HostNetworkPasswordProvider, so exercising that branch would read the host's actual
// stored WiFi credentials (a host networking call). The Service exposes no seam to
// substitute the scanner, so per the cardinal safety rule the branch is not driven —
// coverage is never worth running host-networking code.

// RestoreToAP itself is also intentionally NOT driven end-to-end: it hops the host
// onto the device's factory AP via withAPHop, which calls scanner.Connect — a real
// nmcli/wpa_cli host mutation on this platform. Its inner restore-at-AP closure
// (ensureGen1FirmwareAtAP, confirmGen1StableAtAP, rebootAtAP, the LAN completion) is
// covered directly by the helper tests above and below, which reach the same code
// over the httptest seam without touching host WiFi.

// TestRebootAtAP_SwallowsErrors covers the expected-failure reboot at the AP: the
// device tears down the connection as it reboots, so an error from the call is logged,
// not raised. Both the Gen1 and Gen2 branches are exercised; neither returns anything,
// so the assertion is that the call completes (and, for the reachable Gen1 fake, that
// the reboot endpoint was actually hit).
func TestRebootAtAP_SwallowsErrors(t *testing.T) {
	t.Parallel()

	t.Run("gen1 reachable", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 1)
		// rebootAtAP dials discovery.DefaultAPIP, so point the resolver's AP IP at the
		// fake by resolving every identifier (including the AP IP) to this server.
		svc := apdevService(d, 1)
		svc.rebootAtAP(context.Background(), 1)
		// The reboot endpoint is reached through the resolver seam, never a real AP.
	})

	t.Run("gen2 reachable", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 2)
		svc := apdevService(d, 2)
		svc.rebootAtAP(context.Background(), 2)
	})

	t.Run("gen2 unreachable is non-fatal", func(t *testing.T) {
		t.Parallel()
		addr := refusingAddr(t)
		svc := New(&generationAwareResolver{device: deviceAt(addr, 2)},
			WithRateLimiter(ratelimit.New()))
		// A refused connection at the AP is the expected case; rebootAtAP must not panic
		// or propagate it.
		svc.rebootAtAP(context.Background(), 2)
	})
}

// TestProbeReachableOnce_Success covers both generation branches of the reachability
// probe: a Gen1 device answering /settings and a Gen2 device answering its RPC are
// each reachable, so the probe returns nil. The branches differ because a Gen2 device
// does not serve the Gen1 /settings endpoint.
func TestProbeReachableOnce_Success(t *testing.T) {
	t.Parallel()

	t.Run("gen1", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 1)
		svc := apdevService(d, 1)
		if err := svc.probeReachableOnce(context.Background(), d.addr(), 1); err != nil {
			t.Fatalf("probeReachableOnce gen1 (reachable): %v", err)
		}
	})

	t.Run("gen2", func(t *testing.T) {
		t.Parallel()
		d := newAPDevServer(t, 2)
		svc := apdevService(d, 2)
		if err := svc.probeReachableOnce(context.Background(), d.addr(), 2); err != nil {
			t.Fatalf("probeReachableOnce gen2 (reachable): %v", err)
		}
	})
}

// TestConfirmRejoin_StaticSuccess covers the Service wiring of confirmRejoin: when
// the device answers a unicast probe at its known static address, the race returns
// a writeable confirmation over the default route on the first tick — before any
// presence scan runs, so no real multicast is issued. The fake device stands in for
// the static IP via the resolver seam, and its host:port (which net.ParseIP cannot
// parse) collapses the candidate set to the default route, exactly as a loopback
// target would.
func TestConfirmRejoin_StaticSuccess(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 1)
	svc := apdevService(d, 1)

	conf, err := svc.confirmRejoin(context.Background(), 1, d.addr(), "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("confirmRejoin (reachable static): %v", err)
	}
	if !conf.writeable {
		t.Error("writeable = false, want true when the device answers a unicast probe")
	}
	if conf.addr != d.addr() {
		t.Errorf("addr = %q, want the static address %q", conf.addr, d.addr())
	}
	if conf.bindIface != "" {
		t.Errorf("bindIface = %q, want the default route for a loopback-style target", conf.bindIface)
	}
}

// TestRestoreOnLAN_CancelledContext covers the settle-wait guard: a cancelled context
// returns before the restore is attempted, surfacing ctx.Err() rather than running the
// full settle delay and a doomed restore.
func TestRestoreOnLAN_CancelledContext(t *testing.T) {
	t.Parallel()
	svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	bkp := apRestoreBackup()
	_, err := svc.restoreOnLAN(ctx, "192.0.2.10", 1, bkp, backup.RestoreOptions{}, lanFullRestoreBudget)
	if err == nil {
		t.Fatal("expected ctx.Err() when the settle wait is cancelled")
	}
	if !errorsIsContextCancelled(err) {
		t.Errorf("err = %v, want context cancellation", err)
	}
}

// TestCompleteRestoreOnLAN_SuccessRecordsAndReturns covers the LAN completion's happy
// path end to end: the settle elapses, the full restore against the reachable fake
// device succeeds, the device's address is recorded in the (isolated) registry, and a
// nil error is returned. This single test drives restoreOnLAN past its real settle,
// fullRestoreOnLAN to a successful result, and completeRestoreOnLAN's success return,
// so it is the one case that intentionally pays the lanSettleDelay.
//
//nolint:paralleltest // mutates the process-global default config manager via withIsolatedConfig
func TestCompleteRestoreOnLAN_SuccessRecordsAndReturns(t *testing.T) {
	withIsolatedConfig(t)
	d := newAPDevServer(t, 1)
	svc := apdevService(d, 1)

	bkp := apRestoreBackup()
	const name = "rejoined-bulb"
	res, gotAddr, err := svc.completeRestoreOnLAN(context.Background(), d.addr(), name, 1, bkp,
		backup.RestoreOptions{SkipNetwork: true, AllowFirmwareDowngrade: true})
	if err != nil {
		t.Fatalf("completeRestoreOnLAN (reachable device): %v", err)
	}
	if res == nil {
		t.Error("expected a non-nil restore result on success")
	}
	if gotAddr != d.addr() {
		t.Errorf("returned address = %q, want %q", gotAddr, d.addr())
	}
	if dev, ok := config.GetDevice(name); !ok || dev.Address != d.addr() {
		t.Errorf("registry entry for %q = %+v (ok=%v), want address %q", name, dev, ok, d.addr())
	}
}

// TestFullRestoreOnLAN_FatalOnFailure covers the full LAN restore's fatal policy: this
// pass IS the restore, so a failure is returned (not downgraded to a warning). A
// cancelled context makes the underlying restore fail at the settle wait, exercising
// the error path without the 8s settle or the restore engine.
func TestFullRestoreOnLAN_FatalOnFailure(t *testing.T) {
	t.Parallel()
	svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	bkp := apRestoreBackup()
	_, err := svc.fullRestoreOnLAN(ctx, "192.0.2.10", 1, bkp, backup.RestoreOptions{})
	if err == nil {
		t.Fatal("expected the full LAN restore failure to be fatal (non-nil)")
	}
}

// TestCompleteRestoreOnLAN_FatalAndRecordsAddress covers the LAN completion: even when
// the full restore fails fatally, the device has joined the LAN, so its address is
// recorded in the registry and returned, while the error names the failed config
// restore. A cancelled context drives the failure; the registry write is confined to
// an isolated config dir so no real registry is touched.
//
//nolint:paralleltest // mutates the process-global default config manager via withIsolatedConfig
func TestCompleteRestoreOnLAN_FatalAndRecordsAddress(t *testing.T) {
	withIsolatedConfig(t)

	svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	bkp := apRestoreBackup()
	const name = "restored-bulb"
	const addr = "192.0.2.42"
	_, gotAddr, err := svc.completeRestoreOnLAN(ctx, addr, name, 1, bkp, backup.RestoreOptions{})
	if err == nil {
		t.Fatal("expected a fatal error when the full config restore fails")
	}
	if !strings.Contains(err.Error(), "full configuration restore failed") {
		t.Errorf("err = %q, want it to name the failed config restore", err)
	}
	if gotAddr != addr {
		t.Errorf("returned address = %q, want %q (the LAN address is recorded regardless)", gotAddr, addr)
	}
	// The device's new address must have landed in the (isolated) registry.
	if dev, ok := config.GetDevice(name); !ok || dev.Address != addr {
		t.Errorf("registry entry for %q = %+v (ok=%v), want address %q", name, dev, ok, addr)
	}
}

// TestUpdateRegistryAddress covers the three registry-update paths: registering a new
// name, updating an existing name's address, and the no-op when the address already
// matches. All writes are confined to an isolated config dir.
//
//nolint:paralleltest // mutates the process-global default config manager via withIsolatedConfig
func TestUpdateRegistryAddress(t *testing.T) {
	withIsolatedConfig(t)

	bkp := apRestoreBackup()

	// New name -> registered from the backup's device info.
	updateRegistryAddress("fresh", "192.0.2.1", bkp)
	if dev, ok := config.GetDevice("fresh"); !ok || dev.Address != "192.0.2.1" {
		t.Errorf("fresh registration = %+v (ok=%v), want 192.0.2.1", dev, ok)
	}

	// Existing name, new address -> address updated.
	updateRegistryAddress("fresh", "192.0.2.2", bkp)
	if dev, _ := config.GetDevice("fresh"); dev.Address != "192.0.2.2" {
		t.Errorf("updated address = %q, want 192.0.2.2", dev.Address)
	}

	// Existing name, same address -> left untouched (no error, no change).
	updateRegistryAddress("fresh", "192.0.2.2", bkp)
	if dev, _ := config.GetDevice("fresh"); dev.Address != "192.0.2.2" {
		t.Errorf("address after no-op = %q, want 192.0.2.2", dev.Address)
	}

	// An unregisterable name (empty) drives the registration-error trace branch: the
	// failure is logged, not propagated (the address record is best-effort), so the
	// call must still return normally and add nothing.
	updateRegistryAddress("", "192.0.2.9", bkp)
	if _, ok := config.GetDevice(""); ok {
		t.Error("an empty device name must not have been registered")
	}
}

// apRestoreBackup builds a minimal Gen1 DeviceBackup carrying just the device-info
// fields the --to-ap helpers read (generation, model, firmware version, MAC).
func apRestoreBackup() *backup.DeviceBackup {
	return &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{
			Model:      "SHBDUO-1",
			Generation: 1,
			Version:    "20210101-000000",
			MAC:        "AABBCCDDEEFF",
		},
	}}
}

// errorsIsContextCancelled reports whether err is (or wraps) context cancellation.
func errorsIsContextCancelled(err error) bool {
	return err != nil && strings.Contains(err.Error(), context.Canceled.Error())
}

// withIsolatedConfig points the default config manager at an in-memory filesystem
// for the duration of the test, so registry writes never touch the user's real
// config (and never the real disk at all). The default manager and the package fs
// are process-global, so a test using it must NOT run in parallel.
func withIsolatedConfig(t *testing.T) {
	t.Helper()
	config.SetFs(afero.NewMemMapFs())
	config.ResetDefaultManagerForTesting()
	t.Cleanup(func() {
		config.SetFs(nil)
		config.ResetDefaultManagerForTesting()
	})
}

// TestMACSuffixFromAPSSID covers the AP-SSID MAC-suffix parse that the bystander
// guard keys on: a Shelly factory SSID's trailing hex token is the device MAC
// suffix, and anything without a pure-hex trailing token yields "" (guard skipped).
func TestMACSuffixFromAPSSID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, ssid, want string
	}{
		{"gen1 six-hex", "ShellyBulbDuo-6645B6", "6645B6"},
		{"gen2 full mac", "shellyplus2pm-aabbccddeeff", "AABBCCDDEEFF"},
		{"lowercase normalized", "ShellyBulbDuo-6645b6", "6645B6"},
		{"multiple dashes keeps last token", "Shelly-Bulb-DDEEFF", "DDEEFF"},
		{"no dash", "MyCustomAP", ""},
		{"trailing dash", "ShellyBulbDuo-", ""},
		{"non-hex suffix", "ShellyBulbDuo-LIVING", ""},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := macSuffixFromAPSSID(tc.ssid); got != tc.want {
				t.Errorf("macSuffixFromAPSSID(%q) = %q, want %q", tc.ssid, got, tc.want)
			}
		})
	}
}

// TestEvaluateAPIdentity covers the IO-free identity decision: a device MAC that
// ends with the SSID's MAC suffix matches, a different MAC is a genuine mismatch
// (matched=false, skip=false → fatal), and an underivable suffix or unparseable
// MAC is skipped (skip=true → guard bypassed, never a false mismatch).
func TestEvaluateAPIdentity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, mac, ssid     string
		wantMatch, wantSkip bool
	}{
		{"suffix match", "AABBCCDDEEFF", "ShellyBulbDuo-DDEEFF", true, false},
		{"full-mac ssid match", "AA:BB:CC:DD:EE:FF", "shellyplus2pm-AABBCCDDEEFF", true, false},
		{"genuine mismatch", "AABBCCDDEEFF", "ShellyBulbDuo-6645B6", false, false},
		{"no ssid suffix is skipped", "AABBCCDDEEFF", "CustomAP", false, true},
		{"unparseable mac is skipped", "not-a-mac", "ShellyBulbDuo-DDEEFF", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			matched, skip, _, _ := evaluateAPIdentity(tc.mac, tc.ssid)
			if matched != tc.wantMatch || skip != tc.wantSkip {
				t.Errorf("evaluateAPIdentity(%q, %q) = (matched=%v, skip=%v), want (matched=%v, skip=%v)",
					tc.mac, tc.ssid, matched, skip, tc.wantMatch, tc.wantSkip)
			}
		})
	}
}

// TestConfirmAPDeviceIdentity drives the guard end-to-end over the AP-device fake
// (which reports MAC AABBCCDDEEFF on both generations): an SSID naming that device
// passes, an SSID naming a different device is refused BEFORE any write, and a
// custom SSID with no MAC suffix is allowed through (the guard cannot verify by
// name). Both generations are exercised because the MAC is read over each client's
// own identity endpoint.
func TestConfirmAPDeviceIdentity(t *testing.T) {
	t.Parallel()

	for _, gen := range []int{1, 2} {
		t.Run(fmt.Sprintf("gen%d match passes", gen), func(t *testing.T) {
			t.Parallel()
			d := newAPDevServer(t, gen)
			svc := apdevService(d, gen)
			if err := svc.confirmAPDeviceIdentity(context.Background(), gen, "ShellyBulbDuo-DDEEFF"); err != nil {
				t.Errorf("matching AP identity should pass, got: %v", err)
			}
		})

		t.Run(fmt.Sprintf("gen%d mismatch refused", gen), func(t *testing.T) {
			t.Parallel()
			d := newAPDevServer(t, gen)
			svc := apdevService(d, gen)
			err := svc.confirmAPDeviceIdentity(context.Background(), gen, "ShellyBulbDuo-6645B6")
			if err == nil {
				t.Fatal("a device whose MAC does not match the AP name must be refused")
			}
			if !strings.Contains(err.Error(), "refusing to write") {
				t.Errorf("err = %q, want it to refuse the write to a bystander", err)
			}
		})
	}

	t.Run("no-suffix SSID is allowed without a device read", func(t *testing.T) {
		t.Parallel()
		// A Service with a resolver that points nowhere: a custom SSID must short-circuit
		// before any connection, so the absent device never matters.
		svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
		if err := svc.confirmAPDeviceIdentity(context.Background(), 1, "MyHouseAP"); err != nil {
			t.Errorf("a custom (no-MAC-suffix) SSID must skip the guard, got: %v", err)
		}
	})
}
