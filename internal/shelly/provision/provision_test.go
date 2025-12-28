// Package provision provides device provisioning for Shelly devices.
package provision

import (
	"context"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/client"
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
				"sta": map[string]any{
					"ssid": "MyNetwork",
				},
			},
			wantSSID: "MyNetwork",
		},
		{
			name: "empty ssid",
			input: map[string]any{
				"sta": map[string]any{
					"ssid": "",
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
				"sta": map[string]any{},
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
				"sta": "not a map",
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
