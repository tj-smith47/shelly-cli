package shelly

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
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
func TestPollReachable_DoesNotTripCircuit(t *testing.T) {
	t.Parallel()
	addr := refusingAddr(t)

	// Polling path: many failing probes must leave the circuit closed.
	rl := ratelimit.New()
	svc := New(NewConfigResolver(), WithRateLimiter(rl))
	err := svc.pollReachable(context.Background(), addr, 1,
		400*time.Millisecond, 1*time.Millisecond, 30*time.Millisecond)
	if err == nil {
		t.Fatal("pollReachable to a closed port returned nil error")
	}
	if anyCircuitOpen(rl) {
		t.Error("circuit opened during polling poll — breaker not suppressed for polling")
	}

	// Control: the same failures via a non-polling connection DO open the breaker,
	// proving the suppression above is meaningful (default threshold is 3).
	rl2 := ratelimit.New()
	svc2 := New(NewConfigResolver(), WithRateLimiter(rl2))
	for range 5 {
		pctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		if err := svc2.WithGen1Connection(pctx, addr, func(conn *client.Gen1Client) error {
			_, e := conn.GetSettings(pctx)
			return e
		}); err == nil {
			t.Fatal("control probe unexpectedly reached a closed port")
		}
		cancel()
	}
	if !anyCircuitOpen(rl2) {
		t.Error("control: circuit never opened for non-polling failures — test cannot distinguish the fix")
	}
}

// TestPollReachable_Gen2Branch exercises the Gen2+ probe path (a Gen2 device does
// not serve the Gen1 /settings endpoint, so the probe must route by generation).
// It still must not trip the circuit while polling.
func TestPollReachable_Gen2Branch(t *testing.T) {
	t.Parallel()
	addr := refusingAddr(t)

	rl := ratelimit.New()
	svc := New(NewConfigResolver(), WithRateLimiter(rl))
	err := svc.pollReachable(context.Background(), addr, 2,
		200*time.Millisecond, 1*time.Millisecond, 30*time.Millisecond)
	if err == nil {
		t.Fatal("gen2 pollReachable to a closed port returned nil error")
	}
	if anyCircuitOpen(rl) {
		t.Error("circuit opened during gen2 polling poll")
	}
}

// TestConfirmRejoinedLAN_Notes covers the static (direct address poll) and DHCP
// (mDNS-by-MAC) branches, asserting each surfaces a non-fatal note rather than a
// fabricated success when the device cannot be located.
func TestConfirmRejoinedLAN_Notes(t *testing.T) {
	t.Parallel()

	t.Run("static address not confirmed", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // poll returns immediately
		svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
		dev := &OnboardDevice{Name: "bulb", Generation: 1}
		wifi := &OnboardWiFiConfig{SSID: "Home", StaticIP: "192.0.2.10"}

		addr, note := svc.confirmRejoinedLAN(ctx, dev, wifi)
		if addr != "" {
			t.Errorf("addr = %q, want empty on failed confirm", addr)
		}
		if note == "" || !strings.Contains(note, "not confirmed at 192.0.2.10") {
			t.Errorf("note = %q, want a 'not confirmed at' warning", note)
		}
	})

	t.Run("dhcp device not found", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))
		// Empty MAC skips the 2s mDNS scan, so the cancelled context returns at once.
		dev := &OnboardDevice{Name: "bulb", Generation: 1}
		wifi := &OnboardWiFiConfig{SSID: "Home"} // DHCP (no static IP)

		addr, note := svc.confirmRejoinedLAN(ctx, dev, wifi)
		if addr != "" {
			t.Errorf("addr = %q, want empty on failed detection", addr)
		}
		if note == "" || !strings.Contains(note, "not found on network") {
			t.Errorf("note = %q, want a 'not found on network' warning", note)
		}
	})
}

// TestLocateRejoinedDevice_ReturnsErrorWhenUnconfirmed is the Bug A regression
// guard: the shared --to-ap confirm core must return an ERROR (not a blind
// success) when the device cannot be located, so RestoreToAP fails loudly instead
// of reporting a stranded device as live. Both the static-address and DHCP-by-MAC
// branches are covered.
func TestLocateRejoinedDevice_ReturnsErrorWhenUnconfirmed(t *testing.T) {
	t.Parallel()

	t.Run("static address unreachable", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // poll returns immediately
		svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))

		addr, err := svc.locateRejoinedDevice(ctx, 1, "bulb", "192.0.2.10", "AA:BB:CC:DD:EE:FF")
		if err == nil {
			t.Fatal("expected error when static address is unreachable, got nil (blind success)")
		}
		if addr != "" {
			t.Errorf("addr = %q, want empty on failed confirm", addr)
		}
		if !strings.Contains(err.Error(), "not confirmed at 192.0.2.10") {
			t.Errorf("err = %q, want a 'not confirmed at 192.0.2.10' cause", err)
		}
	})

	t.Run("dhcp device not found", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		svc := New(NewConfigResolver(), WithRateLimiter(ratelimit.New()))

		// No static IP and empty MAC: the MAC scan is skipped and the cancelled
		// context returns at once with the not-found error.
		addr, err := svc.locateRejoinedDevice(ctx, 1, "bulb", "", "")
		if err == nil {
			t.Fatal("expected error when DHCP device is not found, got nil (blind success)")
		}
		if addr != "" {
			t.Errorf("addr = %q, want empty on failed detection", addr)
		}
		if !strings.Contains(err.Error(), "not found on network") {
			t.Errorf("err = %q, want a 'not found on network' cause", err)
		}
	})
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
