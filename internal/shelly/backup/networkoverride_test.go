// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-go/gen1"
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

func TestApplyGen1WiFiOverride(t *testing.T) {
	t.Parallel()
	t.Run("static ip with backup credentials preserved", func(t *testing.T) {
		t.Parallel()
		sta := &gen1.WiFiStaSettings{SSID: "OnyxCheetah4.7", Key: "secret", Ipv4Method: "dhcp"}
		applyGen1WiFiOverride(sta, &NetworkOverride{
			StaticIP: "10.23.47.221",
			Gateway:  "10.23.47.1",
			Netmask:  "255.255.254.0",
			DNS:      "10.23.47.1",
		})

		if !sta.Enabled {
			t.Error("station should be enabled")
		}
		if sta.SSID != "OnyxCheetah4.7" || sta.Key != "secret" {
			t.Errorf("credentials should be preserved, got SSID=%q Key=%q", sta.SSID, sta.Key)
		}
		if sta.Ipv4Method != "static" {
			t.Errorf("Ipv4Method = %q, want static", sta.Ipv4Method)
		}
		if sta.IP != "10.23.47.221" || sta.Gw != "10.23.47.1" || sta.Mask != "255.255.254.0" || sta.DNS != "10.23.47.1" {
			t.Errorf("static fields not applied: %+v", sta)
		}
	})

	t.Run("explicit ssid and password override", func(t *testing.T) {
		t.Parallel()
		sta := &gen1.WiFiStaSettings{SSID: "old", Key: "oldkey"}
		applyGen1WiFiOverride(sta, &NetworkOverride{SSID: "new", Password: "newkey", StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.255.255.0"})
		if sta.SSID != "new" || sta.Key != "newkey" {
			t.Errorf("credentials not overridden, got SSID=%q Key=%q", sta.SSID, sta.Key)
		}
	})

	t.Run("non-static override leaves ipv4 method untouched", func(t *testing.T) {
		t.Parallel()
		sta := &gen1.WiFiStaSettings{SSID: "net", Key: "k", Ipv4Method: "dhcp", IP: "10.0.0.9"}
		applyGen1WiFiOverride(sta, &NetworkOverride{SSID: "net2"})
		if sta.Ipv4Method != "dhcp" || sta.IP != "10.0.0.9" {
			t.Errorf("non-static override mutated addressing: %+v", sta)
		}
		if sta.SSID != "net2" {
			t.Errorf("ssid override not applied: %q", sta.SSID)
		}
	})
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
