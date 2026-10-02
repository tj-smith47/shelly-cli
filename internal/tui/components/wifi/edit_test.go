package wifi

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// fakeWiFiDevice is a Gen2 device on the network "home" that records the
// station each WiFi.SetConfig writes.
type fakeWiFiDevice struct {
	// static, when set, is the station's static address.
	static string

	mu     sync.Mutex
	writes []map[string]any
}

func (d *fakeWiFiDevice) serve(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode rpc body: %v", err)
			return
		}
		var result any = map[string]any{}
		switch req.Method {
		case "Shelly.GetDeviceInfo":
			result = map[string]any{"id": "shellyplus1-aabbcc", "mac": "AABBCCDDEEFF", "gen": 2, "model": "SNSW-001P16EU"}
		case "WiFi.GetConfig", "Wifi.GetConfig":
			sta := map[string]any{"ssid": "home", "enable": true, "ipv4mode": "dhcp"}
			if d.static != "" {
				sta["ipv4mode"], sta["ip"], sta["gw"], sta["netmask"] = "static", d.static, "10.0.0.1", "255.255.255.0"
			}
			result = map[string]any{"sta": sta}
		case "WiFi.SetConfig", "Wifi.SetConfig":
			cfg, ok := req.Params["config"].(map[string]any)
			if !ok {
				t.Errorf("WiFi.SetConfig params = %v, want a config object", req.Params)
			}
			sta, ok := cfg["sta"].(map[string]any)
			if !ok {
				t.Errorf("WiFi.SetConfig config = %v, want a sta object", cfg)
			}
			d.mu.Lock()
			d.writes = append(d.writes, sta)
			d.mu.Unlock()
			result = map[string]any{"restart_required": false}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"id": req.ID, "result": result}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func (d *fakeWiFiDevice) stations() []map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]map[string]any(nil), d.writes...)
}

// hostKeys is a WiFi backend that knows this host's stored passphrases.
type hostKeys struct {
	shelly.OfflineWiFiScanner
	passwords map[string]string
}

func (h hostKeys) HostNetworkPassword(_ context.Context, ssid string) (string, error) {
	if p, ok := h.passwords[ssid]; ok {
		return p, nil
	}
	return "", errors.New("no stored passphrase")
}

// editOnHome shows the form for a fake device whose station is on "home",
// writing through the real Service.SetWiFiConfig.
func editOnHome(t *testing.T, d *fakeWiFiDevice, hostPass map[string]string) EditModel {
	t.Helper()
	resolver := &testutil.MockDeviceResolver{Device: model.Device{Name: "dev", Address: d.serve(t), Generation: 2}}
	svc := shelly.New(resolver, shelly.WithRateLimiter(ratelimit.New()),
		shelly.WithWiFiScanner(hostKeys{passwords: hostPass}))
	m := NewEditModel(context.Background(), svc)
	m = m.Show("dev", &network.WiFiConfigFull{STA: &network.WiFiStationFull{SSID: "home", Enabled: true}}, nil)
	m.setAP = func(context.Context, string, string, string, bool, bool) error { return nil }
	return m
}

func ctrlS() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl} }

// TestEditModel_StationPassword asserts the form's station write follows the
// rule of `shelly wifi set`: a blank password keeps the key on the same
// network, takes this host's stored passphrase on another one, and is refused
// when there is none.
func TestEditModel_StationPassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		ssid      string
		password  string
		open      bool
		hostPass  map[string]string
		wantWrite map[string]any // nil: refused, nothing written
		wantErr   string
	}{
		{name: "blank, same network keeps the key", ssid: "home",
			wantWrite: map[string]any{"ssid": "home", "enable": true}},
		{name: "blank, different network takes the host key", ssid: "guest", hostPass: map[string]string{"guest": "gp"},
			wantWrite: map[string]any{"ssid": "guest", "enable": true, "pass": "gp"}},
		{name: "blank, different network without a host key is refused", ssid: "guest", wantErr: "guest"},
		{name: "open toggle writes open", ssid: "guest", open: true,
			wantWrite: map[string]any{"ssid": "guest", "enable": true, "pass": ""}},
		{name: "password", ssid: "guest", password: "secret12",
			wantWrite: map[string]any{"ssid": "guest", "enable": true, "pass": "secret12"}},
		{name: "open with a password is refused", ssid: "guest", password: "secret12", open: true, wantErr: "open network"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := &fakeWiFiDevice{}
			m := editOnHome(t, d, tt.hostPass)
			m.staSSID.SetValue(tt.ssid)
			m.staPassword.SetValue(tt.password)
			if tt.open {
				m.Cursor = int(FieldOpen)
				m = m.toggleField()
			}

			m, cmd := m.Update(ctrlS())
			if cmd == nil || !m.Saving() {
				t.Fatalf("no save started (err %v)", m.Err)
			}
			res, ok := cmd().(messages.SaveResultMsg)
			if !ok {
				t.Fatal("save did not return a SaveResultMsg")
			}
			got := d.stations()
			if tt.wantWrite == nil {
				if res.Err == nil || !strings.Contains(res.Err.Error(), tt.wantErr) || len(got) != 0 {
					t.Errorf("result %v, writes %v; want a refusal naming %q and no write", res.Err, got, tt.wantErr)
				}
				return
			}
			if res.Err != nil {
				t.Fatalf("save error = %v", res.Err)
			}
			if len(got) != 1 || !maps.Equal(got[0], tt.wantWrite) {
				t.Errorf("station writes = %v, want %v", got, tt.wantWrite)
			}
		})
	}
}

// TestEditModel_StaticAddressWarning asserts a changed network on a
// statically addressed device saves and shows the warning that the address
// is kept.
func TestEditModel_StaticAddressWarning(t *testing.T) {
	t.Parallel()
	d := &fakeWiFiDevice{static: "10.0.0.9"}
	m := editOnHome(t, d, map[string]string{"guest": "gp"})
	m.staSSID.SetValue("guest")
	m, cmd := m.Update(ctrlS())
	if cmd == nil {
		t.Fatalf("no save started (err %v)", m.Err)
	}
	res, ok := cmd().(messages.SaveResultMsg)
	if !ok || res.Err != nil || !strings.Contains(res.Message, "keeps its static address 10.0.0.9") {
		t.Fatalf("result = %+v, want a saved result carrying the static address warning", res)
	}
	if _, toastCmd := m.Update(res); toastCmd == nil {
		t.Error("no toast for the warning")
	}
}

// apWrite is one recorded setAP call.
type apWrite struct {
	password string
	open     bool
}

func TestEditModel_APOpenToggle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		apOpen     bool // the device's current AP state
		toggle     bool
		password   string
		wantToggle bool     // the toggle's state when the form opens
		wantWrite  *apWrite // nil: refused in the form
	}{
		{name: "open AP starts open and stays open", apOpen: true, wantToggle: true, wantWrite: &apWrite{open: true}},
		{name: "secured AP starts secured and keeps its key", wantWrite: &apWrite{}},
		{name: "open AP turned off without a password is refused", apOpen: true, wantToggle: true, toggle: true},
		{name: "open AP secured with a password", apOpen: true, wantToggle: true, toggle: true, password: "secret12",
			wantWrite: &apWrite{password: "secret12"}},
		{name: "secured AP opened", toggle: true, wantWrite: &apWrite{open: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var writes []apWrite
			m := NewEditModel(context.Background(), &shelly.Service{})
			m = m.Show("dev", &network.WiFiConfigFull{
				STA: &network.WiFiStationFull{SSID: "home", Enabled: true, IsOpen: true},
				AP:  &network.WiFiAPFull{SSID: "shelly-ap", Enabled: true, IsOpen: tt.apOpen},
			}, nil)
			if m.staOpen {
				t.Error("the station toggle starts on; it is set only on request")
			}
			if m.apOpen != tt.wantToggle {
				t.Fatalf("AP toggle starts %v, want %v", m.apOpen, tt.wantToggle)
			}
			m.setStation = func(context.Context, string, network.StationWrite) ([]string, error) { return nil, nil }
			m.setAP = func(_ context.Context, _, _, password string, open, _ bool) error {
				writes = append(writes, apWrite{password: password, open: open})
				return nil
			}
			m.mode = EditModeAP
			if tt.toggle {
				m.Cursor = int(FieldOpen)
				m = m.toggleField()
			}
			m.apPassword.SetValue(tt.password)

			m, cmd := m.Update(ctrlS())
			if tt.wantWrite == nil {
				if cmd != nil || m.Err == nil {
					t.Fatalf("cmd %v err %v, want a refusal in the form", cmd != nil, m.Err)
				}
				return
			}
			if cmd == nil {
				t.Fatalf("no save started (err %v)", m.Err)
			}
			cmd()
			if len(writes) != 1 || writes[0] != *tt.wantWrite {
				t.Errorf("AP writes = %+v, want %+v", writes, *tt.wantWrite)
			}
		})
	}
}
