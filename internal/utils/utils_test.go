// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

const testAddress = "192.168.1.100"

func TestDiscoveredDeviceToConfig_Basic(t *testing.T) {
	t.Parallel()
	d := discovery.DiscoveredDevice{
		ID:         "shelly-device-1",
		Address:    net.ParseIP(testAddress),
		Generation: 2,
		Model:      "SHSW-1",
	}

	cfg := DiscoveredDeviceToConfig(d)

	if cfg.Name != d.ID {
		t.Errorf("Name = %q, want %q", cfg.Name, d.ID)
	}
	if cfg.Address != testAddress {
		t.Errorf("Address = %q, want %q", cfg.Address, testAddress)
	}
	if cfg.Generation != int(d.Generation) {
		t.Errorf("Generation = %d, want %d", cfg.Generation, d.Generation)
	}
	// Type is the raw model code, Model is the human-readable name
	if cfg.Type != d.Model {
		t.Errorf("Type = %q, want %q", cfg.Type, d.Model)
	}
	wantModel := types.ModelDisplayName(d.Model)
	if cfg.Model != wantModel {
		t.Errorf("Model = %q, want %q", cfg.Model, wantModel)
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

func TestCalculateLatencyStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		latencies []time.Duration
		errors    int
		wantMin   time.Duration
		wantMax   time.Duration
	}{
		{
			name:      "empty latencies",
			latencies: []time.Duration{},
			errors:    5,
		},
		{
			name:      "single latency",
			latencies: []time.Duration{100 * time.Millisecond},
			errors:    0,
			wantMin:   100 * time.Millisecond,
			wantMax:   100 * time.Millisecond,
		},
		{
			name:      "multiple latencies",
			latencies: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 50 * time.Millisecond},
			errors:    1,
			wantMin:   50 * time.Millisecond,
			wantMax:   200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stats := CalculateLatencyStats(tt.latencies, tt.errors)

			if stats.Errors != tt.errors {
				t.Errorf("Errors = %d, want %d", stats.Errors, tt.errors)
			}
			if len(tt.latencies) > 0 {
				if stats.Min != tt.wantMin {
					t.Errorf("Min = %v, want %v", stats.Min, tt.wantMin)
				}
				if stats.Max != tt.wantMax {
					t.Errorf("Max = %v, want %v", stats.Max, tt.wantMax)
				}
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sorted []time.Duration
		p      int
		want   time.Duration
	}{
		{"empty slice", []time.Duration{}, 50, 0},
		{"single element", []time.Duration{100 * time.Millisecond}, 50, 100 * time.Millisecond},
		{"p50 of 10 elements", []time.Duration{
			10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond,
			60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond,
		}, 50, 60 * time.Millisecond},
		{"p95 of 10 elements", []time.Duration{
			10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond,
			60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond,
		}, 95, 100 * time.Millisecond},
		{"p99 of 10 elements", []time.Duration{
			10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond,
			60 * time.Millisecond, 70 * time.Millisecond, 80 * time.Millisecond, 90 * time.Millisecond, 100 * time.Millisecond,
		}, 99, 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Percentile(tt.sorted, tt.p)
			if got != tt.want {
				t.Errorf("Percentile(%v, %d) = %v, want %v", tt.sorted, tt.p, got, tt.want)
			}
		})
	}
}

func TestDeepEqualJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    any
		b    any
		want bool
	}{
		{"equal maps", map[string]int{"a": 1}, map[string]int{"a": 1}, true},
		{"different maps", map[string]int{"a": 1}, map[string]int{"a": 2}, false},
		{"equal slices", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different slices", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"nil values", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DeepEqualJSON(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("DeepEqualJSON(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMust_Success(t *testing.T) {
	t.Parallel()
	// Should not panic with nil error
	Must(nil)
}

func TestMust_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must(error) should panic")
		}
	}()
	Must(fmt.Errorf("test error"))
}

func TestDetectSubnet(t *testing.T) {
	t.Parallel()

	// This test just verifies the function runs without error
	// The actual result depends on the network configuration
	subnet, err := DetectSubnet()
	if err != nil {
		// On some CI environments there may be no suitable interface
		t.Skipf("DetectSubnet() returned error (may be expected in CI): %v", err)
	}

	// Verify it looks like a valid CIDR
	if subnet == "" {
		t.Error("DetectSubnet() returned empty string")
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
				Model:      types.ModelDisplayName("SHSW-1"), // Human-readable name
			},
		},
		{
			name: "gen2 device",
			d: discovery.DiscoveredDevice{
				ID:         "shellypro1pm-AABBCC",
				Address:    net.ParseIP("10.0.0.2"),
				Generation: 2,
				Model:      "SPSW-001PE16EU",
			},
			want: model.Device{
				Name:       "shellypro1pm-AABBCC",
				Address:    "10.0.0.2",
				Generation: 2,
				Type:       "SPSW-001PE16EU",
				Model:      types.ModelDisplayName("SPSW-001PE16EU"), // Human-readable name
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
				Model:      types.ModelDisplayName("S3PM-001PCEU16"), // Falls back to code if no mapping
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

func TestIsJSONObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"empty string", "", false},
		{"JSON object", `{"key": "value"}`, true},
		{"JSON object minimal", `{}`, true},
		{"not JSON", "hello", false},
		{"array", `[1, 2, 3]`, false},
		{"number", "42", false},
		{"starts with brace", "{", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsJSONObject(tt.s)
			if got != tt.want {
				t.Errorf("IsJSONObject(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestGetEditor(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv

	// Test with EDITOR set
	t.Run("with EDITOR env", func(t *testing.T) {
		t.Setenv("EDITOR", "nano")
		t.Setenv("VISUAL", "")
		editor := GetEditor()
		if editor != "nano" {
			t.Errorf("GetEditor() = %q, want nano", editor)
		}
	})

	// Test with VISUAL set but not EDITOR
	t.Run("with VISUAL env", func(t *testing.T) {
		t.Setenv("EDITOR", "")
		t.Setenv("VISUAL", "vim")
		editor := GetEditor()
		if editor != "vim" {
			t.Errorf("GetEditor() = %q, want vim", editor)
		}
	})
}

func TestResolveBatchTargets_Args(t *testing.T) {
	t.Parallel()

	// When args are provided, they should be returned
	targets, err := ResolveBatchTargets("", false, []string{"device1", "device2"})
	if err != nil {
		t.Fatalf("ResolveBatchTargets() error = %v", err)
	}
	if len(targets) != 2 {
		t.Errorf("len(targets) = %d, want 2", len(targets))
	}
	if targets[0] != "device1" {
		t.Errorf("targets[0] = %q, want device1", targets[0])
	}
}

func TestResolveBatchTargets_NoInput(t *testing.T) {
	t.Parallel()

	// When no input provided and not piped, should error
	// Note: this test may behave differently in TTY vs CI
	_, err := ResolveBatchTargets("", false, nil)
	if err == nil {
		t.Error("ResolveBatchTargets() should error when no input")
	}
}
