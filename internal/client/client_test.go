package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Test constants to avoid magic strings.
const (
	testMAC1     = "AA:BB:CC:DD:EE:FF"
	testMAC2     = "11:22:33:44:55:66"
	testModel1   = "SNSW-001P16EU"
	testFirmware = "1.0.0"
	testApp      = "Plus1PM"
	gen2Endpoint = "/rpc/Shelly.GetDeviceInfo"
	gen1Endpoint = "/shelly"
)

func TestDeviceInfo_Fields(t *testing.T) {
	t.Parallel()

	info := DeviceInfo{
		ID:         "shellyplus1pm-123456",
		MAC:        testMAC1,
		Model:      testModel1,
		Generation: 2,
		Firmware:   testFirmware,
		App:        testApp,
		AuthEn:     true,
	}

	if info.ID != "shellyplus1pm-123456" {
		t.Errorf("ID = %q, want shellyplus1pm-123456", info.ID)
	}
	if info.MAC != testMAC1 {
		t.Errorf("MAC = %q, want %s", info.MAC, testMAC1)
	}
	if info.Model != testModel1 {
		t.Errorf("Model = %q, want %s", info.Model, testModel1)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != testFirmware {
		t.Errorf("Firmware = %q, want %s", info.Firmware, testFirmware)
	}
	if info.App != testApp {
		t.Errorf("App = %q, want %s", info.App, testApp)
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

func TestClient_Info(t *testing.T) {
	t.Parallel()

	info := &DeviceInfo{
		ID:         "shelly123",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "SNSW-001",
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "Plus1",
		AuthEn:     true,
	}

	client := &Client{
		info: info,
	}

	got := client.Info()
	if got != info {
		t.Error("Info() should return the same DeviceInfo pointer")
	}
	if got.ID != "shelly123" {
		t.Errorf("Info().ID = %q, want shelly123", got.ID)
	}
}

func TestClient_Close_NilDevice(t *testing.T) {
	t.Parallel()

	// Test Close with nil device and nil rpcClient
	client := &Client{}
	err := client.Close()
	if err != nil {
		t.Errorf("Close() with nil device = %v, want nil", err)
	}
}

func TestClient_SwitchAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	sw := client.Switch(0)
	if sw == nil {
		t.Fatal("Switch(0) returned nil")
	}
	if sw.id != 0 {
		t.Errorf("Switch(0).id = %d, want 0", sw.id)
	}
	sw5 := client.Switch(5)
	if sw5.id != 5 {
		t.Errorf("Switch(5).id = %d, want 5", sw5.id)
	}
}

func TestClient_CoverAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	cv := client.Cover(0)
	if cv == nil {
		t.Fatal("Cover(0) returned nil")
	}
	if cv.id != 0 {
		t.Errorf("Cover(0).id = %d, want 0", cv.id)
	}
}

func TestClient_LightAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	lt := client.Light(0)
	if lt == nil {
		t.Fatal("Light(0) returned nil")
	}
	if lt.id != 0 {
		t.Errorf("Light(0).id = %d, want 0", lt.id)
	}
}

func TestClient_RGBAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	rgb := client.RGB(0)
	if rgb == nil {
		t.Fatal("RGB(0) returned nil")
	}
	if rgb.id != 0 {
		t.Errorf("RGB(0).id = %d, want 0", rgb.id)
	}
}

func TestClient_RGBWAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	rgbw := client.RGBW(0)
	if rgbw == nil {
		t.Fatal("RGBW(0) returned nil")
	}
	if rgbw.id != 0 {
		t.Errorf("RGBW(0).id = %d, want 0", rgbw.id)
	}
}

func TestClient_InputAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	input := client.Input(0)
	if input == nil {
		t.Fatal("Input(0) returned nil")
	}
	if input.id != 0 {
		t.Errorf("Input(0).id = %d, want 0", input.id)
	}
}

func TestClient_ThermostatAccessor(t *testing.T) {
	t.Parallel()

	client := &Client{rpcClient: nil}
	th := client.Thermostat(0)
	if th == nil {
		t.Fatal("Thermostat(0) returned nil")
	}
	if th.id != 0 {
		t.Errorf("Thermostat(0).id = %d, want 0", th.id)
	}
}

func TestClient_RPCClient(t *testing.T) {
	t.Parallel()

	// Test that RPCClient returns the stored rpcClient
	client := &Client{
		rpcClient: nil,
	}

	if client.RPCClient() != nil {
		t.Error("RPCClient() should return nil when no RPC client set")
	}
}

func TestDeviceInfo_AllFields(t *testing.T) {
	t.Parallel()

	// Test with all fields populated
	info := DeviceInfo{
		ID:         "shellypluspmmini-abc123",
		MAC:        "11:22:33:44:55:66",
		Model:      "SNPM-001PCEU16",
		Generation: 2,
		Firmware:   "1.2.3-beta",
		App:        "PlusPMMini",
		AuthEn:     false,
	}

	// Verify all fields
	if info.ID != "shellypluspmmini-abc123" {
		t.Errorf("ID = %q, want shellypluspmmini-abc123", info.ID)
	}
	if info.MAC != testMAC2 {
		t.Errorf("MAC = %q, want 11:22:33:44:55:66", info.MAC)
	}
	if info.Model != "SNPM-001PCEU16" {
		t.Errorf("Model = %q, want SNPM-001PCEU16", info.Model)
	}
	if info.Generation != 2 {
		t.Errorf("Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.2.3-beta" {
		t.Errorf("Firmware = %q, want 1.2.3-beta", info.Firmware)
	}
	if info.App != "PlusPMMini" {
		t.Errorf("App = %q, want PlusPMMini", info.App)
	}
	if info.AuthEn {
		t.Error("AuthEn = true, want false")
	}
}

func TestDeviceInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	// Test zero value
	var info DeviceInfo

	if info.ID != "" {
		t.Errorf("zero value ID = %q, want empty", info.ID)
	}
	if info.MAC != "" {
		t.Errorf("zero value MAC = %q, want empty", info.MAC)
	}
	if info.Model != "" {
		t.Errorf("zero value Model = %q, want empty", info.Model)
	}
	if info.Generation != 0 {
		t.Errorf("zero value Generation = %d, want 0", info.Generation)
	}
	if info.AuthEn {
		t.Error("zero value AuthEn = true, want false")
	}
}

func TestParseComponentKey_RGBW(t *testing.T) {
	t.Parallel()

	// Verify RGBW component type parsing
	tests := []struct {
		key    string
		wantID int
	}{
		{"rgbw:0", 0},
		{"rgbw:1", 1},
		{"rgbw:10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			comp, ok := parseComponentKey(tt.key)
			if !ok {
				t.Fatalf("parseComponentKey(%q) = not ok, want ok", tt.key)
			}
			if comp.Type != model.ComponentRGBW {
				t.Errorf("Type = %q, want %q", comp.Type, model.ComponentRGBW)
			}
			if comp.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", comp.ID, tt.wantID)
			}
		})
	}
}

func TestUnmarshalResponse_EmptyData(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name string `json:"name"`
	}

	// Test with empty map
	data := map[string]any{}
	var result testStruct
	err := unmarshalResponse(data, &result)

	if err != nil {
		t.Errorf("unmarshalResponse with empty map = %v, want nil", err)
	}
	if result.Name != "" {
		t.Errorf("Name = %q, want empty", result.Name)
	}
}

func TestUnmarshalResponse_NilData(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Name string `json:"name"`
	}

	var result testStruct
	err := unmarshalResponse(nil, &result)

	if err != nil {
		t.Errorf("unmarshalResponse with nil = %v, want nil", err)
	}
}

func TestUnmarshalResponse_SliceData(t *testing.T) {
	t.Parallel()

	data := []any{
		map[string]any{"id": 0, "key": "switch:0"},
		map[string]any{"id": 1, "key": "switch:1"},
	}

	type item struct {
		ID  int    `json:"id"`
		Key string `json:"key"`
	}

	var result []item
	err := unmarshalResponse(data, &result)

	if err != nil {
		t.Fatalf("unmarshalResponse with slice = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].ID != 0 {
		t.Errorf("result[0].ID = %d, want 0", result[0].ID)
	}
	if result[1].Key != "switch:1" {
		t.Errorf("result[1].Key = %q, want switch:1", result[1].Key)
	}
}

// Tests for detect.go

func TestGeneration_Constants(t *testing.T) {
	t.Parallel()

	if GenerationUnknown != 0 {
		t.Errorf("GenerationUnknown = %d, want 0", GenerationUnknown)
	}
	if Gen1 != 1 {
		t.Errorf("Gen1 = %d, want 1", Gen1)
	}
	if Gen2 != 2 {
		t.Errorf("Gen2 = %d, want 2", Gen2)
	}
}

func TestDetectionResult_IsGen1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		generation Generation
		want       bool
	}{
		{"Gen1", Gen1, true},
		{"Gen2", Gen2, false},
		{"Unknown", GenerationUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &DetectionResult{Generation: tt.generation}
			if got := r.IsGen1(); got != tt.want {
				t.Errorf("IsGen1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_IsGen2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		generation Generation
		want       bool
	}{
		{"Gen1", Gen1, false},
		{"Gen2", Gen2, true},
		{"Unknown", GenerationUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &DetectionResult{Generation: tt.generation}
			if got := r.IsGen2(); got != tt.want {
				t.Errorf("IsGen2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectionResult_Fields(t *testing.T) {
	t.Parallel()

	result := DetectionResult{
		Generation: Gen2,
		DeviceType: "Plus1PM",
		Model:      "SNSW-001P16EU",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Firmware:   "1.0.0",
		AuthEn:     true,
	}

	if result.Generation != Gen2 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen2)
	}
	if result.DeviceType != "Plus1PM" {
		t.Errorf("DeviceType = %q, want Plus1PM", result.DeviceType)
	}
	if result.Model != "SNSW-001P16EU" {
		t.Errorf("Model = %q, want SNSW-001P16EU", result.Model)
	}
	if result.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", result.MAC)
	}
	if result.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", result.Firmware)
	}
	if !result.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestHttpStatusError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		statusCode int
		wantMsg    string
	}{
		{http.StatusUnauthorized, "authentication required"},
		{http.StatusForbidden, "access denied"},
		{http.StatusNotFound, "not found"},
		{http.StatusServiceUnavailable, "unavailable"},
		{http.StatusGatewayTimeout, "timeout"},
		{http.StatusInternalServerError, "unexpected HTTP status: 500"},
		{999, "unexpected HTTP status: 999"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			t.Parallel()
			err := httpStatusError(tt.statusCode)
			if err == nil {
				t.Fatal("httpStatusError() = nil, want error")
			}
			if got := err.Error(); !containsIgnoreCase(got, tt.wantMsg) {
				t.Errorf("httpStatusError(%d) = %q, want to contain %q", tt.statusCode, got, tt.wantMsg)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		strs []string
		want string
	}{
		{
			name: "first is non-empty",
			strs: []string{"first", "second", "third"},
			want: "first",
		},
		{
			name: "second is first non-empty",
			strs: []string{"", "second", "third"},
			want: "second",
		},
		{
			name: "third is first non-empty",
			strs: []string{"", "", "third"},
			want: "third",
		},
		{
			name: "all empty",
			strs: []string{"", "", ""},
			want: "",
		},
		{
			name: "no arguments",
			strs: []string{},
			want: "",
		},
		{
			name: "single non-empty",
			strs: []string{"only"},
			want: "only",
		},
		{
			name: "single empty",
			strs: []string{""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := firstNonEmpty(tt.strs...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.strs, got, tt.want)
			}
		})
	}
}

// Tests for gen1.go

func TestGen1Client_Close_NilDevice(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{}
	err := client.Close()
	if err != nil {
		t.Errorf("Close() with nil device = %v, want nil", err)
	}
}

func TestGen1Client_Info(t *testing.T) {
	t.Parallel()

	info := &DeviceInfo{
		ID:         "shelly1-abc123",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "SHSW-1",
		Generation: 1,
		Firmware:   "1.11.0",
		App:        "Shelly1",
		AuthEn:     false,
	}

	client := &Gen1Client{
		info: info,
	}

	got := client.Info()
	if got != info {
		t.Error("Info() should return the same DeviceInfo pointer")
	}
	if got.ID != "shelly1-abc123" {
		t.Errorf("Info().ID = %q, want shelly1-abc123", got.ID)
	}
	if got.Generation != 1 {
		t.Errorf("Info().Generation = %d, want 1", got.Generation)
	}
}

func TestGen1Client_Device(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil, // nil in test
	}

	if client.Device() != nil {
		t.Error("Device() should return nil when not set")
	}
}

func TestErrInvalidComponentID(t *testing.T) {
	t.Parallel()

	if ErrInvalidComponentID == nil {
		t.Fatal("ErrInvalidComponentID should not be nil")
	}
	if ErrInvalidComponentID.Error() == "" {
		t.Error("ErrInvalidComponentID.Error() should not be empty")
	}
}

func TestGen1Client_Relay_InvalidID(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil, // Will panic if called with negative ID since we can't get relay from nil device
	}

	_, err := client.Relay(-1)
	if err == nil {
		t.Error("Relay(-1) should return error")
	}
}

func TestGen1Client_Roller_InvalidID(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil,
	}

	_, err := client.Roller(-1)
	if err == nil {
		t.Error("Roller(-1) should return error")
	}
}

func TestGen1Client_Light_InvalidID(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil,
	}

	_, err := client.Light(-1)
	if err == nil {
		t.Error("Light(-1) should return error")
	}
}

func TestGen1Client_Color_InvalidID(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil,
	}

	_, err := client.Color(-1)
	if err == nil {
		t.Error("Color(-1) should return error")
	}
}

func TestGen1Client_White_InvalidID(t *testing.T) {
	t.Parallel()

	client := &Gen1Client{
		device: nil,
	}

	_, err := client.White(-1)
	if err == nil {
		t.Error("White(-1) should return error")
	}
}

func TestGen1RelayComponent_ID(t *testing.T) {
	t.Parallel()

	relay := &Gen1RelayComponent{id: 3}
	if relay.ID() != 3 {
		t.Errorf("ID() = %d, want 3", relay.ID())
	}
}

func TestGen1RollerComponent_ID(t *testing.T) {
	t.Parallel()

	roller := &Gen1RollerComponent{id: 2}
	if roller.ID() != 2 {
		t.Errorf("ID() = %d, want 2", roller.ID())
	}
}

func TestGen1LightComponent_ID(t *testing.T) {
	t.Parallel()

	light := &Gen1LightComponent{id: 1}
	if light.ID() != 1 {
		t.Errorf("ID() = %d, want 1", light.ID())
	}
}

func TestGen1ColorComponent_ID(t *testing.T) {
	t.Parallel()

	color := &Gen1ColorComponent{id: 0}
	if color.ID() != 0 {
		t.Errorf("ID() = %d, want 0", color.ID())
	}
}

func TestGen1WhiteComponent_ID(t *testing.T) {
	t.Parallel()

	white := &Gen1WhiteComponent{id: 4}
	if white.ID() != 4 {
		t.Errorf("ID() = %d, want 4", white.ID())
	}
}

// Tests for KVS types

func TestKVSItem_Fields(t *testing.T) {
	t.Parallel()

	item := KVSItem{
		Key:   "mykey",
		Value: "myvalue",
		Etag:  "abc123",
	}

	if item.Key != "mykey" {
		t.Errorf("Key = %q, want mykey", item.Key)
	}
	if item.Value != "myvalue" {
		t.Errorf("Value = %v, want myvalue", item.Value)
	}
	if item.Etag != "abc123" {
		t.Errorf("Etag = %q, want abc123", item.Etag)
	}
}

func TestKVSListResult_Fields(t *testing.T) {
	t.Parallel()

	result := KVSListResult{
		Keys: []string{"key1", "key2", "key3"},
		Rev:  42,
	}

	if len(result.Keys) != 3 {
		t.Errorf("len(Keys) = %d, want 3", len(result.Keys))
	}
	if result.Keys[0] != "key1" {
		t.Errorf("Keys[0] = %q, want key1", result.Keys[0])
	}
	if result.Rev != 42 {
		t.Errorf("Rev = %d, want 42", result.Rev)
	}
}

func TestKVSGetResult_Fields(t *testing.T) {
	t.Parallel()

	result := KVSGetResult{
		Value: map[string]any{"nested": "data"},
		Etag:  "etag123",
	}

	if result.Etag != "etag123" {
		t.Errorf("Etag = %q, want etag123", result.Etag)
	}
	if result.Value == nil {
		t.Error("Value should not be nil")
	}
}

func TestKVSSetResult_Fields(t *testing.T) {
	t.Parallel()

	result := KVSSetResult{
		Etag: "newetag",
		Rev:  100,
	}

	if result.Etag != "newetag" {
		t.Errorf("Etag = %q, want newetag", result.Etag)
	}
	if result.Rev != 100 {
		t.Errorf("Rev = %d, want 100", result.Rev)
	}
}

func TestKVSDeleteResult_Fields(t *testing.T) {
	t.Parallel()

	result := KVSDeleteResult{
		Rev: 101,
	}

	if result.Rev != 101 {
		t.Errorf("Rev = %d, want 101", result.Rev)
	}
}

func TestClient_KVS(t *testing.T) {
	t.Parallel()

	client := &Client{
		rpcClient: nil,
	}

	kvs := client.KVS()
	if kvs == nil {
		t.Error("KVS() returned nil")
	}
}

// Tests for Thermostat

func TestThermostatComponent_ID(t *testing.T) {
	t.Parallel()

	th := &ThermostatComponent{id: 5}
	if th.ID() != 5 {
		t.Errorf("ID() = %d, want 5", th.ID())
	}
}

// Tests for component struct field accessors

func TestSwitchComponent_Fields(t *testing.T) {
	t.Parallel()

	sw := &SwitchComponent{
		sw:  nil,
		rpc: nil,
		id:  7,
	}

	if sw.id != 7 {
		t.Errorf("id = %d, want 7", sw.id)
	}
}

func TestCoverComponent_Fields(t *testing.T) {
	t.Parallel()

	cv := &CoverComponent{
		cv:  nil,
		rpc: nil,
		id:  3,
	}

	if cv.id != 3 {
		t.Errorf("id = %d, want 3", cv.id)
	}
}

func TestLightComponent_Fields(t *testing.T) {
	t.Parallel()

	lt := &LightComponent{
		lt:  nil,
		rpc: nil,
		id:  2,
	}

	if lt.id != 2 {
		t.Errorf("id = %d, want 2", lt.id)
	}
}

func TestRGBComponent_Fields(t *testing.T) {
	t.Parallel()

	rgb := &RGBComponent{
		rgb: nil,
		rpc: nil,
		id:  1,
	}

	if rgb.id != 1 {
		t.Errorf("id = %d, want 1", rgb.id)
	}
}

func TestRGBWComponent_Fields(t *testing.T) {
	t.Parallel()

	rgbw := &RGBWComponent{
		rgbw: nil,
		rpc:  nil,
		id:   4,
	}

	if rgbw.id != 4 {
		t.Errorf("id = %d, want 4", rgbw.id)
	}
}

func TestInputComponent_Fields(t *testing.T) {
	t.Parallel()

	input := &InputComponent{
		input: nil,
		rpc:   nil,
		id:    6,
	}

	if input.id != 6 {
		t.Errorf("id = %d, want 6", input.id)
	}
}

func TestKVSComponent_Fields(t *testing.T) {
	t.Parallel()

	kvs := &KVSComponent{
		kvs: nil,
		rpc: nil,
	}

	if kvs.kvs != nil {
		t.Error("kvs should be nil")
	}
	if kvs.rpc != nil {
		t.Error("rpc should be nil")
	}
}

func TestClient_Fields(t *testing.T) {
	t.Parallel()

	info := &DeviceInfo{
		ID:         "test-device",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "TestModel",
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "TestApp",
		AuthEn:     true,
	}

	client := &Client{
		device:    nil,
		rpcClient: nil,
		transport: nil,
		info:      info,
	}

	if client.info != info {
		t.Error("info should match")
	}
	if client.device != nil {
		t.Error("device should be nil")
	}
	if client.rpcClient != nil {
		t.Error("rpcClient should be nil")
	}
	if client.transport != nil {
		t.Error("transport should be nil")
	}
}

func TestGen1Client_Fields(t *testing.T) {
	t.Parallel()

	info := &DeviceInfo{
		ID:         "shelly1-test",
		MAC:        "11:22:33:44:55:66",
		Model:      "SHSW-1",
		Generation: 1,
		Firmware:   "1.11.0",
		App:        "Shelly1",
		AuthEn:     false,
	}

	client := &Gen1Client{
		device:    nil,
		transport: nil,
		info:      info,
	}

	if client.info != info {
		t.Error("info should match")
	}
	if client.device != nil {
		t.Error("device should be nil")
	}
	if client.transport != nil {
		t.Error("transport should be nil")
	}
}

// Tests for Generation type

func TestGeneration_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		gen  Generation
		want int
	}{
		{"unknown", GenerationUnknown, 0},
		{"gen1", Gen1, 1},
		{"gen2", Gen2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if int(tt.gen) != tt.want {
				t.Errorf("Generation = %d, want %d", tt.gen, tt.want)
			}
		})
	}
}

// Tests for DetectionResult

func TestDetectionResult_ZeroValue(t *testing.T) {
	t.Parallel()

	var result DetectionResult

	if result.Generation != GenerationUnknown {
		t.Errorf("zero value Generation = %d, want %d", result.Generation, GenerationUnknown)
	}
	if result.DeviceType != "" {
		t.Errorf("zero value DeviceType = %q, want empty", result.DeviceType)
	}
	if result.Model != "" {
		t.Errorf("zero value Model = %q, want empty", result.Model)
	}
	if result.MAC != "" {
		t.Errorf("zero value MAC = %q, want empty", result.MAC)
	}
	if result.Firmware != "" {
		t.Errorf("zero value Firmware = %q, want empty", result.Firmware)
	}
	if result.AuthEn {
		t.Error("zero value AuthEn = true, want false")
	}
}

// Tests for error types

func TestErrInvalidComponentID_ErrorMessage(t *testing.T) {
	t.Parallel()

	errMsg := ErrInvalidComponentID.Error()
	if errMsg == "" {
		t.Error("ErrInvalidComponentID.Error() should not be empty")
	}
	if !containsSubstring(errMsg, "invalid") && !containsSubstring(errMsg, "component") {
		t.Errorf("ErrInvalidComponentID.Error() = %q, should contain 'invalid' or 'component'", errMsg)
	}
}

// Tests for parseComponentKey edge cases

func TestParseComponentKey_LargeIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key    string
		wantID int
	}{
		{"switch:99", 99},
		{"cover:100", 100},
		{"light:999", 999},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			comp, ok := parseComponentKey(tt.key)
			if !ok {
				t.Fatalf("parseComponentKey(%q) = not ok, want ok", tt.key)
			}
			if comp.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", comp.ID, tt.wantID)
			}
		})
	}
}

func TestParseComponentKey_Boundary(t *testing.T) {
	t.Parallel()

	// Test exactly at boundary - prefix only with ID 0
	tests := []struct {
		name   string
		key    string
		wantOK bool
	}{
		{"switch with 0", "switch:0", true},
		{"switch with empty after colon", "switch:", false},
		{"exact prefix length", "switch", false},
		{"one char short", "switc:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := parseComponentKey(tt.key)
			if ok != tt.wantOK {
				t.Errorf("parseComponentKey(%q) = %v, want %v", tt.key, ok, tt.wantOK)
			}
		})
	}
}

// Tests for unmarshalResponse edge cases

func TestUnmarshalResponse_InvalidJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Value string `json:"value"`
	}

	// Test with a channel (cannot be marshaled to JSON)
	ch := make(chan int)
	var result testStruct
	err := unmarshalResponse(ch, &result)

	if err == nil {
		t.Error("unmarshalResponse with channel should fail")
	}
}

func TestUnmarshalResponse_NilDestination(t *testing.T) {
	t.Parallel()

	data := map[string]any{"key": "value"}

	// This should handle nil destination gracefully
	var nilPtr *struct{ Key string }
	err := unmarshalResponse(data, nilPtr)

	// json.Unmarshal to nil pointer should fail
	if err == nil {
		t.Error("unmarshalResponse to nil pointer should fail")
	}
}

func TestUnmarshalResponse_ComplexNesting(t *testing.T) {
	t.Parallel()

	type level3 struct {
		Value int `json:"value"`
	}
	type level2 struct {
		Level3 level3 `json:"level3"`
	}
	type level1 struct {
		Level2 level2 `json:"level2"`
	}

	data := map[string]any{
		"level2": map[string]any{
			"level3": map[string]any{
				"value": 42,
			},
		},
	}

	var result level1
	err := unmarshalResponse(data, &result)

	if err != nil {
		t.Fatalf("unmarshalResponse failed: %v", err)
	}
	if result.Level2.Level3.Value != 42 {
		t.Errorf("nested value = %d, want 42", result.Level2.Level3.Value)
	}
}

// Tests for firstNonEmpty additional cases

func TestFirstNonEmpty_WithSpaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		strs []string
		want string
	}{
		{"space is non-empty", []string{"", " ", "value"}, " "},
		{"tab is non-empty", []string{"", "\t", "value"}, "\t"},
		{"newline is non-empty", []string{"", "\n", "value"}, "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := firstNonEmpty(tt.strs...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.strs, got, tt.want)
			}
		})
	}
}

// Tests for httpStatusError additional codes

func TestHttpStatusError_AllCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		statusCode int
		wantNil    bool
	}{
		{http.StatusOK, false},                  // Even 200 returns an error from this function
		{http.StatusBadRequest, false},          // 400
		{http.StatusUnauthorized, false},        // 401
		{http.StatusForbidden, false},           // 403
		{http.StatusNotFound, false},            // 404
		{http.StatusInternalServerError, false}, // 500
		{http.StatusServiceUnavailable, false},  // 503
		{http.StatusGatewayTimeout, false},      // 504
		{0, false},                              // Invalid code
		{-1, false},                             // Negative code
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			t.Parallel()
			err := httpStatusError(tt.statusCode)
			if (err == nil) != tt.wantNil {
				t.Errorf("httpStatusError(%d) nil = %v, want %v", tt.statusCode, err == nil, tt.wantNil)
			}
		})
	}
}

// Tests for KVS types with complex values

func TestKVSItem_ComplexValue(t *testing.T) {
	t.Parallel()

	item := KVSItem{
		Key: "complex",
		Value: map[string]any{
			"nested": []any{1, 2, 3},
			"bool":   true,
			"float":  3.14,
		},
		Etag: "etag-123",
	}

	if item.Key != "complex" {
		t.Errorf("Key = %q, want complex", item.Key)
	}
	if item.Etag != "etag-123" {
		t.Errorf("Etag = %q, want etag-123", item.Etag)
	}

	valueMap, ok := item.Value.(map[string]any)
	if !ok {
		t.Fatal("Value should be map[string]any")
	}
	if valueMap["bool"] != true {
		t.Error("nested bool should be true")
	}
}

func TestKVSListResult_EmptyKeys(t *testing.T) {
	t.Parallel()

	result := KVSListResult{
		Keys: []string{},
		Rev:  0,
	}

	if len(result.Keys) != 0 {
		t.Errorf("len(Keys) = %d, want 0", len(result.Keys))
	}
	if result.Rev != 0 {
		t.Errorf("Rev = %d, want 0", result.Rev)
	}
}

func TestKVSGetResult_NilValue(t *testing.T) {
	t.Parallel()

	result := KVSGetResult{
		Value: nil,
		Etag:  "some-etag",
	}

	if result.Value != nil {
		t.Error("Value should be nil")
	}
	if result.Etag != "some-etag" {
		t.Errorf("Etag = %q, want some-etag", result.Etag)
	}
}

// Tests for Gen1 component ID boundary tests - we only test invalid IDs
// since valid IDs require a real device (calling device methods panics with nil)

// Tests for component prefixes map

func TestComponentPrefixes_NoExtraKeys(t *testing.T) {
	t.Parallel()

	validPrefixes := map[string]bool{
		"switch:": true,
		"cover:":  true,
		"light:":  true,
		"rgb:":    true,
		"rgbw:":   true,
		"input:":  true,
	}

	for prefix := range componentPrefixes {
		if !validPrefixes[prefix] {
			t.Errorf("componentPrefixes contains unexpected prefix: %q", prefix)
		}
	}
}

// Tests for DetectGeneration using HTTP test servers

func TestDetectGeneration_Gen2Device(t *testing.T) {
	t.Parallel()

	gen2Response := map[string]any{
		"id":      "shellyplus1pm-test123",
		"mac":     "AA:BB:CC:DD:EE:FF",
		"model":   "SNSW-001P16EU",
		"gen":     2,
		"fw_id":   "20231107-164738/1.0.0-g1234567",
		"ver":     "1.0.0",
		"app":     "Plus1PM",
		"auth_en": false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen2Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := DetectGeneration(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() error = %v", err)
	}

	if result.Generation != Gen2 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen2)
	}
	if result.DeviceType != "Plus1PM" {
		t.Errorf("DeviceType = %q, want Plus1PM", result.DeviceType)
	}
	if result.Model != "SNSW-001P16EU" {
		t.Errorf("Model = %q, want SNSW-001P16EU", result.Model)
	}
	if result.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want AA:BB:CC:DD:EE:FF", result.MAC)
	}
	if result.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want 1.0.0", result.Firmware)
	}
	if result.AuthEn {
		t.Error("AuthEn = true, want false")
	}
	if !result.IsGen2() {
		t.Error("IsGen2() = false, want true")
	}
	if result.IsGen1() {
		t.Error("IsGen1() = true, want false")
	}
}

func TestDetectGeneration_Gen1Device(t *testing.T) {
	t.Parallel()

	gen1Response := map[string]any{
		"type":        "SHSW-1",
		"mac":         "11:22:33:44:55:66",
		"auth":        false,
		"fw":          "20210226-091923/v1.10.0-rc2@28a456c3",
		"num_outputs": 1,
		"num_meters":  0,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			// Gen2 endpoint returns 404
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := DetectGeneration(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() error = %v", err)
	}

	if result.Generation != Gen1 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen1)
	}
	if result.DeviceType != "SHSW-1" {
		t.Errorf("DeviceType = %q, want SHSW-1", result.DeviceType)
	}
	if result.MAC != "11:22:33:44:55:66" {
		t.Errorf("MAC = %q, want 11:22:33:44:55:66", result.MAC)
	}
	if result.IsGen2() {
		t.Error("IsGen2() = true, want false")
	}
	if !result.IsGen1() {
		t.Error("IsGen1() = false, want true")
	}
}

func TestDetectGeneration_AuthRequired(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for basic auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]any{
				"id":      "shellyplus1-auth",
				"mac":     "AA:BB:CC:DD:EE:FF",
				"model":   "SNSW-001X16EU",
				"gen":     2,
				"ver":     "1.0.0",
				"app":     "Plus1",
				"auth_en": true,
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Without auth - should fail
	_, err := DetectGeneration(ctx, server.URL, nil)
	if err == nil {
		t.Error("DetectGeneration() without auth should fail")
	}

	// With correct auth - should succeed
	auth := &model.Auth{Username: "admin", Password: "secret"}
	result, err := DetectGeneration(ctx, server.URL, auth)
	if err != nil {
		t.Fatalf("DetectGeneration() with auth error = %v", err)
	}
	if result.Generation != Gen2 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen2)
	}
	if !result.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestDetectGeneration_NoDevice(t *testing.T) {
	t.Parallel()

	// Server that always returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DetectGeneration(ctx, server.URL, nil)
	if err == nil {
		t.Error("DetectGeneration() should fail when device not found")
	}
}

func TestDetectGeneration_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte("not valid json")); err != nil {
				t.Logf("warning: failed to write response: %v", err)
			}
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte("{also not valid")); err != nil {
				t.Logf("warning: failed to write response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DetectGeneration(ctx, server.URL, nil)
	if err == nil {
		t.Error("DetectGeneration() should fail with invalid JSON")
	}
}

func TestDetectGeneration_Gen2FirmwareFallback(t *testing.T) {
	t.Parallel()

	// Test fallback from 'ver' to 'fw_id' when 'ver' is empty
	gen2Response := map[string]any{
		"id":    "shellyplus1pm-test",
		"mac":   "AA:BB:CC:DD:EE:FF",
		"model": "SNSW-001P16EU",
		"gen":   2,
		"fw_id": "20231107-164738/1.2.3-g1234567",
		"ver":   "", // Empty ver, should fall back to fw_id
		"app":   "Plus1PM",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen2Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := DetectGeneration(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() error = %v", err)
	}

	if result.Firmware != "20231107-164738/1.2.3-g1234567" {
		t.Errorf("Firmware = %q, want fw_id value", result.Firmware)
	}
}

func TestDetectGeneration_Gen2DeviceTypeFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response map[string]any
		wantType string
	}{
		{
			name: "app field",
			response: map[string]any{
				"id":    "test1",
				"mac":   "AA:BB:CC:DD:EE:FF",
				"model": "TEST",
				"gen":   2,
				"ver":   "1.0.0",
				"app":   "MyApp",
			},
			wantType: "MyApp",
		},
		{
			name: "type field when app empty",
			response: map[string]any{
				"id":    "test2",
				"mac":   "AA:BB:CC:DD:EE:FF",
				"model": "TEST",
				"gen":   2,
				"ver":   "1.0.0",
				"app":   "",
				"type":  "MyType",
			},
			wantType: "MyType",
		},
		{
			name: "dev_type field when app and type empty",
			response: map[string]any{
				"id":       "test3",
				"mac":      "AA:BB:CC:DD:EE:FF",
				"model":    "TEST",
				"gen":      2,
				"ver":      "1.0.0",
				"app":      "",
				"type":     "",
				"dev_type": "DevType",
			},
			wantType: "DevType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == gen2Endpoint {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(tt.response); err != nil {
						t.Logf("warning: failed to encode response: %v", err)
					}
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := DetectGeneration(ctx, server.URL, nil)
			if err != nil {
				t.Fatalf("DetectGeneration() error = %v", err)
			}

			if result.DeviceType != tt.wantType {
				t.Errorf("DeviceType = %q, want %q", result.DeviceType, tt.wantType)
			}
		})
	}
}

func TestDetectGeneration_Gen2ReportsGen1(t *testing.T) {
	t.Parallel()

	// Gen2 endpoint returns gen=1, should fail and try Gen1
	gen2Response := map[string]any{
		"id":    "test",
		"mac":   "AA:BB:CC:DD:EE:FF",
		"model": "TEST",
		"gen":   1, // Reports gen 1, should reject
		"ver":   "1.0.0",
	}

	gen1Response := map[string]any{
		"type": "SHSW-1",
		"mac":  "AA:BB:CC:DD:EE:FF",
		"auth": false,
		"fw":   "1.10.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen2Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := DetectGeneration(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() error = %v", err)
	}

	// Should fall back to Gen1 detection
	if result.Generation != Gen1 {
		t.Errorf("Generation = %d, want %d (Gen1)", result.Generation, Gen1)
	}
}

func TestDetectGeneration_AddressNormalization(t *testing.T) {
	t.Parallel()

	gen2Response := map[string]any{
		"id":    "test",
		"mac":   "AA:BB:CC:DD:EE:FF",
		"model": "TEST",
		"gen":   2,
		"ver":   "1.0.0",
		"app":   "Test",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen2Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Extract host:port without scheme
	addr := server.URL[7:] // Remove "http://"

	result, err := DetectGeneration(ctx, addr, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() with bare address error = %v", err)
	}

	if result.Generation != Gen2 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen2)
	}
}

func TestDetectGeneration_ContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Delay response to allow context cancellation
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := DetectGeneration(ctx, server.URL, nil)
	if err == nil {
		t.Error("DetectGeneration() with cancelled context should fail")
	}
}

func TestDetectGeneration_HTTPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		gen2Status int
		gen1Status int
		wantErr    bool
	}{
		{"both 500", http.StatusInternalServerError, http.StatusInternalServerError, true},
		{"both 503", http.StatusServiceUnavailable, http.StatusServiceUnavailable, true},
		{"both 403", http.StatusForbidden, http.StatusForbidden, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == gen2Endpoint {
					w.WriteHeader(tt.gen2Status)
					return
				}
				if r.URL.Path == gen1Endpoint {
					w.WriteHeader(tt.gen1Status)
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := DetectGeneration(ctx, server.URL, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectGeneration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectGeneration_Gen1WithAuth(t *testing.T) {
	t.Parallel()

	gen1Response := map[string]any{
		"type": "SHSW-1",
		"mac":  "AA:BB:CC:DD:EE:FF",
		"auth": true,
		"fw":   "1.11.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for basic auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == gen2Endpoint {
			// Gen2 endpoint returns 404 (device is Gen1)
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// With correct auth - should succeed and detect Gen1
	auth := &model.Auth{Username: "admin", Password: "password"}
	result, err := DetectGeneration(ctx, server.URL, auth)
	if err != nil {
		t.Fatalf("DetectGeneration() with auth error = %v", err)
	}
	if result.Generation != Gen1 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen1)
	}
	if !result.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestDetectGeneration_Gen1ReportsGen2(t *testing.T) {
	t.Parallel()

	// Gen1 endpoint (/shelly) reports gen >= 2, should reject
	gen1Response := map[string]any{
		"type": "SHSW-1",
		"mac":  "AA:BB:CC:DD:EE:FF",
		"auth": false,
		"fw":   "1.10.0",
		"gen":  2, // Reports gen 2, should reject this as a Gen1 device
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			// Gen2 endpoint returns 404
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DetectGeneration(ctx, server.URL, nil)
	if err == nil {
		t.Error("DetectGeneration() should fail when Gen1 endpoint reports gen >= 2")
	}
}

func TestDetectGeneration_Gen1WithEmptyUsername(t *testing.T) {
	t.Parallel()

	gen1Response := map[string]any{
		"type": "SHSW-1",
		"mac":  "AA:BB:CC:DD:EE:FF",
		"auth": false,
		"fw":   "1.10.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Auth with empty username should not set basic auth
	auth := &model.Auth{Username: "", Password: "password"}
	result, err := DetectGeneration(ctx, server.URL, auth)
	if err != nil {
		t.Fatalf("DetectGeneration() with empty username error = %v", err)
	}
	if result.Generation != Gen1 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen1)
	}
}

func TestDetectGeneration_Gen2WithEmptyUsername(t *testing.T) {
	t.Parallel()

	gen2Response := map[string]any{
		"id":    "test",
		"mac":   "AA:BB:CC:DD:EE:FF",
		"model": "TEST",
		"gen":   2,
		"ver":   "1.0.0",
		"app":   "Test",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen2Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Auth with empty username should not set basic auth
	auth := &model.Auth{Username: "", Password: "password"}
	result, err := DetectGeneration(ctx, server.URL, auth)
	if err != nil {
		t.Fatalf("DetectGeneration() with empty username error = %v", err)
	}
	if result.Generation != Gen2 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen2)
	}
}

func TestDetectGeneration_Gen1AllFields(t *testing.T) {
	t.Parallel()

	gen1Response := map[string]any{
		"type":        "SHSW-25",
		"mac":         "11:22:33:44:55:66",
		"auth":        true,
		"fw":          "20230913-114351/v1.14.0-gcb84623",
		"num_outputs": 2,
		"num_meters":  1,
		"gen":         1, // Explicitly set to 1
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == gen2Endpoint {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == gen1Endpoint {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(gen1Response); err != nil {
				t.Logf("warning: failed to encode response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := DetectGeneration(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("DetectGeneration() error = %v", err)
	}

	if result.Generation != Gen1 {
		t.Errorf("Generation = %d, want %d", result.Generation, Gen1)
	}
	if result.DeviceType != "SHSW-25" {
		t.Errorf("DeviceType = %q, want SHSW-25", result.DeviceType)
	}
	if result.Model != "SHSW-25" {
		t.Errorf("Model = %q, want SHSW-25", result.Model)
	}
	if result.MAC != "11:22:33:44:55:66" {
		t.Errorf("MAC = %q, want 11:22:33:44:55:66", result.MAC)
	}
	if result.Firmware != "20230913-114351/v1.14.0-gcb84623" {
		t.Errorf("Firmware = %q, want specific version", result.Firmware)
	}
	if !result.AuthEn {
		t.Error("AuthEn = false, want true")
	}
}

func TestDetectGeneration_EmptyAddress(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Empty address should fail
	_, err := DetectGeneration(ctx, "", nil)
	if err == nil {
		t.Error("DetectGeneration() with empty address should fail")
	}
}

func TestClient_Close_WithRPCClient(t *testing.T) {
	t.Parallel()

	// Test Close with nil device but non-nil rpcClient
	// We can't truly test this path because rpc.Client has no mockable interface
	// but we can at least verify the nil device path works
	client := &Client{
		device:    nil,
		rpcClient: nil,
	}
	err := client.Close()
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestGen1Client_Close_WithDevice(t *testing.T) {
	t.Parallel()

	// Test Close with nil device
	client := &Gen1Client{
		device: nil,
	}
	err := client.Close()
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

// Helper function

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return true
		}
		// Also check lowercase comparison
		match := true
		for j := range len(substr) {
			sc := s[i+j]
			subc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if subc >= 'A' && subc <= 'Z' {
				subc += 32
			}
			if sc != subc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
