package shelly

import (
	"encoding/json"
	"testing"
)

func TestVirtualComponentType_Constants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vt   VirtualComponentType
		want string
	}{
		{"boolean", VirtualBoolean, "boolean"},
		{"number", VirtualNumber, "number"},
		{"text", VirtualText, "text"},
		{"enum", VirtualEnum, "enum"},
		{"button", VirtualButton, "button"},
		{"group", VirtualGroup, "group"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := string(tt.vt); got != tt.want {
				t.Errorf("VirtualComponentType = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsVirtualType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		vt   VirtualComponentType
		want bool
	}{
		{"boolean", VirtualBoolean, true},
		{"number", VirtualNumber, true},
		{"text", VirtualText, true},
		{"enum", VirtualEnum, true},
		{"button", VirtualButton, true},
		{"group", VirtualGroup, true},
		{"switch", VirtualComponentType("switch"), false},
		{"unknown", VirtualComponentType("unknown"), false},
		{"empty", VirtualComponentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isVirtualType(tt.vt); got != tt.want {
				t.Errorf("isVirtualType(%q) = %v, want %v", tt.vt, got, tt.want)
			}
		})
	}
}

func TestParseVirtualComponent_ValidBoolean(t *testing.T) {
	t.Parallel()
	comp := componentInfo{
		Key:    "boolean:200",
		Config: json.RawMessage(`{"name":"Test Bool"}`),
		Status: json.RawMessage(`{"value":true}`),
	}

	vc, ok := parseVirtualComponent(comp)
	if !ok {
		t.Fatal("parseVirtualComponent() returned false, want true")
	}
	if vc.Key != "boolean:200" {
		t.Errorf("Key = %q, want %q", vc.Key, "boolean:200")
	}
	if vc.Type != VirtualBoolean {
		t.Errorf("Type = %q, want %q", vc.Type, VirtualBoolean)
	}
	if vc.ID != 200 {
		t.Errorf("ID = %d, want 200", vc.ID)
	}
	if vc.Name != "Test Bool" {
		t.Errorf("Name = %q, want %q", vc.Name, "Test Bool")
	}
	if vc.BoolValue == nil || !*vc.BoolValue {
		t.Error("BoolValue should be true")
	}
}

func TestParseVirtualComponent_ValidNumber(t *testing.T) {
	t.Parallel()
	minVal := 0.0
	maxVal := 100.0
	comp := componentInfo{
		Key:    "number:201",
		Config: json.RawMessage(`{"name":"Temperature","min":0,"max":100,"unit":"°C"}`),
		Status: json.RawMessage(`{"value":22.5}`),
	}

	vc, ok := parseVirtualComponent(comp)
	if !ok {
		t.Fatal("parseVirtualComponent() returned false, want true")
	}
	if vc.Type != VirtualNumber {
		t.Errorf("Type = %q, want %q", vc.Type, VirtualNumber)
	}
	if vc.ID != 201 {
		t.Errorf("ID = %d, want 201", vc.ID)
	}
	if vc.Min == nil || *vc.Min != minVal {
		t.Errorf("Min = %v, want %v", vc.Min, minVal)
	}
	if vc.Max == nil || *vc.Max != maxVal {
		t.Errorf("Max = %v, want %v", vc.Max, maxVal)
	}
	if vc.NumValue == nil || *vc.NumValue != 22.5 {
		t.Error("NumValue should be 22.5")
	}
}

func TestParseVirtualComponent_ValidText(t *testing.T) {
	t.Parallel()
	comp := componentInfo{
		Key:    "text:202",
		Config: json.RawMessage(`{"name":"Note"}`),
		Status: json.RawMessage(`{"value":"Hello"}`),
	}

	vc, ok := parseVirtualComponent(comp)
	if !ok {
		t.Fatal("parseVirtualComponent() returned false, want true")
	}
	if vc.Type != VirtualText {
		t.Errorf("Type = %q, want %q", vc.Type, VirtualText)
	}
	if vc.StrValue == nil || *vc.StrValue != "Hello" {
		t.Error("StrValue should be 'Hello'")
	}
}

func TestParseVirtualComponent_ValidEnum(t *testing.T) {
	t.Parallel()
	comp := componentInfo{
		Key:    "enum:203",
		Config: json.RawMessage(`{"name":"Mode","options":["off","low","high"]}`),
		Status: json.RawMessage(`{"value":"low"}`),
	}

	vc, ok := parseVirtualComponent(comp)
	if !ok {
		t.Fatal("parseVirtualComponent() returned false, want true")
	}
	if vc.Type != VirtualEnum {
		t.Errorf("Type = %q, want %q", vc.Type, VirtualEnum)
	}
	if len(vc.Options) != 3 {
		t.Errorf("Options length = %d, want 3", len(vc.Options))
	}
}

func TestParseVirtualComponent_NonVirtualID(t *testing.T) {
	t.Parallel()
	// ID < 200 should not be recognized as virtual
	comp := componentInfo{
		Key: "boolean:0",
	}

	_, ok := parseVirtualComponent(comp)
	if ok {
		t.Error("parseVirtualComponent() returned true for ID 0, want false")
	}

	// ID > 299 should not be recognized as virtual
	comp = componentInfo{
		Key: "boolean:300",
	}

	_, ok = parseVirtualComponent(comp)
	if ok {
		t.Error("parseVirtualComponent() returned true for ID 300, want false")
	}
}

func TestParseVirtualComponent_NonVirtualType(t *testing.T) {
	t.Parallel()
	comp := componentInfo{
		Key: "switch:200",
	}

	_, ok := parseVirtualComponent(comp)
	if ok {
		t.Error("parseVirtualComponent() returned true for switch type, want false")
	}
}

func TestParseVirtualComponent_InvalidKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"no colon", "boolean200"},
		{"too many parts", "boolean:200:extra"},
		{"non-numeric id", "boolean:abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			comp := componentInfo{Key: tt.key}
			_, ok := parseVirtualComponent(comp)
			if ok {
				t.Errorf("parseVirtualComponent(%q) returned true, want false", tt.key)
			}
		})
	}
}

func TestVirtualComponent_Fields(t *testing.T) {
	t.Parallel()
	boolVal := true
	numVal := 42.0
	strVal := "test"
	minVal := 0.0
	maxVal := 100.0
	unit := "°C"

	vc := VirtualComponent{
		Key:       "boolean:200",
		Type:      VirtualBoolean,
		ID:        200,
		Name:      "Test",
		Value:     true,
		BoolValue: &boolVal,
		NumValue:  &numVal,
		StrValue:  &strVal,
		Options:   []string{"a", "b"},
		Min:       &minVal,
		Max:       &maxVal,
		Unit:      &unit,
	}

	if vc.Key != "boolean:200" {
		t.Errorf("Key = %q, want %q", vc.Key, "boolean:200")
	}
	if vc.Type != VirtualBoolean {
		t.Errorf("Type = %q, want %q", vc.Type, VirtualBoolean)
	}
	if vc.ID != 200 {
		t.Errorf("ID = %d, want 200", vc.ID)
	}
	if vc.Name != "Test" {
		t.Errorf("Name = %q, want %q", vc.Name, "Test")
	}
}

func TestAddVirtualComponentParams(t *testing.T) {
	t.Parallel()
	params := AddVirtualComponentParams{
		Type:   VirtualBoolean,
		ID:     200,
		Name:   "Test",
		Config: map[string]any{"persisted": true},
	}

	if params.Type != VirtualBoolean {
		t.Errorf("Type = %q, want %q", params.Type, VirtualBoolean)
	}
	if params.ID != 200 {
		t.Errorf("ID = %d, want 200", params.ID)
	}
	if params.Name != "Test" {
		t.Errorf("Name = %q, want %q", params.Name, "Test")
	}
}
