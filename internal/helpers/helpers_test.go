// Package helpers provides common functionality shared across CLI commands.
package helpers

import (
	"net"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDiscoveredDeviceToConfig_Basic(t *testing.T) {
	t.Parallel()
	d := discovery.DiscoveredDevice{
		ID:         "shelly-device-1",
		Address:    net.ParseIP("192.168.1.100"),
		Generation: 2,
		Model:      "SHSW-1",
	}

	cfg := DiscoveredDeviceToConfig(d)

	if cfg.Name != d.ID {
		t.Errorf("Name = %q, want %q", cfg.Name, d.ID)
	}
	if cfg.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want %q", cfg.Address, "192.168.1.100")
	}
	if cfg.Generation != int(d.Generation) {
		t.Errorf("Generation = %d, want %d", cfg.Generation, d.Generation)
	}
	if cfg.Model != d.Model {
		t.Errorf("Model = %q, want %q", cfg.Model, d.Model)
	}
}

func TestDiscoveredDeviceToConfig_WithName(t *testing.T) {
	t.Parallel()
	d := discovery.DiscoveredDevice{
		ID:         "shelly-device-1",
		Name:       "Living Room Light",
		Address:    net.ParseIP("192.168.1.101"),
		Generation: 2,
		Model:      "SHSW-1",
	}

	cfg := DiscoveredDeviceToConfig(d)

	// Name should be used instead of ID when available
	if cfg.Name != d.Name {
		t.Errorf("Name = %q, want %q", cfg.Name, d.Name)
	}
}

func TestDiscoveredDeviceToConfig_EmptyName(t *testing.T) {
	t.Parallel()
	d := discovery.DiscoveredDevice{
		ID:         "shelly-device-1",
		Name:       "",
		Address:    net.ParseIP("192.168.1.102"),
		Generation: 2,
		Model:      "SHSW-1",
	}

	cfg := DiscoveredDeviceToConfig(d)

	// Should fall back to ID when Name is empty
	if cfg.Name != d.ID {
		t.Errorf("Name = %q, want %q", cfg.Name, d.ID)
	}
}

// Note: DefaultTimeout is now defined in internal/shelly/shelly.go.
// Tests for the constant value are in the shelly package.

func TestUnmarshalJSON_ValidData(t *testing.T) {
	t.Parallel()
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Create input data as a map (similar to RPC response)
	input := map[string]any{
		"name":  "test",
		"value": 42,
	}

	var result TestStruct
	err := UnmarshalJSON(input, &result)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}
	if result.Value != 42 {
		t.Errorf("Value = %d, want %d", result.Value, 42)
	}
}

func TestUnmarshalJSON_NestedData(t *testing.T) {
	t.Parallel()
	type Inner struct {
		ID int `json:"id"`
	}
	type Outer struct {
		Inner Inner `json:"inner"`
	}

	input := map[string]any{
		"inner": map[string]any{
			"id": 123,
		},
	}

	var result Outer
	err := UnmarshalJSON(input, &result)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if result.Inner.ID != 123 {
		t.Errorf("Inner.ID = %d, want %d", result.Inner.ID, 123)
	}
}

func TestUnmarshalJSON_InvalidTarget(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"name": "test",
	}

	// Try to unmarshal into an incompatible type
	var result int
	err := UnmarshalJSON(input, &result)
	if err == nil {
		t.Error("UnmarshalJSON() expected error for incompatible type, got nil")
	}
}

func TestUnmarshalJSON_SliceData(t *testing.T) {
	t.Parallel()
	input := []any{
		map[string]any{"id": 1},
		map[string]any{"id": 2},
	}

	type Item struct {
		ID int `json:"id"`
	}

	var result []Item
	err := UnmarshalJSON(input, &result)
	if err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("result[0].ID = %d, want 1", result[0].ID)
	}
	if result[1].ID != 2 {
		t.Errorf("result[1].ID = %d, want 2", result[1].ID)
	}
}

func TestFormatAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		hasAuth bool
	}{
		{"with auth", true},
		{"no auth", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatAuth(tt.hasAuth)
			if got == "" {
				t.Error("FormatAuth returned empty string")
			}
		})
	}
}

func TestFormatGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		gen  int
	}{
		{"zero", 0},
		{"gen1", 1},
		{"gen2", 2},
		{"gen3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatGeneration(tt.gen)
			if got == "" {
				t.Errorf("FormatGeneration(%d) returned empty string", tt.gen)
			}
		})
	}
}

func TestFormatOnOff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		on   bool
	}{
		{"on", true},
		{"off", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatOnOff(tt.on)
			if got == "" {
				t.Errorf("FormatOnOff(%v) returned empty string", tt.on)
			}
		})
	}
}

func TestFormatState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state string
	}{
		{"open", "open"},
		{"closed", "closed"},
		{"opening", "opening"},
		{"closing", "closing"},
		{"stopped", "stopped"},
		{"idle", "idle"},
		{"calibrating", "calibrating"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatState(tt.state)
			if got == "" {
				t.Errorf("FormatState(%q) returned empty string", tt.state)
			}
		})
	}
}

func TestDiscoveredDeviceToConfig_AllFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    discovery.DiscoveredDevice
		want model.Device
	}{
		{
			name: "gen1 device",
			d: discovery.DiscoveredDevice{
				ID:         "shelly1-AABBCC",
				Address:    net.ParseIP("10.0.0.1"),
				Generation: 1,
				Model:      "SHSW-1",
			},
			want: model.Device{
				Name:       "shelly1-AABBCC",
				Address:    "10.0.0.1",
				Generation: 1,
				Type:       "SHSW-1",
				Model:      "SHSW-1",
			},
		},
		{
			name: "gen2 device",
			d: discovery.DiscoveredDevice{
				ID:         "shellypro1pm-AABBCC",
				Address:    net.ParseIP("10.0.0.2"),
				Generation: 2,
				Model:      "SPSW-001P16EU",
			},
			want: model.Device{
				Name:       "shellypro1pm-AABBCC",
				Address:    "10.0.0.2",
				Generation: 2,
				Type:       "SPSW-001P16EU",
				Model:      "SPSW-001P16EU",
			},
		},
		{
			name: "gen3 device",
			d: discovery.DiscoveredDevice{
				ID:         "shelly3pm-AABBCC",
				Address:    net.ParseIP("10.0.0.3"),
				Generation: 3,
				Model:      "S3PM-001PCEU16",
			},
			want: model.Device{
				Name:       "shelly3pm-AABBCC",
				Address:    "10.0.0.3",
				Generation: 3,
				Type:       "S3PM-001PCEU16",
				Model:      "S3PM-001PCEU16",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DiscoveredDeviceToConfig(tt.d)

			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Address != tt.want.Address {
				t.Errorf("Address = %q, want %q", got.Address, tt.want.Address)
			}
			if got.Generation != tt.want.Generation {
				t.Errorf("Generation = %d, want %d", got.Generation, tt.want.Generation)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %q, want %q", got.Model, tt.want.Model)
			}
		})
	}
}
