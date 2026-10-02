package shelly

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

// TestService_SetWiFiConfig_Payload asserts what reaches the device for each
// station write: the Gen1 /settings/sta query and the Gen2 WiFi.SetConfig sta.
func TestService_SetWiFiConfig_Payload(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		w          network.StationWrite
		deviceSSID string
		staStatic  string
		hostPass   map[string]string

		wantErr     error
		wantWarning string // "" expects no warning
		// gen1 is the expected /settings/sta query; gen2 the expected sta object.
		gen1 url.Values
		gen2 map[string]any
	}{
		{name: "given password", w: network.StationWrite{SSID: "home", Password: "pw", Enable: new(true)},
			deviceSSID: "old",
			gen1:       url.Values{"enabled": {"true"}, "ssid": {"home"}, "key": {"pw"}},
			gen2:       map[string]any{"ssid": "home", "pass": "pw", "enable": true}},
		{name: "same network keeps the key", w: network.StationWrite{SSID: "home", Enable: new(true)},
			deviceSSID: "home",
			gen1:       url.Values{"enabled": {"true"}, "ssid": {"home"}},
			gen2:       map[string]any{"ssid": "home", "enable": true}},
		{name: "changed network takes the host's key", w: network.StationWrite{SSID: "home", Enable: new(true)},
			deviceSSID: "old", hostPass: map[string]string{"home": "stored"},
			gen1: url.Values{"enabled": {"true"}, "ssid": {"home"}, "key": {"stored"}},
			gen2: map[string]any{"ssid": "home", "pass": "stored", "enable": true}},
		{name: "changed network without a key is refused", w: network.StationWrite{SSID: "home", Enable: new(true)},
			deviceSSID: "old", wantErr: reprovision.ErrNoPassphrase},
		{name: "open", w: network.StationWrite{SSID: "guest", Open: true, Enable: new(true)}, deviceSSID: "old",
			gen1: url.Values{"enabled": {"true"}, "ssid": {"guest"}, "key": {""}},
			gen2: map[string]any{"ssid": "guest", "pass": "", "enable": true}},
		{name: "static with every field", w: network.StationWrite{SSID: "home", Password: "pw", StaticIP: "10.0.0.9",
			Gateway: "10.0.0.254", Netmask: "255.255.0.0", DNS: "1.1.1.1"}, deviceSSID: "home",
			gen1: url.Values{"enabled": {"true"}, "ssid": {"home"}, "key": {"pw"}, "ipv4_method": {"static"},
				"ip": {"10.0.0.9"}, "gateway": {"10.0.0.254"}, "netmask": {"255.255.0.0"}, "dns": {"1.1.1.1"}},
			gen2: map[string]any{"ssid": "home", "pass": "pw", "ipv4mode": "static", "ip": "10.0.0.9",
				"gw": "10.0.0.254", "netmask": "255.255.0.0", "nameserver": "1.1.1.1"}},
		{name: "static keeps the device's gateway and netmask", w: network.StationWrite{SSID: "home", StaticIP: "10.0.0.9"},
			deviceSSID: "home", staStatic: "10.0.0.5",
			gen1: url.Values{"enabled": {"true"}, "ssid": {"home"}, "ipv4_method": {"static"},
				"ip": {"10.0.0.9"}, "gateway": {testGateway}, "netmask": {testNetmask}},
			gen2: map[string]any{"ssid": "home", "ipv4mode": "static", "ip": "10.0.0.9",
				"gw": testGateway, "netmask": testNetmask}},
		{name: "changed network keeps the device's static address: warned",
			w: network.StationWrite{SSID: "guest", Password: "gp"}, deviceSSID: "home", staStatic: "10.0.0.5",
			wantWarning: "keeps its static address 10.0.0.5",
			gen1:        url.Values{"enabled": {"true"}, "ssid": {"guest"}, "key": {"gp"}},
			gen2:        map[string]any{"ssid": "guest", "pass": "gp"}},
		{name: "changed network with a new static address: no warning",
			w:          network.StationWrite{SSID: "guest", Password: "gp", StaticIP: "10.0.0.9"},
			deviceSSID: "home", staStatic: "10.0.0.5",
			gen1: url.Values{"enabled": {"true"}, "ssid": {"guest"}, "key": {"gp"}, "ipv4_method": {"static"},
				"ip": {"10.0.0.9"}, "gateway": {testGateway}, "netmask": {testNetmask}},
			gen2: map[string]any{"ssid": "guest", "pass": "gp", "ipv4mode": "static", "ip": "10.0.0.9",
				"gw": testGateway, "netmask": testNetmask}},
		{name: "same network on a static device: no warning", w: network.StationWrite{SSID: "home"},
			deviceSSID: "home", staStatic: "10.0.0.5",
			gen1: url.Values{"enabled": {"true"}, "ssid": {"home"}},
			gen2: map[string]any{"ssid": "home"}},
		{name: "static on a DHCP device needs a gateway", w: network.StationWrite{SSID: "home", StaticIP: "10.0.0.9"},
			deviceSSID: "home", wantErr: types.ErrInvalidParam},
		{name: "disable", w: network.StationWrite{Enable: new(false)}, deviceSSID: "home",
			gen1: url.Values{"enabled": {"false"}},
			gen2: map[string]any{"enable": false}},
		{name: "open with a password", w: network.StationWrite{SSID: "home", Open: true, Password: "pw"},
			wantErr: types.ErrInvalidParam},
		{name: "password without an ssid", w: network.StationWrite{Password: "pw"}, wantErr: types.ErrInvalidParam},
		{name: "open without an ssid", w: network.StationWrite{Open: true}, wantErr: types.ErrInvalidParam},
		{name: "static without an ssid", w: network.StationWrite{StaticIP: "10.0.0.9", Gateway: testGateway,
			Netmask: testNetmask}, wantErr: types.ErrInvalidParam},
	}
	for _, gen := range []int{1, 2} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("gen%d/%s", gen, tt.name), func(t *testing.T) {
				t.Parallel()
				d := newAPDevServer(t, gen)
				d.staSSID, d.staStatic = tt.deviceSSID, tt.staStatic
				scanner := &recordingScanner{passwords: tt.hostPass}
				svc := New(d.resolver(gen), WithRateLimiter(ratelimit.New()), WithWiFiScanner(scanner))

				warnings, err := svc.SetWiFiConfig(context.Background(), "apdev", tt.w)
				writes := d.written()
				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("err = %v, want %v", err, tt.wantErr)
					}
					if len(writes) != 0 {
						t.Errorf("refused write reached the device: %+v", writes)
					}
					return
				}
				if err != nil {
					t.Fatalf("SetWiFiConfig() error = %v", err)
				}
				if got := strings.Join(warnings, "; "); (tt.wantWarning == "") != (got == "") ||
					!strings.Contains(got, tt.wantWarning) {
					t.Errorf("warnings = %q, want %q", got, tt.wantWarning)
				}
				if len(writes) != 1 {
					t.Fatalf("writes = %+v, want one", writes)
				}
				if gen == 1 {
					got := writes[0].query
					if writes[0].method != "/settings/sta" || !maps.EqualFunc(got, tt.gen1, slices.Equal) {
						t.Errorf("%s %v, want /settings/sta %v", writes[0].method, got, tt.gen1)
					}
					return
				}
				sta, ok := writes[0].params["config"].(map[string]any)["sta"].(map[string]any)
				if writes[0].method != "Wifi.SetConfig" || !ok || !maps.Equal(sta, tt.gen2) {
					t.Errorf("%s %v, want Wifi.SetConfig sta %v", writes[0].method, writes[0].params, tt.gen2)
				}
			})
		}
	}
}

// TestService_ProvisionDevice_Open checks that a bulk entry's open field
// reaches the device as an open network.
func TestService_ProvisionDevice_Open(t *testing.T) {
	t.Parallel()
	d := newAPDevServer(t, 2)
	d.staSSID = "old"
	svc := apdevService(d, 2)
	_, err := svc.ProvisionDevice(context.Background(), model.DeviceProvisionConfig{Name: "apdev"},
		&model.ProvisionWiFiConfig{SSID: "guest", Open: true})
	if err != nil {
		t.Fatalf("ProvisionDevice() error = %v", err)
	}
	writes := d.written()
	if len(writes) != 1 {
		t.Fatalf("writes = %+v, want one", writes)
	}
	sta, ok := writes[0].params["config"].(map[string]any)["sta"].(map[string]any)
	if !ok {
		t.Fatalf("write %+v has no sta", writes[0])
	}
	if pass, ok := sta["pass"]; !ok || pass != "" {
		t.Errorf("sta = %v, want an empty pass for an open network", sta)
	}
}

func TestValidateBulkProvisionConfig_OpenWithPassword(t *testing.T) {
	t.Parallel()
	cfg := &model.BulkProvisionConfig{
		WiFi: &model.ProvisionWiFiConfig{SSID: "home", Password: "pw", Open: true},
		Devices: []model.DeviceProvisionConfig{
			{Name: "porch", Address: "127.0.0.1", WiFi: &model.ProvisionWiFiConfig{SSID: "g", Password: "p", Open: true}},
			{Name: "hall", Address: "127.0.0.1", WiFi: &model.ProvisionWiFiConfig{SSID: "g", Open: true}},
		},
	}
	err := ValidateBulkProvisionConfig(cfg, func(string) bool { return true })
	if err == nil {
		t.Fatal("want an error for open with a password")
	}
	for _, want := range []string{"wifi: open and password", "porch: wifi open and password"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to contain %q", err, want)
		}
	}
	if strings.Contains(err.Error(), "hall") {
		t.Errorf("err = %v, an open entry without a password is valid", err)
	}
}
