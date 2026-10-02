package shelly

import (
	"context"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
)

// TestConfigStations_Payload asserts the Shelly.SetConfig body that template
// apply, provision --from-template, sync push, config import and SetConfig
// send for each station case.
func TestConfigStations_Payload(t *testing.T) {
	t.Parallel()
	type send func(svc *Service) ([]string, error)
	template := func(cfg map[string]any) send {
		return func(svc *Service) ([]string, error) {
			_, warnings, err := svc.ApplyTemplate(context.Background(), "apdev", cfg, false)
			return warnings, err
		}
	}
	fromTemplate := func(cfg map[string]any) send {
		return func(svc *Service) ([]string, error) {
			_, warnings, err := svc.applyTemplate(context.Background(), "apdev", cfg, false, stationsOmit)
			return warnings, err
		}
	}
	push := func(cfg map[string]any) send {
		return func(svc *Service) ([]string, error) {
			return svc.PushDeviceConfig(context.Background(), "apdev", cfg)
		}
	}
	importFile := func(cfg map[string]any) send {
		return func(svc *Service) ([]string, error) {
			_, warnings, err := svc.ImportConfig(context.Background(), "apdev", cfg, false)
			return warnings, err
		}
	}
	setConfig := func(cfg map[string]any) send {
		return func(svc *Service) ([]string, error) {
			return svc.SetConfig(context.Background(), "apdev", cfg)
		}
	}
	staticSta := func(ssid string) map[string]any {
		return map[string]any{"ssid": ssid, "enable": true, "is_open": false, "ipv4mode": "static",
			"ip": "10.0.0.50", "gw": "10.0.0.254", "netmask": "255.255.0.0", "nameserver": "1.1.1.1"}
	}
	withWiFi := func(wifi map[string]any) map[string]any {
		return map[string]any{"sys": map[string]any{"device": map[string]any{"name": "x"}}, "wifi": wifi}
	}
	withMAC := func(mac string, wifi map[string]any) map[string]any {
		return map[string]any{"sys": map[string]any{"device": map[string]any{"name": "x", "mac": mac}}, "wifi": wifi}
	}

	tests := []struct {
		name       string
		send       send
		deviceSSID string
		deviceSta1 string
		staStatic  string
		hostPass   map[string]string

		wantWiFi    map[string]any // nil: no wifi key in the write
		wantWarning string
	}{
		{name: "template, same network: station not written",
			send: template(withWiFi(map[string]any{"sta": staticSta("home")})), deviceSSID: "home"},
		{name: "template, other network with host key: address never copied",
			send: template(withWiFi(map[string]any{"sta": staticSta("guest")})), deviceSSID: "home",
			hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}}},
		{name: "template, other network without a key: left out with a warning",
			send:       template(withWiFi(map[string]any{"sta": staticSta("guest"), "ap": map[string]any{"enable": false}})),
			deviceSSID: "home", wantWiFi: map[string]any{"ap": map[string]any{"enable": false}},
			wantWarning: "shelly wifi set"},
		{name: "template, sta1 other network without a key",
			send:       template(withWiFi(map[string]any{"sta": staticSta("home"), "sta1": map[string]any{"ssid": "spare", "enable": true}})),
			deviceSSID: "home", deviceSta1: "old", wantWarning: "sta1"},
		{name: "template, sta1 with no network is not written",
			send:       template(withWiFi(map[string]any{"sta1": map[string]any{"ssid": nil, "enable": false}})),
			deviceSSID: "home", deviceSta1: "spare"},
		{name: "provision --from-template never writes a station",
			send:       fromTemplate(withWiFi(map[string]any{"sta": staticSta("guest"), "ap": map[string]any{"enable": false}})),
			deviceSSID: "home", hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"ap": map[string]any{"enable": false}}},
		{name: "sync, same network: written without a key, address kept",
			send: push(withWiFi(map[string]any{"sta": staticSta("home")})), deviceSSID: "home",
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "home", "enable": true, "ipv4mode": "static",
				"ip": "10.0.0.50", "gw": "10.0.0.254", "netmask": "255.255.0.0", "nameserver": "1.1.1.1"}}},
		{name: "sync, changed network with host key",
			send: push(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true}})), deviceSSID: "home",
			hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}}},
		{name: "sync, changed network without a key: left out",
			send: push(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true}})), deviceSSID: "home",
			wantWarning: "shelly wifi set"},
		{name: "sync, changed network recorded open",
			send:       push(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "is_open": true}})),
			deviceSSID: "home",
			wantWiFi:   map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": ""}}},
		{name: "template with a pass given: written, address never copied",
			send: template(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "pass": "typed", "ip": "10.0.0.50",
				"ipv4mode": "static"}})), deviceSSID: "home",
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "pass": "typed"}}},
		{name: "import, matching MAC, same network: written with its address",
			send: importFile(withMAC("aa:bb:cc:dd:ee:ff", map[string]any{"sta": staticSta("home")})), deviceSSID: "home",
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "home", "enable": true, "ipv4mode": "static",
				"ip": "10.0.0.50", "gw": "10.0.0.254", "netmask": "255.255.0.0", "nameserver": "1.1.1.1"}}},
		{name: "import, another MAC, other network with host key: address never copied",
			send: importFile(withMAC("112233445566", map[string]any{"sta": staticSta("guest")})), deviceSSID: "home",
			hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}}},
		{name: "import, another MAC, same network: station not written",
			send: importFile(withMAC("112233445566", map[string]any{"sta": staticSta("home")})), deviceSSID: "home"},
		{name: "import, no MAC, other network with host key: address never copied",
			send: importFile(withWiFi(map[string]any{"sta": staticSta("guest")})), deviceSSID: "home",
			hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}}},
		{name: "import, no MAC, other network without a key: left out",
			send: importFile(withWiFi(map[string]any{"sta": staticSta("guest")})), deviceSSID: "home",
			wantWarning: "shelly wifi set"},
		{name: "template, changed network on a static device: warned it keeps its address",
			send: template(withWiFi(map[string]any{"sta": staticSta("guest")})), deviceSSID: "home", staStatic: "10.0.0.9",
			hostPass:    map[string]string{"guest": "gp"},
			wantWiFi:    map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}},
			wantWarning: "keeps its static address 10.0.0.9"},
		{name: "import, no MAC, changed network on a static device: warned it keeps its address",
			send:       importFile(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true}})),
			deviceSSID: "home", staStatic: "10.0.0.9", hostPass: map[string]string{"guest": "gp"},
			wantWiFi:    map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}},
			wantWarning: "keeps its static address 10.0.0.9"},
		{name: "import, matching MAC, new static address on a static device: no warning",
			send:       importFile(withMAC("AABBCCDDEEFF", map[string]any{"sta": staticSta("guest")})),
			deviceSSID: "home", staStatic: "10.0.0.9", hostPass: map[string]string{"guest": "gp"},
			wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true, "pass": "gp", "ipv4mode": "static",
				"ip": "10.0.0.50", "gw": "10.0.0.254", "netmask": "255.255.0.0", "nameserver": "1.1.1.1"}}},
		{name: "SetConfig, same network: written without a key",
			send:       setConfig(withWiFi(map[string]any{"sta": map[string]any{"ssid": "home", "enable": true, "is_open": false}})),
			deviceSSID: "home", wantWiFi: map[string]any{"sta": map[string]any{"ssid": "home", "enable": true}}},
		{name: "SetConfig, changed network without a key: left out",
			send:       setConfig(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "enable": true}})),
			deviceSSID: "home", wantWarning: "shelly wifi set"},
		{name: "SetConfig, given pass",
			send:       setConfig(withWiFi(map[string]any{"sta": map[string]any{"ssid": "guest", "pass": "typed"}})),
			deviceSSID: "home", wantWiFi: map[string]any{"sta": map[string]any{"ssid": "guest", "pass": "typed"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newAPDevServer(t, 2)
			d.staSSID, d.sta1SSID, d.staStatic = tt.deviceSSID, tt.deviceSta1, tt.staStatic
			svc := New(d.resolver(2), WithRateLimiter(ratelimit.New()),
				WithWiFiScanner(&recordingScanner{passwords: tt.hostPass}))

			warnings, err := tt.send(svc)
			if err != nil {
				t.Fatalf("send error = %v", err)
			}
			joined := strings.Join(warnings, "\n")
			if (tt.wantWarning == "") != (len(warnings) == 0) || !strings.Contains(joined, tt.wantWarning) {
				t.Errorf("warnings = %q, want one naming %q", warnings, tt.wantWarning)
			}
			var sets []apdevWrite
			for _, w := range d.written() {
				if w.method == "Shelly.SetConfig" {
					sets = append(sets, w)
				}
			}
			if len(sets) != 1 {
				t.Fatalf("Shelly.SetConfig writes = %+v, want one", sets)
			}
			cfg, ok := sets[0].params["config"].(map[string]any)
			if !ok || cfg["sys"] == nil {
				t.Fatalf("config = %v, want the non-wifi sections kept", sets[0].params)
			}
			gotWiFi, hasWiFi := cfg["wifi"].(map[string]any)
			if tt.wantWiFi == nil {
				if hasWiFi {
					t.Errorf("wifi = %v, want no wifi section", gotWiFi)
				}
				return
			}
			if !wifiEqual(gotWiFi, tt.wantWiFi) {
				t.Errorf("wifi = %v, want %v", gotWiFi, tt.wantWiFi)
			}
		})
	}
}

func wifiEqual(got, want map[string]any) bool {
	if len(got) != len(want) {
		return false
	}
	for k, w := range want {
		g, ok1 := got[k].(map[string]any)
		wm, ok2 := w.(map[string]any)
		if !ok1 || !ok2 || !maps.Equal(g, wm) {
			return false
		}
	}
	return true
}

// TestApplyTemplate_DoesNotModifyTemplate guards the caller's template, which
// the TUI and the template store keep after the write.
func TestApplyTemplate_DoesNotModifyTemplate(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 2)
	d.staSSID = "home"
	svc := apdevService(d, 2)
	sta := map[string]any{"ssid": "home", "ip": "10.0.0.50", "ipv4mode": "static"}
	cfg := map[string]any{"wifi": map[string]any{"sta": sta}}
	if _, _, err := svc.ApplyTemplate(context.Background(), "apdev", cfg, false); err != nil {
		t.Fatalf("ApplyTemplate() error = %v", err)
	}
	if sta["ip"] != "10.0.0.50" || field[map[string]any](field[map[string]any](cfg, "wifi"), "sta") == nil {
		t.Errorf("template modified: %v", cfg)
	}
}

// TestImportConfig_DryRun asserts the dry run of a file from another device
// names the planned station, sends nothing and never shows the password.
func TestImportConfig_DryRun(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 2)
	d.staSSID = "home"
	svc := New(d.resolver(2), WithRateLimiter(ratelimit.New()),
		WithWiFiScanner(&recordingScanner{passwords: map[string]string{"guest": "s3cret-key"}}))
	cfg := map[string]any{"wifi": map[string]any{"sta": map[string]any{
		"ssid": "guest", "enable": true, "ipv4mode": "static", "ip": "10.0.0.50"}}}

	changes, warnings, err := svc.ImportConfig(context.Background(), "apdev", cfg, true)
	if err != nil {
		t.Fatalf("ImportConfig(dry run) error = %v", err)
	}
	joined := strings.Join(append(changes, warnings...), "\n")
	if !strings.Contains(joined, `wifi.sta: "guest", password sent (not shown), address unchanged`) {
		t.Errorf("dry run = %q, want the planned station line", joined)
	}
	if strings.Contains(joined, "s3cret-key") || strings.Contains(joined, "10.0.0.50") {
		t.Errorf("dry run = %q, shows the password or the copied address", joined)
	}
	if w := d.written(); len(w) != 0 {
		t.Errorf("dry run wrote %+v", w)
	}
}

// TestSetComponentConfig_RefusesStationKeys asserts config set refuses wifi
// station keys before any write and still sends the other wifi keys.
func TestSetComponentConfig_RefusesStationKeys(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"sta.ssid", "sta1.pass", "sta", "sta1"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			d := newAPDevServer(t, 2)
			d.staSSID = "home"
			_, err := apdevService(d, 2).SetComponentConfig(context.Background(), "apdev", "wifi", map[string]any{key: "X"})
			if !errors.Is(err, types.ErrInvalidParam) || !strings.Contains(err.Error(), "shelly wifi set") {
				t.Errorf("SetComponentConfig(%s) error = %v, want ErrInvalidParam naming shelly wifi set", key, err)
			}
			if w := d.written(); len(w) != 0 {
				t.Errorf("writes = %+v, want none", w)
			}
		})
	}
	d := newAPDevServer(t, 2)
	d.staSSID = "home"
	if _, err := apdevService(d, 2).SetComponentConfig(context.Background(), "apdev", "wifi",
		map[string]any{"ap": map[string]any{"enable": false}}); err != nil {
		t.Fatalf("SetComponentConfig(ap) error = %v", err)
	}
	if w := d.written(); len(w) != 1 || w[0].method != "Shelly.SetConfig" {
		t.Errorf("writes = %+v, want one Shelly.SetConfig", w)
	}
}
