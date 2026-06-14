// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNetworkOverride_IsStatic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ov   *NetworkOverride
		want bool
	}{
		{name: "nil", ov: nil, want: false},
		{name: "empty static ip", ov: &NetworkOverride{Gateway: "10.0.0.1"}, want: false},
		{name: "static ip set", ov: &NetworkOverride{StaticIP: "10.23.47.221"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ov.IsStatic(); got != tt.want {
				t.Errorf("IsStatic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyGen2WiFiOverride(t *testing.T) {
	t.Parallel()
	staMap := func(t *testing.T, blob json.RawMessage) map[string]any {
		t.Helper()
		var cfg map[string]any
		if err := json.Unmarshal(blob, &cfg); err != nil {
			t.Fatalf("unmarshal result: %v", err)
		}
		sta, ok := cfg["sta"].(map[string]any)
		if !ok {
			t.Fatalf("result has no sta object: %v", cfg)
		}
		return sta
	}

	t.Run("static ip preserves existing ssid", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"OnyxCheetah4.7","enable":true},"ap":{"enable":false}}`)
		out, err := applyGen2WiFiOverride(in, &NetworkOverride{
			StaticIP: "10.23.47.221",
			Gateway:  "10.23.47.1",
			Netmask:  "255.255.254.0",
			DNS:      "10.23.47.1",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sta := staMap(t, out)
		if sta["ssid"] != "OnyxCheetah4.7" {
			t.Errorf("ssid should be preserved, got %v", sta["ssid"])
		}
		if sta["ipv4mode"] != "static" || sta["ip"] != "10.23.47.221" || sta["netmask"] != "255.255.254.0" || sta["gw"] != "10.23.47.1" || sta["nameserver"] != "10.23.47.1" {
			t.Errorf("static fields not applied: %v", sta)
		}
		// The ap section must round-trip untouched.
		var cfg map[string]any
		if err := json.Unmarshal(out, &cfg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := cfg["ap"]; !ok {
			t.Error("ap section was dropped")
		}
	})

	t.Run("empty blob builds sta from scratch", func(t *testing.T) {
		t.Parallel()
		out, err := applyGen2WiFiOverride(nil, &NetworkOverride{
			SSID: "Net", Password: "pw", StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.255.255.0",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sta := staMap(t, out)
		if sta["ssid"] != "Net" || sta["pass"] != "pw" || sta["enable"] != true {
			t.Errorf("credentials not set: %v", sta)
		}
		if _, ok := sta["nameserver"]; ok {
			t.Error("nameserver should be absent when DNS is empty")
		}
	})

	t.Run("invalid blob returns error", func(t *testing.T) {
		t.Parallel()
		if _, err := applyGen2WiFiOverride(json.RawMessage(`{bad`), &NetworkOverride{StaticIP: "10.0.0.5"}); err == nil {
			t.Error("expected error for invalid WiFi blob")
		}
	})
}

// TestToGen1RestoreOptions guards the CLI→engine translation: every restore option
// the user can set must reach the shelly-go Gen1 restore unchanged. A silently
// dropped field here would, for example, make --allow-firmware-downgrade a no-op.
func TestToGen1RestoreOptions(t *testing.T) {
	t.Parallel()
	var trace bytes.Buffer
	in := RestoreOptions{
		Name:                   "FR",
		SkipNetwork:            true,
		SkipAuth:               true,
		SkipState:              true,
		SkipMeters:             true,
		SkipWebhooks:           true,
		ClockDependentOnly:     true,
		AllowFirmwareDowngrade: true,
		FirmwareURL:            "http://firmware.shelly.cloud/gen1/SHBDUO-1.zip",
		NetworkOnly:            true,
		SkipClockWait:          true,
		StepTrace:              &trace,
		NetworkOverride: &NetworkOverride{
			SSID: "Home", Password: "pw",
			StaticIP: "10.23.47.227", Gateway: "10.23.47.1",
			Netmask: "255.255.254.0", DNS: "10.23.47.1",
		},
	}
	out := toGen1RestoreOptions(in)

	if out.Name != in.Name ||
		out.SkipNetwork != in.SkipNetwork ||
		out.SkipAuth != in.SkipAuth ||
		out.SkipState != in.SkipState ||
		out.SkipMeters != in.SkipMeters ||
		out.SkipWebhooks != in.SkipWebhooks ||
		out.ClockDependentOnly != in.ClockDependentOnly ||
		out.AllowFirmwareDowngrade != in.AllowFirmwareDowngrade ||
		out.FirmwareURL != in.FirmwareURL ||
		out.NetworkOnly != in.NetworkOnly ||
		out.SkipClockWait != in.SkipClockWait {
		t.Errorf("scalar option dropped in translation: in=%+v out=%+v", in, out)
	}
	// StepTrace is the debug seam behind --trace-file; a dropped writer would
	// silently disable per-step tracing on a fragile device.
	if out.StepTrace != &trace {
		t.Error("StepTrace writer dropped in translation")
	}
	if out.NetworkOverride == nil {
		t.Fatal("NetworkOverride dropped in translation")
	}
	if out.NetworkOverride.SSID != "Home" || out.NetworkOverride.Password != "pw" ||
		out.NetworkOverride.StaticIP != "10.23.47.227" || out.NetworkOverride.Gateway != "10.23.47.1" ||
		out.NetworkOverride.Netmask != "255.255.254.0" || out.NetworkOverride.DNS != "10.23.47.1" {
		t.Errorf("NetworkOverride fields not translated: %+v", out.NetworkOverride)
	}
}

func TestToGen1RestoreOptions_NilOverride(t *testing.T) {
	t.Parallel()
	if out := toGen1RestoreOptions(RestoreOptions{}); out.NetworkOverride != nil {
		t.Errorf("expected nil NetworkOverride, got %+v", out.NetworkOverride)
	}
}
