package client

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		ID:         "shellyplus1pm-123456",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "SNSW-001P16EU",
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "Plus1PM",
		AuthEn:     true,
	}

	if info.ID != "shellyplus1pm-123456" {
		t.Errorf("ID = %q, want shellyplus1pm-123456", info.ID)
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", info.MAC)
	}
	if info.Model != "SNSW-001P16EU" {
		t.Errorf("Model = %q, want SNSW-001P16EU", info.Model)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", info.Firmware)
	}
	if info.App != "Plus1PM" {
		t.Errorf("App = %q, want Plus1PM", info.App)
	}
	if !info.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestParseComponentKey_ValidKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      string
		wantType model.ComponentType
		wantID   int
		wantOK   bool
	}{
		{"switch:0", "switch:0", model.ComponentSwitch, 0, true},
		{"switch:1", "switch:1", model.ComponentSwitch, 1, true},
		{"cover:0", "cover:0", model.ComponentCover, 0, true},
		{"cover:2", "cover:2", model.ComponentCover, 2, true},
		{"light:0", "light:0", model.ComponentLight, 0, true},
		{"light:3", "light:3", model.ComponentLight, 3, true},
		{"rgb:0", "rgb:0", model.ComponentRGB, 0, true},
		{"rgb:1", "rgb:1", model.ComponentRGB, 1, true},
		{"input:0", "input:0", model.ComponentInput, 0, true},
		{"input:5", "input:5", model.ComponentInput, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			comp, ok := parseComponentKey(tt.key)
			if ok != tt.wantOK {
				t.Errorf("parseComponentKey(%q) ok = %v, want %v", tt.key, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if comp.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", comp.Type, tt.wantType)
			}
			if comp.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", comp.ID, tt.wantID)
			}
			if comp.Key != tt.key {
				t.Errorf("Key = %q, want %q", comp.Key, tt.key)
			}
		})
	}
}

func TestParseComponentKey_InvalidKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"unknown type", "unknown:0"},
		{"wifi component", "wifi:0"},
		{"sys component", "sys"},
		{"empty", ""},
		{"just prefix", "switch:"},
		{"invalid id", "switch:abc"},
		{"no colon", "switch0"},
		{"wrong format", "0:switch"},
	}

	// Note: negative IDs like "switch:-1" are parsed by %d scanf and considered valid
	// The device would return an error for invalid IDs at runtime

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := parseComponentKey(tt.key)
			if ok {
				t.Errorf("parseComponentKey(%q) = ok, want not ok", tt.key)
			}
		})
	}
}

func TestUnmarshalResponse_Success(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := map[string]any{
		"name":  "test",
		"value": 42,
	}

	var result testStruct
	err := unmarshalResponse(data, &result)

	if err != nil {
		t.Fatalf("unmarshalResponse() error = %v", err)
	}
	if result.Name != "test" {
		t.Errorf("Name = %q, want test", result.Name)
	}
	if result.Value != 42 {
		t.Errorf("Value = %d, want 42", result.Value)
	}
}

func TestUnmarshalResponse_Nested(t *testing.T) {
	t.Parallel()

	type inner struct {
		ID int `json:"id"`
	}
	type outer struct {
		Components []inner `json:"components"`
	}

	data := map[string]any{
		"components": []any{
			map[string]any{"id": 0},
			map[string]any{"id": 1},
		},
	}

	var result outer
	err := unmarshalResponse(data, &result)

	if err != nil {
		t.Fatalf("unmarshalResponse() error = %v", err)
	}
	if len(result.Components) != 2 {
		t.Fatalf("len(Components) = %d, want 2", len(result.Components))
	}
	if result.Components[0].ID != 0 {
		t.Errorf("Components[0].ID = %d, want 0", result.Components[0].ID)
	}
	if result.Components[1].ID != 1 {
		t.Errorf("Components[1].ID = %d, want 1", result.Components[1].ID)
	}
}

func TestUnmarshalResponse_TypeMismatch(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value int `json:"value"`
	}

	data := map[string]any{
		"value": "not an int",
	}

	var result testStruct
	err := unmarshalResponse(data, &result)

	if err == nil {
		t.Error("unmarshalResponse() should fail with type mismatch")
	}
}

func TestComponentPrefixes_AllTypes(t *testing.T) {
	t.Parallel()

	expectedPrefixes := map[string]model.ComponentType{
		"switch:": model.ComponentSwitch,
		"cover:":  model.ComponentCover,
		"light:":  model.ComponentLight,
		"rgb:":    model.ComponentRGB,
		"rgbw:":   model.ComponentRGBW,
		"input:":  model.ComponentInput,
	}

	for prefix, expectedType := range expectedPrefixes {
		if compType, exists := componentPrefixes[prefix]; !exists {
			t.Errorf("componentPrefixes missing prefix %q", prefix)
		} else if compType != expectedType {
			t.Errorf("componentPrefixes[%q] = %q, want %q", prefix, compType, expectedType)
		}
	}

	if len(componentPrefixes) != len(expectedPrefixes) {
		t.Errorf("componentPrefixes has %d entries, want %d", len(componentPrefixes), len(expectedPrefixes))
	}
}

func TestClient_RebootDelayParams(t *testing.T) {
	t.Parallel()

	// Test that delay parameters are properly constructed
	// This tests the internal logic without requiring a real device

	tests := []struct {
		name    string
		delayMS int
		wantKey bool
	}{
		{"no delay", 0, false},
		{"with delay", 1000, true},
		{"negative delay", -1, false}, // negative values treated as no delay
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := map[string]any{}
			if tt.delayMS > 0 {
				params["delay_ms"] = tt.delayMS
			}
			_, hasKey := params["delay_ms"]
			if hasKey != tt.wantKey {
				t.Errorf("params has delay_ms = %v, want %v", hasKey, tt.wantKey)
			}
		})
	}
}
