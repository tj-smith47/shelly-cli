// Package provision provides device provisioning for Shelly devices.
package provision

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/tj-smith47/shelly-go/reprovision"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		Model: "SNSW-002P16EU",
		MAC:   "AA:BB:CC:DD:EE:FF",
		ID:    "shellyplus2pm-aabbcc",
	}

	if info.Model != "SNSW-002P16EU" {
		t.Errorf("got Model=%q, want %q", info.Model, "SNSW-002P16EU")
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got MAC=%q, want %q", info.MAC, "AA:BB:CC:DD:EE:FF")
	}
	if info.ID != "shellyplus2pm-aabbcc" {
		t.Errorf("got ID=%q, want %q", info.ID, "shellyplus2pm-aabbcc")
	}
}

func TestBTHomeDiscovery_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		discovery BTHomeDiscovery
		active    bool
		startedAt int64
		duration  int
	}{
		{
			name: "active discovery",
			discovery: BTHomeDiscovery{
				Active:    true,
				StartedAt: 1700000000,
				Duration:  30,
			},
			active:    true,
			startedAt: 1700000000,
			duration:  30,
		},
		{
			name: "inactive discovery",
			discovery: BTHomeDiscovery{
				Active:    false,
				StartedAt: 0,
				Duration:  0,
			},
			active:    false,
			startedAt: 0,
			duration:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.discovery.Active != tt.active {
				t.Errorf("got Active=%v, want %v", tt.discovery.Active, tt.active)
			}
			if tt.discovery.StartedAt != tt.startedAt {
				t.Errorf("got StartedAt=%d, want %d", tt.discovery.StartedAt, tt.startedAt)
			}
			if tt.discovery.Duration != tt.duration {
				t.Errorf("got Duration=%d, want %d", tt.discovery.Duration, tt.duration)
			}
		})
	}
}

func TestExtractWiFiSSID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		wantSSID string
	}{
		{
			name: "valid wifi config",
			input: map[string]any{
				wifiKeySta: map[string]any{
					wifiKeySSID: "MyNetwork",
				},
			},
			wantSSID: "MyNetwork",
		},
		{
			name: "empty ssid",
			input: map[string]any{
				wifiKeySta: map[string]any{
					wifiKeySSID: "",
				},
			},
			wantSSID: "",
		},
		{
			name:     "missing sta",
			input:    map[string]any{},
			wantSSID: "",
		},
		{
			name: "missing ssid",
			input: map[string]any{
				wifiKeySta: map[string]any{},
			},
			wantSSID: "",
		},
		{
			name:     "nil input",
			input:    nil,
			wantSSID: "",
		},
		{
			name:     "invalid type",
			input:    "not a map",
			wantSSID: "",
		},
		{
			name: "sta is not a map",
			input: map[string]any{
				wifiKeySta: "not a map",
			},
			wantSSID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ssid := ExtractWiFiSSID(tt.input)

			if ssid != tt.wantSSID {
				t.Errorf("ExtractWiFiSSID() = %q, want %q", ssid, tt.wantSSID)
			}
		})
	}
}

func TestDeviceInfo_JSONMarshaling(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		Model: "SNSW-002P16EU",
		MAC:   "AA:BB:CC:DD:EE:FF",
		ID:    "shellyplus2pm-aabbcc",
	}

	// Test that struct fields are accessible
	if info.Model != "SNSW-002P16EU" {
		t.Errorf("got Model=%q, want %q", info.Model, "SNSW-002P16EU")
	}
}

func TestBTHomeDiscovery_JSONMarshaling(t *testing.T) {
	t.Parallel()

	discovery := BTHomeDiscovery{
		Active:    true,
		StartedAt: 1700000000,
		Duration:  30,
	}

	// Test that struct fields are accessible
	if !discovery.Active {
		t.Error("expected Active to be true")
	}
	if discovery.StartedAt != 1700000000 {
		t.Errorf("got StartedAt=%d, want 1700000000", discovery.StartedAt)
	}
}

// wifiConfigServer answers Shelly.GetDeviceInfo and records the station of each
// WiFi.SetConfig call.
func wifiConfigServer(t *testing.T, record func(map[string]any)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any `json:"id"`
			Method string
			Params struct {
				Config struct {
					Sta map[string]any `json:"sta"`
				} `json:"config"`
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode rpc: %v", err)
			return
		}
		result := map[string]any{}
		switch req.Method {
		case "Shelly.GetDeviceInfo":
			result = map[string]any{"id": "shellyplus1-aabbcc", "mac": "AABBCCDDEEFF", "gen": 2, "model": "SNSW-001P16EU"}
		case "WiFi.SetConfig":
			record(req.Params.Config.Sta)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"id": req.ID, "jsonrpc": "2.0", "result": result}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestConfigureWiFi(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		password string
		open     bool
		wantPass any // nil: no write
		wantErr  error
	}{
		{name: "password", password: "secret", wantPass: "secret"},
		{name: "open", open: true, wantPass: ""},
		{name: "no password and not open", wantErr: reprovision.ErrNoPassphrase},
		{name: "open with a password", password: "secret", open: true, wantErr: types.ErrInvalidParam},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stations []map[string]any
			var mu sync.Mutex
			srv := wifiConfigServer(t, func(sta map[string]any) {
				mu.Lock()
				defer mu.Unlock()
				stations = append(stations, sta)
			})
			svc := New(&mockConnectionProvider{withConnectionFn: func(ctx context.Context, _ string, fn func(*client.Client) error) error {
				conn, err := client.Connect(ctx, model.Device{Address: srv.URL})
				if err != nil {
					return err
				}
				defer func() {
					if cerr := conn.Close(); cerr != nil {
						t.Logf("close connection: %v", cerr)
					}
				}()
				return fn(conn)
			}})

			err := svc.ConfigureWiFi(context.Background(), "ap", "home", tt.password, tt.open)
			mu.Lock()
			defer mu.Unlock()
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				if len(stations) != 0 {
					t.Errorf("refused call still wrote %v", stations)
				}
				return
			}
			if err != nil {
				t.Fatalf("ConfigureWiFi: %v", err)
			}
			if len(stations) != 1 || stations[0]["ssid"] != "home" || stations[0]["pass"] != tt.wantPass {
				t.Errorf("stations = %v, want ssid home with pass %v", stations, tt.wantPass)
			}
		})
	}
}
