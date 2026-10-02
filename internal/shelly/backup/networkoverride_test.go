// Package backup provides backup and restore operations for Shelly devices.
package backup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/types"
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
		}, Station1Write{})
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
		}, Station1Write{})
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

	t.Run("static ip alone takes the backup's gateway, netmask and DNS", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"home","enable":true,"ipv4mode":"static","ip":"10.0.0.9",` +
			`"gw":"10.0.0.1","netmask":"255.255.255.0","nameserver":"10.0.0.2"}}`)
		out, err := applyGen2WiFiOverride(in, &NetworkOverride{StaticIP: "10.0.0.5", Gateway: "10.0.0.254"}, Station1Write{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sta := staMap(t, out)
		if sta["ip"] != "10.0.0.5" || sta["gw"] != "10.0.0.254" || sta["netmask"] != "255.255.255.0" ||
			sta["nameserver"] != "10.0.0.2" {
			t.Errorf("want the given ip and gateway with the backup's netmask and DNS, got %v", sta)
		}
	})

	t.Run("static ip on a DHCP backup without a gateway is refused", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"home","enable":true,"ipv4mode":"dhcp"}}`)
		_, err := applyGen2WiFiOverride(in, &NetworkOverride{StaticIP: "10.0.0.5", Netmask: "255.255.255.0"}, Station1Write{})
		if !errors.Is(err, shellybackup.ErrIncompleteStaticNetwork) {
			t.Errorf("err = %v, want ErrIncompleteStaticNetwork", err)
		}
	})

	t.Run("open clears the passphrase and the is_open flag", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"home","pass":"old","is_open":true,"enable":true}}`)
		out, err := applyGen2WiFiOverride(in, &NetworkOverride{SSID: "guest", Open: true}, Station1Write{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sta := staMap(t, out)
		if sta["ssid"] != "guest" || sta["pass"] != "" {
			t.Errorf("want ssid guest with an empty passphrase, got %v", sta)
		}
		if _, ok := sta["is_open"]; ok {
			t.Errorf("is_open kept: %v", sta)
		}
	})

	t.Run("a password drops a stale is_open flag", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"home","is_open":true,"enable":true}}`)
		out, err := applyGen2WiFiOverride(in, &NetworkOverride{Password: "pw"}, Station1Write{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sta := staMap(t, out)
		if _, ok := sta["is_open"]; ok || sta["pass"] != "pw" {
			t.Errorf("want pass pw and no is_open, got %v", sta)
		}
	})

	t.Run("is_open is never written", func(t *testing.T) {
		t.Parallel()
		in := json.RawMessage(`{"sta":{"ssid":"home","pass":"stale","is_open":true,"enable":true},"sta1":{"ssid":"b","is_open":false}}`)
		for _, ov := range []*NetworkOverride{nil, {SSID: "x"}, {Password: "pw"}, {SSID: "g", Open: true}, {StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.255.255.0"}} {
			out, err := applyGen2WiFiOverride(in, ov, Station1Write{})
			if err != nil {
				t.Fatalf("override %+v: %v", ov, err)
			}
			var cfg map[string]map[string]any
			if err := json.Unmarshal(out, &cfg); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			for key, station := range cfg {
				if _, ok := station["is_open"]; ok {
					t.Errorf("override %+v: %s carries is_open", ov, key)
				}
			}
			if pass, ok := cfg["sta"]["pass"]; ok && pass == "stale" {
				t.Errorf("override %+v: the backup's pass was written", ov)
			}
		}
	})

	t.Run("open with a password is refused", func(t *testing.T) {
		t.Parallel()
		_, err := applyGen2WiFiOverride(nil, &NetworkOverride{SSID: "guest", Password: "pw", Open: true}, Station1Write{})
		if !errors.Is(err, types.ErrInvalidParam) {
			t.Errorf("err = %v, want ErrInvalidParam", err)
		}
	})

	t.Run("invalid blob returns error", func(t *testing.T) {
		t.Parallel()
		if _, err := applyGen2WiFiOverride(json.RawMessage(`{bad`), &NetworkOverride{StaticIP: "10.0.0.5"}, Station1Write{}); err == nil {
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
			Open: true,
		},
	}
	out := toGen1RestoreOptions(in, in.Name)

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
		out.NetworkOverride.Netmask != "255.255.254.0" || out.NetworkOverride.DNS != "10.23.47.1" ||
		!out.NetworkOverride.Open {
		t.Errorf("NetworkOverride fields not translated: %+v", out.NetworkOverride)
	}
}

func TestToGen1RestoreOptions_NilOverride(t *testing.T) {
	t.Parallel()
	if out := toGen1RestoreOptions(RestoreOptions{}, ""); out.NetworkOverride != nil {
		t.Errorf("expected nil NetworkOverride, got %+v", out.NetworkOverride)
	}
}

func TestStaticNetworkError(t *testing.T) {
	t.Parallel()

	refused := fmt.Errorf("%w: static address 10.0.0.5 needs a gateway and a netmask, and the backup has none; "+
		"set Network.Gateway and Network.Netmask", shellybackup.ErrIncompleteStaticNetwork)
	other := errors.New("device unreachable")
	tests := []struct {
		name     string
		err      error
		override *NetworkOverride
		want     string
	}{
		{
			name: "--static-ip given", err: refused, override: &NetworkOverride{StaticIP: "10.0.0.5"},
			want: "static address 10.0.0.5 needs a gateway and a netmask, and the backup has none — pass --gateway and --netmask",
		},
		{name: "backup's own static address", err: refused, override: &NetworkOverride{SSID: "iot"}, want: MsgIncompleteBackupStatic},
		{name: "no override", err: refused, want: MsgIncompleteBackupStatic},
		{name: "other errors pass through", err: other, override: &NetworkOverride{StaticIP: "10.0.0.5"}, want: other.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StaticNetworkError(tt.err, tt.override)
			if got.Error() != tt.want {
				t.Errorf("message =\n%s\nwant\n%s", got, tt.want)
			}
			if !errors.Is(got, tt.err) {
				t.Error("errors.Is no longer reaches the original error")
			}
		})
	}
	if StaticNetworkError(nil, nil) != nil {
		t.Error("a nil error must stay nil")
	}
}

func TestDeviceBackup_StaticNetwork(t *testing.T) {
	t.Parallel()

	gen1Config := &DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
		Config: json.RawMessage(`{"wifi_sta":{"enabled":true,"ssid":"home","ipv4_method":"static",` +
			`"ip":"10.0.0.8","gw":"10.0.0.1","mask":"255.255.255.0","dns":"10.0.0.2"}}`),
	}}
	dhcp := &DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Generation: 2},
		WiFi:       json.RawMessage(`{"sta":{"ssid":"home","ipv4mode":"dhcp"}}`),
	}}
	tests := []struct {
		name     string
		bkp      *DeviceBackup
		override *NetworkOverride
		want     shellybackup.StaticNetwork
		wantErr  bool
	}{
		{
			name: "Gen1 settings fill a lone --static-ip", bkp: gen1Config,
			override: &NetworkOverride{StaticIP: "10.0.0.9"},
			want:     shellybackup.StaticNetwork{IP: "10.0.0.9", Gateway: "10.0.0.1", Netmask: "255.255.255.0", DNS: "10.0.0.2"},
		},
		{
			name: "no address keeps the backup's", bkp: gen1Config,
			want: shellybackup.StaticNetwork{IP: "10.0.0.8", Gateway: "10.0.0.1", Netmask: "255.255.255.0", DNS: "10.0.0.2"},
		},
		{name: "DHCP backup stays DHCP", bkp: dhcp, override: &NetworkOverride{SSID: "iot"}},
		{name: "DHCP backup cannot fill a lone --static-ip", bkp: dhcp, override: &NetworkOverride{StaticIP: "10.0.0.9"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.bkp.StaticNetwork(tt.override)
			if tt.wantErr {
				if !errors.Is(err, shellybackup.ErrIncompleteStaticNetwork) {
					t.Errorf("err = %v, want ErrIncompleteStaticNetwork", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("StaticNetwork() = %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}
