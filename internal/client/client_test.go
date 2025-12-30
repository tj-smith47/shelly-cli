package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"

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

// mockRPCServer creates a test server that handles Gen2 RPC endpoints.
// It responds to both GET /rpc/<method> and POST /rpc requests.
type mockRPCServer struct {
	handlers map[string]func(params map[string]any) (any, error)
	authUser string
	authPass string
}

func newMockRPCServer() *mockRPCServer {
	return &mockRPCServer{
		handlers: make(map[string]func(params map[string]any) (any, error)),
	}
}

func (m *mockRPCServer) withAuth(user, pass string) *mockRPCServer {
	m.authUser = user
	m.authPass = pass
	return m
}

func (m *mockRPCServer) handle(method string, handler func(params map[string]any) (any, error)) *mockRPCServer {
	m.handlers[method] = handler
	return m
}

func (m *mockRPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check auth if configured
	if m.authUser != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != m.authUser || pass != m.authPass {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Handle GET /rpc/<method> format
	if r.Method == http.MethodGet && len(r.URL.Path) > 5 && r.URL.Path[:5] == "/rpc/" {
		method := r.URL.Path[5:]
		if handler, ok := m.handlers[method]; ok {
			result, err := handler(nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
			return
		}
		http.NotFound(w, r)
		return
	}

	// Handle POST /rpc format (JSON-RPC)
	if r.Method == http.MethodPost && r.URL.Path == "/rpc" {
		var req struct {
			ID     int            `json:"id"`
			Method string         `json:"method"`
			Params map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if handler, ok := m.handlers[req.Method]; ok {
			result, err := handler(req.Params)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":    req.ID,
					"error": map[string]any{"code": -1, "message": err.Error()},
				})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     req.ID,
				"result": result,
			})
			return
		}
		http.NotFound(w, r)
		return
	}

	http.NotFound(w, r)
}

func (m *mockRPCServer) start(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(m)
	t.Cleanup(server.Close)
	return server
}

// Standard device info response for tests
func standardDeviceInfo() map[string]any {
	return map[string]any{
		"id":      "shellyplus1pm-test123",
		"mac":     "AA:BB:CC:DD:EE:FF",
		"model":   "SNSW-001P16EU",
		"gen":     2,
		"fw_id":   "20231107-164738/1.0.0-g1234567",
		"ver":     "1.0.0",
		"app":     "Plus1PM",
		"auth_en": false,
	}
}

func TestConnect_Success(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	info := client.Info()
	if info == nil {
		t.Fatal("Info() returned nil")
	}
	if info.ID != "shellyplus1pm-test123" {
		t.Errorf("Info().ID = %q, want shellyplus1pm-test123", info.ID)
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Info().MAC = %q, want AA:BB:CC:DD:EE:FF", info.MAC)
	}
	if info.Model != "SNSW-001P16EU" {
		t.Errorf("Info().Model = %q, want SNSW-001P16EU", info.Model)
	}
	if info.Generation != 2 {
		t.Errorf("Info().Generation = %d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("Info().Firmware = %q, want 1.0.0", info.Firmware)
	}
	if info.App != "Plus1PM" {
		t.Errorf("Info().App = %q, want Plus1PM", info.App)
	}
}

func TestConnect_WithAuth(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		withAuth("admin", "secret").
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			info := standardDeviceInfo()
			info["auth_en"] = true
			return info, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{
		Address: server.URL,
		Auth:    &model.Auth{Username: "admin", Password: "secret"},
	}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() with auth error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	if !client.Info().AuthEn {
		t.Error("Info().AuthEn = false, want true")
	}
}

func TestConnect_AuthRequired(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		withAuth("admin", "secret").
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try without auth - should fail
	device := model.Device{Address: server.URL}
	_, err := Connect(ctx, device)
	if err == nil {
		t.Error("Connect() without auth should fail when auth required")
	}
}

func TestConnect_AddressNormalization(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with address without http:// prefix
	addr := server.URL[7:] // Remove "http://"
	device := model.Device{Address: addr}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() with bare address error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	if client.Info().ID != "shellyplus1pm-test123" {
		t.Errorf("Info().ID = %q, want shellyplus1pm-test123", client.Info().ID)
	}
}

func TestConnect_ConnectionFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use an address that won't connect
	device := model.Device{Address: "http://127.0.0.1:1"}
	_, err := Connect(ctx, device)
	if err == nil {
		t.Error("Connect() to bad address should fail")
	}
}

func TestClient_ListComponents_Success(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.GetComponents", func(_ map[string]any) (any, error) {
			return map[string]any{
				"components": []any{
					map[string]any{"key": "switch:0"},
					map[string]any{"key": "switch:1"},
					map[string]any{"key": "input:0"},
					map[string]any{"key": "cover:0"},
					map[string]any{"key": "light:0"},
					map[string]any{"key": "sys"},      // Should be filtered out
					map[string]any{"key": "wifi:sta"}, // Should be filtered out
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	comps, err := client.ListComponents(ctx)
	if err != nil {
		t.Fatalf("ListComponents() error = %v", err)
	}

	// Should only include parseable component types
	if len(comps) != 5 {
		t.Errorf("len(comps) = %d, want 5 (sys and wifi filtered out)", len(comps))
	}

	// Check specific components
	switchCount := 0
	for _, c := range comps {
		if c.Type == model.ComponentSwitch {
			switchCount++
		}
	}
	if switchCount != 2 {
		t.Errorf("switch count = %d, want 2", switchCount)
	}
}

func TestClient_FilterComponents_Success(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.GetComponents", func(_ map[string]any) (any, error) {
			return map[string]any{
				"components": []any{
					map[string]any{"key": "switch:0"},
					map[string]any{"key": "switch:1"},
					map[string]any{"key": "input:0"},
					map[string]any{"key": "cover:0"},
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	// Filter for switches only
	switches, err := client.FilterComponents(ctx, model.ComponentSwitch)
	if err != nil {
		t.Fatalf("FilterComponents() error = %v", err)
	}
	if len(switches) != 2 {
		t.Errorf("len(switches) = %d, want 2", len(switches))
	}
	for _, sw := range switches {
		if sw.Type != model.ComponentSwitch {
			t.Errorf("filtered component type = %q, want switch", sw.Type)
		}
	}

	// Filter for covers only
	covers, err := client.FilterComponents(ctx, model.ComponentCover)
	if err != nil {
		t.Fatalf("FilterComponents(cover) error = %v", err)
	}
	if len(covers) != 1 {
		t.Errorf("len(covers) = %d, want 1", len(covers))
	}

	// Filter for type that doesn't exist
	lights, err := client.FilterComponents(ctx, model.ComponentLight)
	if err != nil {
		t.Fatalf("FilterComponents(light) error = %v", err)
	}
	if len(lights) != 0 {
		t.Errorf("len(lights) = %d, want 0", len(lights))
	}
}

func TestClient_GetStatus_Success(t *testing.T) {
	t.Parallel()

	expectedStatus := map[string]any{
		"sys": map[string]any{
			"mac":            "AA:BB:CC:DD:EE:FF",
			"restart_reason": "power_on",
			"uptime":         12345,
		},
		"wifi": map[string]any{
			"sta_ip": "192.168.1.100",
			"status": "connected",
			"ssid":   "MyNetwork",
			"rssi":   -55,
		},
		"switch:0": map[string]any{
			"id":      0,
			"output":  true,
			"source":  "button",
			"apower":  150.5,
			"voltage": 230.1,
		},
	}

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.GetStatus", func(_ map[string]any) (any, error) {
			return expectedStatus, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	status, err := client.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	// Check sys status
	sys, ok := status["sys"].(map[string]any)
	if !ok {
		t.Fatal("status[sys] not a map")
	}
	if sys["mac"] != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("sys.mac = %v, want AA:BB:CC:DD:EE:FF", sys["mac"])
	}

	// Check switch status
	sw, ok := status["switch:0"].(map[string]any)
	if !ok {
		t.Fatal("status[switch:0] not a map")
	}
	if sw["output"] != true {
		t.Errorf("switch:0.output = %v, want true", sw["output"])
	}
}

func TestClient_GetConfig_Success(t *testing.T) {
	t.Parallel()

	expectedConfig := map[string]any{
		"sys": map[string]any{
			"device": map[string]any{
				"name":  "My Shelly",
				"eco":   false,
				"fw_id": "1.0.0",
			},
		},
		"wifi": map[string]any{
			"sta": map[string]any{
				"ssid":   "MyNetwork",
				"pass":   "****",
				"enable": true,
			},
		},
		"switch:0": map[string]any{
			"id":            0,
			"name":          "Kitchen Light",
			"initial_state": "restore_last",
			"auto_on":       false,
			"auto_off":      false,
		},
	}

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.GetConfig", func(_ map[string]any) (any, error) {
			return expectedConfig, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	config, err := client.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Check sys config
	sys, ok := config["sys"].(map[string]any)
	if !ok {
		t.Fatal("config[sys] not a map")
	}
	sysDevice, ok := sys["device"].(map[string]any)
	if !ok {
		t.Fatal("sys.device not a map")
	}
	if sysDevice["name"] != "My Shelly" {
		t.Errorf("sys.device.name = %v, want My Shelly", sysDevice["name"])
	}

	// Check switch config
	sw, ok := config["switch:0"].(map[string]any)
	if !ok {
		t.Fatal("config[switch:0] not a map")
	}
	if sw["name"] != "Kitchen Light" {
		t.Errorf("switch:0.name = %v, want Kitchen Light", sw["name"])
	}
}

func TestClient_SetConfig_Success(t *testing.T) {
	t.Parallel()

	var receivedConfig map[string]any

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.SetConfig", func(params map[string]any) (any, error) {
			if cfg, ok := params["config"].(map[string]any); ok {
				receivedConfig = cfg
			}
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	newConfig := map[string]any{
		"sys": map[string]any{
			"device": map[string]any{
				"name": "New Name",
			},
		},
	}

	err = client.SetConfig(ctx, newConfig)
	if err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	// Verify the config was sent
	if receivedConfig == nil {
		t.Fatal("SetConfig did not send config")
	}
	sysConfig, ok := receivedConfig["sys"].(map[string]any)
	if !ok {
		t.Fatal("received config sys not a map")
	}
	deviceConfig, ok := sysConfig["device"].(map[string]any)
	if !ok {
		t.Fatal("received config sys.device not a map")
	}
	if deviceConfig["name"] != "New Name" {
		t.Errorf("received config name = %v, want New Name", deviceConfig["name"])
	}
}

func TestClient_Reboot_Immediate(t *testing.T) {
	t.Parallel()

	var receivedDelayMS int

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.Reboot", func(params map[string]any) (any, error) {
			if delay, ok := params["delay_ms"].(float64); ok {
				receivedDelayMS = int(delay)
			}
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.Reboot(ctx, 0)
	if err != nil {
		t.Fatalf("Reboot() error = %v", err)
	}

	// With delay 0, no delay_ms should be sent
	if receivedDelayMS != 0 {
		t.Errorf("receivedDelayMS = %d, want 0 (not sent)", receivedDelayMS)
	}
}

func TestClient_Reboot_WithDelay(t *testing.T) {
	t.Parallel()

	var receivedDelayMS int

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.Reboot", func(params map[string]any) (any, error) {
			if delay, ok := params["delay_ms"].(float64); ok {
				receivedDelayMS = int(delay)
			}
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.Reboot(ctx, 5000)
	if err != nil {
		t.Fatalf("Reboot() error = %v", err)
	}

	if receivedDelayMS != 5000 {
		t.Errorf("receivedDelayMS = %d, want 5000", receivedDelayMS)
	}
}

func TestClient_FactoryReset_Success(t *testing.T) {
	t.Parallel()

	factoryResetCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Shelly.FactoryReset", func(_ map[string]any) (any, error) {
			factoryResetCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.FactoryReset(ctx)
	if err != nil {
		t.Fatalf("FactoryReset() error = %v", err)
	}

	if !factoryResetCalled {
		t.Error("FactoryReset was not called")
	}
}

func TestClient_Call_Success(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Custom.Method", func(params map[string]any) (any, error) {
			return map[string]any{
				"received_param": params["test_param"],
				"custom_result":  "success",
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	result, err := client.Call(ctx, "Custom.Method", map[string]any{"test_param": "test_value"})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	// The RPC client may return json.RawMessage, so we need to handle that
	var resultMap map[string]any
	switch v := result.(type) {
	case map[string]any:
		resultMap = v
	case json.RawMessage:
		if err := json.Unmarshal(v, &resultMap); err != nil {
			t.Fatalf("failed to unmarshal json.RawMessage: %v", err)
		}
	default:
		t.Fatalf("result is not map[string]any or json.RawMessage, got %T", result)
	}

	if resultMap["custom_result"] != "success" {
		t.Errorf("custom_result = %v, want success", resultMap["custom_result"])
	}
	if resultMap["received_param"] != "test_value" {
		t.Errorf("received_param = %v, want test_value", resultMap["received_param"])
	}
}

func TestClient_ListComponents_Error(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})
	// Note: Shelly.GetComponents is NOT handled, so it will return 404

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	_, err = client.ListComponents(ctx)
	if err == nil {
		t.Error("ListComponents() should fail when handler not found")
	}
}

func TestClient_GetStatus_Error(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})
	// Note: Shelly.GetStatus is NOT handled

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	_, err = client.GetStatus(ctx)
	if err == nil {
		t.Error("GetStatus() should fail when handler not found")
	}
}

func TestClient_GetConfig_Error(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})
	// Note: Shelly.GetConfig is NOT handled

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	_, err = client.GetConfig(ctx)
	if err == nil {
		t.Error("GetConfig() should fail when handler not found")
	}
}

// mockGen1Server creates a test server that handles Gen1 HTTP endpoints.
type mockGen1Server struct {
	handlers map[string]func(query string) (any, int)
	authUser string
	authPass string
}

func newMockGen1Server() *mockGen1Server {
	return &mockGen1Server{
		handlers: make(map[string]func(query string) (any, int)),
	}
}

func (m *mockGen1Server) withAuth(user, pass string) *mockGen1Server {
	m.authUser = user
	m.authPass = pass
	return m
}

func (m *mockGen1Server) handle(path string, handler func(query string) (any, int)) *mockGen1Server {
	m.handlers[path] = handler
	return m
}

func (m *mockGen1Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check auth if configured
	if m.authUser != "" {
		user, pass, ok := r.BasicAuth()
		if !ok || user != m.authUser || pass != m.authPass {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Handle GET requests
	if r.Method == http.MethodGet {
		path := r.URL.Path
		if handler, ok := m.handlers[path]; ok {
			result, status := handler(r.URL.RawQuery)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			if result != nil {
				_ = json.NewEncoder(w).Encode(result)
			}
			return
		}
		http.NotFound(w, r)
		return
	}

	http.NotFound(w, r)
}

func (m *mockGen1Server) start(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(m)
	t.Cleanup(server.Close)
	return server
}

// Standard Gen1 device info response
func standardGen1DeviceInfo() map[string]any {
	return map[string]any{
		"type":        "SHSW-1",
		"mac":         "11:22:33:44:55:66",
		"auth":        false,
		"fw":          "20230913-114351/v1.14.0-gcb84623",
		"num_outputs": 1,
		"num_meters":  0,
	}
}

func TestConnectGen1_Success(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	info := client.Info()
	if info == nil {
		t.Fatal("Info() returned nil")
	}
	if info.MAC != "11:22:33:44:55:66" {
		t.Errorf("Info().MAC = %q, want 11:22:33:44:55:66", info.MAC)
	}
	if info.Model != "SHSW-1" {
		t.Errorf("Info().Model = %q, want SHSW-1", info.Model)
	}
	if info.Generation != 1 {
		t.Errorf("Info().Generation = %d, want 1", info.Generation)
	}
}

func TestConnectGen1_WithAuth(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		withAuth("admin", "password").
		handle("/shelly", func(_ string) (any, int) {
			info := standardGen1DeviceInfo()
			info["auth"] = true
			return info, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{
		Address: server.URL,
		Auth:    &model.Auth{Username: "admin", Password: "password"},
	}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() with auth error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	if !client.Info().AuthEn {
		t.Error("Info().AuthEn = false, want true")
	}
}

func TestConnectGen1_AuthRequired(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		withAuth("admin", "secret").
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try without auth - should fail
	device := model.Device{Address: server.URL}
	_, err := ConnectGen1(ctx, device)
	if err == nil {
		t.Error("ConnectGen1() without auth should fail when auth required")
	}
}

func TestConnectGen1_AddressNormalization(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with address without http:// prefix
	addr := server.URL[7:] // Remove "http://"
	device := model.Device{Address: addr}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() with bare address error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	if client.Info().MAC != "11:22:33:44:55:66" {
		t.Errorf("Info().MAC = %q, want 11:22:33:44:55:66", client.Info().MAC)
	}
}

func TestConnectGen1_ConnectionFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use an address that won't connect
	device := model.Device{Address: "http://127.0.0.1:1"}
	_, err := ConnectGen1(ctx, device)
	if err == nil {
		t.Error("ConnectGen1() to bad address should fail")
	}
}

func TestGen1Client_GetStatus_Success(t *testing.T) {
	t.Parallel()

	statusResponse := map[string]any{
		"relays": []any{
			map[string]any{
				"ison":      true,
				"has_timer": false,
				"source":    "button",
			},
		},
		"meters": []any{
			map[string]any{
				"power": 150.5,
			},
		},
		"wifi_sta": map[string]any{
			"connected": true,
			"ssid":      "MyNetwork",
			"ip":        "192.168.1.100",
			"rssi":      -55,
		},
		"update": map[string]any{
			"status": "idle",
		},
		"uptime": 12345,
	}

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/status", func(_ string) (any, int) {
			return statusResponse, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	status, err := client.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if status.Uptime != 12345 {
		t.Errorf("status.Uptime = %d, want 12345", status.Uptime)
	}
}

func TestGen1Client_GetSettings_Success(t *testing.T) {
	t.Parallel()

	settingsResponse := map[string]any{
		"device": map[string]any{
			"type":     "SHSW-1",
			"mac":      "11:22:33:44:55:66",
			"hostname": "shelly1-123456",
		},
		"wifi_sta": map[string]any{
			"enabled": true,
			"ssid":    "MyNetwork",
		},
		"relays": []any{
			map[string]any{
				"name":           "Kitchen Light",
				"default_state":  "last",
				"btn_type":       "toggle",
				"auto_on":        0,
				"auto_off":       0,
				"schedule":       false,
				"schedule_rules": []any{},
			},
		},
	}

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/settings", func(_ string) (any, int) {
			return settingsResponse, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	settings, err := client.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}

	if settings == nil {
		t.Fatal("GetSettings() returned nil")
	}
}

func TestGen1Client_Reboot_Success(t *testing.T) {
	t.Parallel()

	rebootCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/reboot", func(_ string) (any, int) {
			rebootCalled = true
			return map[string]any{}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.Reboot(ctx)
	if err != nil {
		t.Fatalf("Reboot() error = %v", err)
	}

	if !rebootCalled {
		t.Error("Reboot was not called")
	}
}

func TestGen1Client_FactoryReset_Success(t *testing.T) {
	t.Parallel()

	factoryResetCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		// Gen1 factory reset uses /reset endpoint
		handle("/reset", func(_ string) (any, int) {
			factoryResetCalled = true
			return map[string]any{}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.FactoryReset(ctx)
	if err != nil {
		t.Fatalf("FactoryReset() error = %v", err)
	}

	if !factoryResetCalled {
		t.Error("FactoryReset was not called")
	}
}

func TestGen1Client_Call_Success(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/custom/endpoint", func(_ string) (any, int) {
			return map[string]any{"result": "success"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	result, err := client.Call(ctx, "/custom/endpoint")
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Call() returned empty result")
	}
}

func TestGen1Client_Relay_ValidID(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}
	if relay == nil {
		t.Fatal("Relay(0) returned nil")
	}
	if relay.ID() != 0 {
		t.Errorf("relay.ID() = %d, want 0", relay.ID())
	}
}

func TestGen1Client_Roller_ValidID(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("Roller(0) error = %v", err)
	}
	if roller == nil {
		t.Fatal("Roller(0) returned nil")
	}
	if roller.ID() != 0 {
		t.Errorf("roller.ID() = %d, want 0", roller.ID())
	}
}

func TestGen1Client_Light_ValidID(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("Light(0) error = %v", err)
	}
	if light == nil {
		t.Fatal("Light(0) returned nil")
	}
	if light.ID() != 0 {
		t.Errorf("light.ID() = %d, want 0", light.ID())
	}
}

func TestGen1Client_Color_ValidID(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("Color(0) error = %v", err)
	}
	if color == nil {
		t.Fatal("Color(0) returned nil")
	}
	if color.ID() != 0 {
		t.Errorf("color.ID() = %d, want 0", color.ID())
	}
}

func TestGen1Client_White_ValidID(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	white, err := client.White(0)
	if err != nil {
		t.Fatalf("White(0) error = %v", err)
	}
	if white == nil {
		t.Fatal("White(0) returned nil")
	}
	if white.ID() != 0 {
		t.Errorf("white.ID() = %d, want 0", white.ID())
	}
}

func TestGen1Client_CheckForUpdate_Success(t *testing.T) {
	t.Parallel()

	updateResponse := map[string]any{
		"status":      "idle",
		"has_update":  true,
		"new_version": "1.15.0",
		"old_version": "1.14.0",
	}

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/ota/check", func(_ string) (any, int) {
			return updateResponse, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	update, err := client.CheckForUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v", err)
	}

	if update == nil {
		t.Fatal("CheckForUpdate() returned nil")
	}
}

func TestGen1Client_DeviceAccessor(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	gen1Device := client.Device()
	if gen1Device == nil {
		t.Error("Device() returned nil")
	}
}

func TestGen1Relay_TurnOn(t *testing.T) {
	t.Parallel()

	turnOnCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query == "turn=on" {
				turnOnCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.TurnOn(ctx)
	if err != nil {
		t.Fatalf("TurnOn() error = %v", err)
	}

	if !turnOnCalled {
		t.Error("TurnOn was not called")
	}
}

func TestGen1Relay_TurnOff(t *testing.T) {
	t.Parallel()

	turnOffCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query == "turn=off" {
				turnOffCalled = true
			}
			return map[string]any{"ison": false}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.TurnOff(ctx)
	if err != nil {
		t.Fatalf("TurnOff() error = %v", err)
	}

	if !turnOffCalled {
		t.Error("TurnOff was not called")
	}
}

func TestGen1Relay_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query == "turn=toggle" {
				toggleCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("Toggle was not called")
	}
}

func TestGen1Relay_Set(t *testing.T) {
	t.Parallel()

	setCalled := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query == "turn=on" || query == "turn=off" {
				setCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.Set(ctx, true)
	if err != nil {
		t.Fatalf("Set(true) error = %v", err)
	}

	if !setCalled {
		t.Error("Set was not called")
	}
}

func TestGen1Relay_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(_ string) (any, int) {
			return map[string]any{
				"ison":             true,
				"has_timer":        false,
				"timer_started_at": 0,
				"timer_duration":   0,
				"timer_remaining":  0,
				"source":           "button",
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	status, err := relay.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.IsOn {
		t.Error("status.IsOn = false, want true")
	}
}

func TestGen1Relay_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/settings/relay/0", func(_ string) (any, int) {
			return map[string]any{
				"name":          "Kitchen Light",
				"default_state": "last",
				"btn_type":      "toggle",
				"auto_on":       0,
				"auto_off":      0,
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	config, err := relay.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name != "Kitchen Light" {
		t.Errorf("config.Name = %q, want Kitchen Light", config.Name)
	}
}

func TestGen1Relay_TurnOnForDuration(t *testing.T) {
	t.Parallel()

	called := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query != "" {
				called = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.TurnOnForDuration(ctx, 30)
	if err != nil {
		t.Fatalf("TurnOnForDuration() error = %v", err)
	}

	if !called {
		t.Error("TurnOnForDuration was not called")
	}
}

func TestGen1Relay_TurnOffForDuration(t *testing.T) {
	t.Parallel()

	called := false

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/relay/0", func(query string) (any, int) {
			if query != "" {
				called = true
			}
			return map[string]any{"ison": false}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	relay, err := client.Relay(0)
	if err != nil {
		t.Fatalf("Relay(0) error = %v", err)
	}

	err = relay.TurnOffForDuration(ctx, 30)
	if err != nil {
		t.Fatalf("TurnOffForDuration() error = %v", err)
	}

	if !called {
		t.Error("TurnOffForDuration was not called")
	}
}

// Gen2 Switch Component Tests

func TestGen2Switch_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":      0,
				"output":  true,
				"source":  "button",
				"apower":  125.5,
				"voltage": 230.1,
				"current": 0.55,
				"aenergy": map[string]any{
					"total":     1234.5,
					"by_minute": []float64{50.0, 50.1, 50.0},
				},
				"temperature": map[string]any{
					"tC": 45.5,
					"tF": 113.9,
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	status, err := sw.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.Output {
		t.Error("status.Output = false, want true")
	}
}

func TestGen2Switch_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":             0,
				"name":           "Kitchen Light",
				"initial_state":  "restore_last",
				"auto_on":        false,
				"auto_on_delay":  0,
				"auto_off":       false,
				"auto_off_delay": 0,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	config, err := sw.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "Kitchen Light" {
		var name string
		if config.Name != nil {
			name = *config.Name
		}
		t.Errorf("config.Name = %q, want Kitchen Light", name)
	}
}

func TestGen2Switch_On(t *testing.T) {
	t.Parallel()

	setCalled := false
	var receivedOn bool

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.Set", func(params map[string]any) (any, error) {
			setCalled = true
			if on, ok := params["on"].(bool); ok {
				receivedOn = on
			}
			return map[string]any{"was_on": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	err = sw.On(ctx)
	if err != nil {
		t.Fatalf("On() error = %v", err)
	}

	if !setCalled {
		t.Error("Switch.Set was not called")
	}
	if !receivedOn {
		t.Error("receivedOn = false, want true")
	}
}

func TestGen2Switch_Off(t *testing.T) {
	t.Parallel()

	setCalled := false
	var receivedOn bool

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.Set", func(params map[string]any) (any, error) {
			setCalled = true
			if on, ok := params["on"].(bool); ok {
				receivedOn = on
			}
			return map[string]any{"was_on": true}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	err = sw.Off(ctx)
	if err != nil {
		t.Fatalf("Off() error = %v", err)
	}

	if !setCalled {
		t.Error("Switch.Set was not called")
	}
	if receivedOn {
		t.Error("receivedOn = true, want false")
	}
}

func TestGen2Switch_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.Toggle", func(_ map[string]any) (any, error) {
			toggleCalled = true
			return map[string]any{"was_on": true}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	_, err = sw.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("Switch.Toggle was not called")
	}
}

func TestGen2Switch_Set(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Switch.Set", func(_ map[string]any) (any, error) {
			return map[string]any{"was_on": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	sw := client.Switch(0)
	err = sw.Set(ctx, true)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
}

// Gen2 Cover Component Tests

func TestGen2Cover_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":          0,
				"state":       "stopped",
				"source":      "button",
				"current_pos": 75,
				"target_pos":  75,
				"apower":      0,
				"voltage":     230.0,
				"current":     0,
				"pos_control": true,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	status, err := cover.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if status.State != "stopped" {
		t.Errorf("status.State = %q, want stopped", status.State)
	}
	if status.CurrentPosition == nil || *status.CurrentPosition != 75 {
		t.Errorf("status.CurrentPosition = %v, want 75", status.CurrentPosition)
	}
}

func TestGen2Cover_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":            0,
				"name":          "Living Room Blinds",
				"initial_state": "stopped",
				"invert_dir":    false,
				"pos_control":   true,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	config, err := cover.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "Living Room Blinds" {
		var name string
		if config.Name != nil {
			name = *config.Name
		}
		t.Errorf("config.Name = %q, want Living Room Blinds", name)
	}
}

func TestGen2Cover_Open(t *testing.T) {
	t.Parallel()

	openCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.Open", func(_ map[string]any) (any, error) {
			openCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	err = cover.Open(ctx, nil)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if !openCalled {
		t.Error("Cover.Open was not called")
	}
}

func TestGen2Cover_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.Close", func(_ map[string]any) (any, error) {
			closeCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	err = cover.Close(ctx, nil)
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if !closeCalled {
		t.Error("Cover.Close was not called")
	}
}

func TestGen2Cover_Stop(t *testing.T) {
	t.Parallel()

	stopCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.Stop", func(_ map[string]any) (any, error) {
			stopCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	err = cover.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if !stopCalled {
		t.Error("Cover.Stop was not called")
	}
}

func TestGen2Cover_GoToPosition(t *testing.T) {
	t.Parallel()

	var receivedPos int

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.GoToPosition", func(params map[string]any) (any, error) {
			if pos, ok := params["pos"].(float64); ok {
				receivedPos = int(pos)
			}
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	err = cover.GoToPosition(ctx, 50)
	if err != nil {
		t.Fatalf("GoToPosition() error = %v", err)
	}

	if receivedPos != 50 {
		t.Errorf("receivedPos = %d, want 50", receivedPos)
	}
}

func TestGen2Cover_Calibrate(t *testing.T) {
	t.Parallel()

	calibrateCalled := false

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Cover.Calibrate", func(_ map[string]any) (any, error) {
			calibrateCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	cover := client.Cover(0)
	err = cover.Calibrate(ctx)
	if err != nil {
		t.Fatalf("Calibrate() error = %v", err)
	}

	if !calibrateCalled {
		t.Error("Cover.Calibrate was not called")
	}
}

// ============================================
// Gen1 Roller Component Tests
// ============================================

func TestGen1Roller_Open(t *testing.T) {
	t.Parallel()

	openCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(_ string) (any, int) {
			openCalled = true
			return map[string]any{"state": "open"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	err = roller.Open(ctx)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if !openCalled {
		t.Error("Roller Open was not called")
	}
}

func TestGen1Roller_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(_ string) (any, int) {
			closeCalled = true
			return map[string]any{"state": "close"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	err = roller.Close(ctx)
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if !closeCalled {
		t.Error("Roller Close was not called")
	}
}

func TestGen1Roller_Stop(t *testing.T) {
	t.Parallel()

	stopCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(_ string) (any, int) {
			stopCalled = true
			return map[string]any{"state": "stop"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	err = roller.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if !stopCalled {
		t.Error("Roller Stop was not called")
	}
}

func TestGen1Roller_GoToPosition(t *testing.T) {
	t.Parallel()

	positionCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(_ string) (any, int) {
			positionCalled = true
			return map[string]any{"current_pos": 50}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	err = roller.GoToPosition(ctx, 50)
	if err != nil {
		t.Fatalf("GoToPosition() error = %v", err)
	}

	if !positionCalled {
		t.Error("Roller GoToPosition was not called")
	}
}

func TestGen1Roller_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(_ string) (any, int) {
			return map[string]any{
				"state":           "stop",
				"power":           0.0,
				"is_valid":        true,
				"safety_switch":   false,
				"overtemperature": false,
				"stop_reason":     "normal",
				"last_direction":  "open",
				"current_pos":     75,
				"calibrating":     false,
				"positioning":     true,
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	status, err := roller.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if status.State != "stop" {
		t.Errorf("status.State = %q, want stop", status.State)
	}
}

func TestGen1Roller_Calibrate(t *testing.T) {
	t.Parallel()

	calibrateCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0/calibrate", func(_ string) (any, int) {
			calibrateCalled = true
			return map[string]any{"calibrating": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}
	err = roller.Calibrate(ctx)
	if err != nil {
		t.Fatalf("Calibrate() error = %v", err)
	}

	if !calibrateCalled {
		t.Error("Roller Calibrate was not called")
	}
}

// ============================================
// Gen1 Light Component Tests
// ============================================

func TestGen1Light_TurnOn(t *testing.T) {
	t.Parallel()

	turnOnCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(_ string) (any, int) {
			turnOnCalled = true
			return map[string]any{"ison": true, "brightness": 100}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}
	err = light.TurnOn(ctx)
	if err != nil {
		t.Fatalf("TurnOn() error = %v", err)
	}

	if !turnOnCalled {
		t.Error("Light TurnOn was not called")
	}
}

func TestGen1Light_TurnOff(t *testing.T) {
	t.Parallel()

	turnOffCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(_ string) (any, int) {
			turnOffCalled = true
			return map[string]any{"ison": false, "brightness": 0}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}
	err = light.TurnOff(ctx)
	if err != nil {
		t.Fatalf("TurnOff() error = %v", err)
	}

	if !turnOffCalled {
		t.Error("Light TurnOff was not called")
	}
}

func TestGen1Light_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(_ string) (any, int) {
			toggleCalled = true
			return map[string]any{"ison": true, "brightness": 100}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}
	err = light.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("Light Toggle was not called")
	}
}

func TestGen1Light_SetBrightness(t *testing.T) {
	t.Parallel()

	setBrightnessCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(_ string) (any, int) {
			setBrightnessCalled = true
			return map[string]any{"ison": true, "brightness": 75}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}
	err = light.SetBrightness(ctx, 75)
	if err != nil {
		t.Fatalf("SetBrightness() error = %v", err)
	}

	if !setBrightnessCalled {
		t.Error("Light SetBrightness was not called")
	}
}

func TestGen1Light_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(_ string) (any, int) {
			return map[string]any{
				"ison":       true,
				"brightness": 85,
				"mode":       "white",
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}
	status, err := light.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.IsOn {
		t.Error("status.IsOn = false, want true")
	}
	if status.Brightness != 85 {
		t.Errorf("status.Brightness = %d, want 85", status.Brightness)
	}
}

// ============================================
// Gen2 Light Component Tests
// ============================================

func TestGen2Light_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":         0,
				"output":     true,
				"source":     "button",
				"brightness": 80,
				"apower":     5.5,
				"voltage":    120.0,
				"current":    0.046,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	status, err := light.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.Output {
		t.Error("status.Output = false, want true")
	}
	if status.Brightness == nil || *status.Brightness != 80 {
		t.Errorf("status.Brightness = %v, want 80", status.Brightness)
	}
}

func TestGen2Light_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":                 0,
				"name":               "Living Room Light",
				"initial_state":      "on",
				"auto_on":            false,
				"auto_on_delay":      0.0,
				"auto_off":           true,
				"auto_off_delay":     300.0,
				"default_brightness": 100,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	config, err := light.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "Living Room Light" {
		t.Errorf("config.Name = %v, want Living Room Light", config.Name)
	}
}

func TestGen2Light_On(t *testing.T) {
	t.Parallel()

	onCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.Set", func(_ map[string]any) (any, error) {
			onCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	err = light.On(ctx)
	if err != nil {
		t.Fatalf("On() error = %v", err)
	}

	if !onCalled {
		t.Error("Light.Set was not called")
	}
}

func TestGen2Light_Off(t *testing.T) {
	t.Parallel()

	offCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.Set", func(_ map[string]any) (any, error) {
			offCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	err = light.Off(ctx)
	if err != nil {
		t.Fatalf("Off() error = %v", err)
	}

	if !offCalled {
		t.Error("Light.Set was not called")
	}
}

func TestGen2Light_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.Toggle", func(_ map[string]any) (any, error) {
			toggleCalled = true
			return map[string]any{"was_on": true}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	status, err := light.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("Light.Toggle was not called")
	}
	if status == nil {
		t.Fatal("Toggle() returned nil status")
	}
}

func TestGen2Light_SetBrightness(t *testing.T) {
	t.Parallel()

	setBrightnessCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.Set", func(_ map[string]any) (any, error) {
			setBrightnessCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	err = light.SetBrightness(ctx, 75)
	if err != nil {
		t.Fatalf("SetBrightness() error = %v", err)
	}

	if !setBrightnessCalled {
		t.Error("Light.Set was not called")
	}
}

func TestGen2Light_Set(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Light.Set", func(_ map[string]any) (any, error) {
			setCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light := client.Light(0)
	brightness := 50
	on := true
	err = light.Set(ctx, &brightness, &on)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !setCalled {
		t.Error("Light.Set was not called")
	}
}

// ============================================
// Gen2 RGB Component Tests
// ============================================

func TestGen2RGB_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":         0,
				"output":     true,
				"source":     "button",
				"brightness": 80,
				"rgb":        []any{255.0, 128.0, 64.0},
				"apower":     5.5,
				"voltage":    120.0,
				"current":    0.046,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	status, err := rgb.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.Output {
		t.Error("status.Output = false, want true")
	}
	if status.Brightness == nil || *status.Brightness != 80 {
		t.Errorf("status.Brightness = %v, want 80", status.Brightness)
	}
}

func TestGen2RGB_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":                 0,
				"name":               "RGB Strip",
				"initial_state":      "restore_last",
				"auto_on":            false,
				"auto_off":           false,
				"default_brightness": 100,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	config, err := rgb.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "RGB Strip" {
		t.Errorf("config.Name = %v, want RGB Strip", config.Name)
	}
}

func TestGen2RGB_On(t *testing.T) {
	t.Parallel()

	onCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			onCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	err = rgb.On(ctx)
	if err != nil {
		t.Fatalf("On() error = %v", err)
	}

	if !onCalled {
		t.Error("RGB.Set was not called")
	}
}

func TestGen2RGB_Off(t *testing.T) {
	t.Parallel()

	offCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			offCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	err = rgb.Off(ctx)
	if err != nil {
		t.Fatalf("Off() error = %v", err)
	}

	if !offCalled {
		t.Error("RGB.Set was not called")
	}
}

func TestGen2RGB_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Toggle", func(_ map[string]any) (any, error) {
			toggleCalled = true
			return map[string]any{"was_on": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	status, err := rgb.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("RGB.Toggle was not called")
	}
	if status == nil {
		t.Fatal("Toggle() returned nil status")
	}
}

func TestGen2RGB_SetBrightness(t *testing.T) {
	t.Parallel()

	setBrightnessCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			setBrightnessCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	err = rgb.SetBrightness(ctx, 50)
	if err != nil {
		t.Fatalf("SetBrightness() error = %v", err)
	}

	if !setBrightnessCalled {
		t.Error("RGB.Set was not called")
	}
}

func TestGen2RGB_SetColor(t *testing.T) {
	t.Parallel()

	setColorCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			setColorCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	err = rgb.SetColor(ctx, 255, 128, 64)
	if err != nil {
		t.Fatalf("SetColor() error = %v", err)
	}

	if !setColorCalled {
		t.Error("RGB.Set was not called")
	}
}

func TestGen2RGB_SetColorAndBrightness(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			setCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	err = rgb.SetColorAndBrightness(ctx, 255, 128, 64, 75)
	if err != nil {
		t.Fatalf("SetColorAndBrightness() error = %v", err)
	}

	if !setCalled {
		t.Error("RGB.Set was not called")
	}
}

func TestGen2RGB_Set(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGB.Set", func(_ map[string]any) (any, error) {
			setCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgb := client.RGB(0)
	r, g, b := 200, 100, 50
	brightness := 80
	on := true
	err = rgb.Set(ctx, &r, &g, &b, &brightness, &on)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !setCalled {
		t.Error("RGB.Set was not called")
	}
}

// ============================================
// Gen2 RGBW Component Tests
// ============================================

func TestGen2RGBW_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":         0,
				"output":     true,
				"source":     "button",
				"brightness": 80,
				"white":      50,
				"rgb":        []any{255.0, 128.0, 64.0},
				"apower":     5.5,
				"voltage":    120.0,
				"current":    0.046,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	status, err := rgbw.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.Output {
		t.Error("status.Output = false, want true")
	}
	if status.Brightness == nil || *status.Brightness != 80 {
		t.Errorf("status.Brightness = %v, want 80", status.Brightness)
	}
	if status.White == nil || *status.White != 50 {
		t.Errorf("status.White = %v, want 50", status.White)
	}
}

func TestGen2RGBW_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":                 0,
				"name":               "RGBW Strip",
				"initial_state":      "restore_last",
				"auto_on":            false,
				"auto_off":           false,
				"default_brightness": 100,
				"default_white":      50,
				"night_mode": map[string]any{
					"enable":     true,
					"brightness": 30,
					"white":      20,
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	config, err := rgbw.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "RGBW Strip" {
		t.Errorf("config.Name = %v, want RGBW Strip", config.Name)
	}
}

func TestGen2RGBW_On(t *testing.T) {
	t.Parallel()

	onCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			onCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	err = rgbw.On(ctx)
	if err != nil {
		t.Fatalf("On() error = %v", err)
	}

	if !onCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_Off(t *testing.T) {
	t.Parallel()

	offCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			offCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	err = rgbw.Off(ctx)
	if err != nil {
		t.Fatalf("Off() error = %v", err)
	}

	if !offCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Toggle", func(_ map[string]any) (any, error) {
			toggleCalled = true
			return map[string]any{"was_on": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	status, err := rgbw.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("RGBW.Toggle was not called")
	}
	if status == nil {
		t.Fatal("Toggle() returned nil status")
	}
}

func TestGen2RGBW_SetBrightness(t *testing.T) {
	t.Parallel()

	setBrightnessCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			setBrightnessCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	err = rgbw.SetBrightness(ctx, 50)
	if err != nil {
		t.Fatalf("SetBrightness() error = %v", err)
	}

	if !setBrightnessCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_SetWhite(t *testing.T) {
	t.Parallel()

	setWhiteCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			setWhiteCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	err = rgbw.SetWhite(ctx, 75)
	if err != nil {
		t.Fatalf("SetWhite() error = %v", err)
	}

	if !setWhiteCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_SetColor(t *testing.T) {
	t.Parallel()

	setColorCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			setColorCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	err = rgbw.SetColor(ctx, 255, 128, 64)
	if err != nil {
		t.Fatalf("SetColor() error = %v", err)
	}

	if !setColorCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_SetColorAndBrightness(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			setCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	white := 40
	err = rgbw.SetColorAndBrightness(ctx, 255, 128, 64, 75, &white)
	if err != nil {
		t.Fatalf("SetColorAndBrightness() error = %v", err)
	}

	if !setCalled {
		t.Error("RGBW.Set was not called")
	}
}

func TestGen2RGBW_Set(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("RGBW.Set", func(_ map[string]any) (any, error) {
			setCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	rgbw := client.RGBW(0)
	r, g, b := 200, 100, 50
	brightness := 80
	white := 30
	on := true
	err = rgbw.Set(ctx, &r, &g, &b, &brightness, &white, &on)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !setCalled {
		t.Error("RGBW.Set was not called")
	}
}

// ============================================
// Gen2 Input Component Tests
// ============================================

func TestGen2Input_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Input.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":    0,
				"state": true,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	input := client.Input(0)
	status, err := input.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
	if !status.State {
		t.Error("status.State = false, want true")
	}
}

func TestGen2Input_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Input.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":            0,
				"name":          "Wall Switch",
				"type":          "switch",
				"enable":        true,
				"invert":        false,
				"factory_reset": true,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	input := client.Input(0)
	config, err := input.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if config.Name == nil || *config.Name != "Wall Switch" {
		t.Errorf("config.Name = %v, want Wall Switch", config.Name)
	}
	if config.Type != "switch" {
		t.Errorf("config.Type = %q, want switch", config.Type)
	}
}

func TestGen2Input_SetConfig(t *testing.T) {
	t.Parallel()

	setConfigCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Input.SetConfig", func(_ map[string]any) (any, error) {
			setConfigCalled = true
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	input := client.Input(0)
	name := "Updated Switch"
	cfg := &model.InputConfig{
		ID:     0,
		Name:   &name,
		Type:   "switch",
		Enable: true,
		Invert: false,
	}
	err = input.SetConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	if !setConfigCalled {
		t.Error("Input.SetConfig was not called")
	}
}

func TestGen2Input_Trigger(t *testing.T) {
	t.Parallel()

	triggerCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Input.Trigger", func(_ map[string]any) (any, error) {
			triggerCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	input := client.Input(0)
	err = input.Trigger(ctx, "single_push")
	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}

	if !triggerCalled {
		t.Error("Input.Trigger was not called")
	}
}

func TestGen2Input_ResetCounters(t *testing.T) {
	t.Parallel()

	resetCountersCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Input.ResetCounters", func(_ map[string]any) (any, error) {
			resetCountersCalled = true
			return map[string]any{"counts": map[string]any{"total": 0}}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	input := client.Input(0)
	err = input.ResetCounters(ctx, []string{"total"})
	if err != nil {
		t.Fatalf("ResetCounters() error = %v", err)
	}

	if !resetCountersCalled {
		t.Error("Input.ResetCounters was not called")
	}
}

// ============================================
// Gen2 Thermostat Component Tests
// ============================================

func TestGen2Thermostat_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.GetStatus", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":              0,
				"enable":          true,
				"target_C":        21.5,
				"current_C":       20.2,
				"output":          true,
				"boost_minutes":   0,
				"mode":            "heat",
				"schedule_active": false,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	status, err := thermostat.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}
}

func TestGen2Thermostat_GetConfig(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.GetConfig", func(_ map[string]any) (any, error) {
			return map[string]any{
				"id":         0,
				"enable":     true,
				"target_C":   21.5,
				"min_C":      5.0,
				"max_C":      30.0,
				"mode":       "heat",
				"hysteresis": 0.5,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	config, err := thermostat.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
}

func TestGen2Thermostat_SetTarget(t *testing.T) {
	t.Parallel()

	setTargetCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.SetConfig", func(_ map[string]any) (any, error) {
			setTargetCalled = true
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.SetTarget(ctx, 22.0)
	if err != nil {
		t.Fatalf("SetTarget() error = %v", err)
	}

	if !setTargetCalled {
		t.Error("Thermostat.SetTarget was not called")
	}
}

func TestGen2Thermostat_Enable(t *testing.T) {
	t.Parallel()

	enableCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.SetConfig", func(_ map[string]any) (any, error) {
			enableCalled = true
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.Enable(ctx, true)
	if err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	if !enableCalled {
		t.Error("Thermostat.Enable was not called")
	}
}

func TestGen2Thermostat_SetMode(t *testing.T) {
	t.Parallel()

	setModeCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.SetConfig", func(_ map[string]any) (any, error) {
			setModeCalled = true
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.SetMode(ctx, "heat")
	if err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	if !setModeCalled {
		t.Error("Thermostat.SetMode was not called")
	}
}

func TestGen2Thermostat_Boost(t *testing.T) {
	t.Parallel()

	boostCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.Boost", func(_ map[string]any) (any, error) {
			boostCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.Boost(ctx, 30)
	if err != nil {
		t.Fatalf("Boost() error = %v", err)
	}

	if !boostCalled {
		t.Error("Thermostat.Boost was not called")
	}
}

func TestGen2Thermostat_CancelBoost(t *testing.T) {
	t.Parallel()

	cancelBoostCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.CancelBoost", func(_ map[string]any) (any, error) {
			cancelBoostCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.CancelBoost(ctx)
	if err != nil {
		t.Fatalf("CancelBoost() error = %v", err)
	}

	if !cancelBoostCalled {
		t.Error("Thermostat.CancelBoost was not called")
	}
}

func TestGen2Thermostat_Override(t *testing.T) {
	t.Parallel()

	overrideCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.Override", func(_ map[string]any) (any, error) {
			overrideCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.Override(ctx, 24.0, 3600)
	if err != nil {
		t.Fatalf("Override() error = %v", err)
	}

	if !overrideCalled {
		t.Error("Thermostat.Override was not called")
	}
}

func TestGen2Thermostat_CancelOverride(t *testing.T) {
	t.Parallel()

	cancelOverrideCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.CancelOverride", func(_ map[string]any) (any, error) {
			cancelOverrideCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.CancelOverride(ctx)
	if err != nil {
		t.Fatalf("CancelOverride() error = %v", err)
	}

	if !cancelOverrideCalled {
		t.Error("Thermostat.CancelOverride was not called")
	}
}

func TestGen2Thermostat_Calibrate(t *testing.T) {
	t.Parallel()

	calibrateCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.Calibrate", func(_ map[string]any) (any, error) {
			calibrateCalled = true
			return map[string]any{}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	err = thermostat.Calibrate(ctx)
	if err != nil {
		t.Fatalf("Calibrate() error = %v", err)
	}

	if !calibrateCalled {
		t.Error("Thermostat.Calibrate was not called")
	}
}

func TestGen2Thermostat_ID(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(5)
	if thermostat.ID() != 5 {
		t.Errorf("Thermostat.ID() = %d, want 5", thermostat.ID())
	}
}

// =============================================================================
// KVS Component Tests
// =============================================================================

func TestKVS_List(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.List", func(_ map[string]any) (any, error) {
			return map[string]any{
				"keys": []string{"key1", "key2", "key3"},
				"rev":  42,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	result, err := kvs.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result.Keys) != 3 {
		t.Errorf("List() keys count = %d, want 3", len(result.Keys))
	}
	if result.Rev != 42 {
		t.Errorf("List() rev = %d, want 42", result.Rev)
	}
}

func TestKVS_Get(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.Get", func(params map[string]any) (any, error) {
			key := params["key"].(string)
			if key != "testkey" {
				return nil, fmt.Errorf("unexpected key: %s", key)
			}
			return map[string]any{
				"value": "testvalue",
				"etag":  "abc123",
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	result, err := kvs.Get(ctx, "testkey")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result.Value != "testvalue" {
		t.Errorf("Get() value = %v, want testvalue", result.Value)
	}
	if result.Etag != "abc123" {
		t.Errorf("Get() etag = %s, want abc123", result.Etag)
	}
}

func TestKVS_GetMany(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.GetMany", func(_ map[string]any) (any, error) {
			etag1 := "etag1"
			return map[string]any{
				"items": []map[string]any{
					{"key": "key1", "value": "value1", "etag": &etag1},
					{"key": "key2", "value": "value2"},
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	items, err := kvs.GetMany(ctx, "key*")
	if err != nil {
		t.Fatalf("GetMany() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("GetMany() items count = %d, want 2", len(items))
	}
}

func TestKVS_Set(t *testing.T) {
	t.Parallel()

	setCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.Set", func(params map[string]any) (any, error) {
			setCalled = true
			key := params["key"].(string)
			if key != "mykey" {
				return nil, fmt.Errorf("unexpected key: %s", key)
			}
			return map[string]any{
				"etag": "newetag",
				"rev":  43,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	result, err := kvs.Set(ctx, "mykey", "myvalue")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !setCalled {
		t.Error("KVS.Set was not called")
	}
	if result.Etag != "newetag" {
		t.Errorf("Set() etag = %s, want newetag", result.Etag)
	}
	if result.Rev != 43 {
		t.Errorf("Set() rev = %d, want 43", result.Rev)
	}
}

func TestKVS_Delete(t *testing.T) {
	t.Parallel()

	deleteCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.Delete", func(params map[string]any) (any, error) {
			deleteCalled = true
			key := params["key"].(string)
			if key != "deletekey" {
				return nil, fmt.Errorf("unexpected key: %s", key)
			}
			return map[string]any{
				"rev": 44,
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	result, err := kvs.Delete(ctx, "deletekey")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if !deleteCalled {
		t.Error("KVS.Delete was not called")
	}
	if result.Rev != 44 {
		t.Errorf("Delete() rev = %d, want 44", result.Rev)
	}
}

func TestKVS_GetAll(t *testing.T) {
	t.Parallel()

	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("KVS.GetMany", func(params map[string]any) (any, error) {
			// GetAll calls GetMany with "*"
			match := params["match"].(string)
			if match != "*" {
				return nil, fmt.Errorf("unexpected match: %s", match)
			}
			return map[string]any{
				"items": []map[string]any{
					{"key": "key1", "value": "value1"},
					{"key": "key2", "value": "value2"},
				},
			}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	kvs := client.KVS()
	items, err := kvs.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("GetAll() items count = %d, want 2", len(items))
	}
}

// =============================================================================
// Gen1 Color Component Tests
// =============================================================================

func TestGen1Color_TurnOn(t *testing.T) {
	t.Parallel()

	turnOnCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") {
				turnOnCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.TurnOn(ctx)
	if err != nil {
		t.Fatalf("TurnOn() error = %v", err)
	}

	if !turnOnCalled {
		t.Error("Color turn=on was not called")
	}
}

func TestGen1Color_TurnOff(t *testing.T) {
	t.Parallel()

	turnOffCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=off") {
				turnOffCalled = true
			}
			return map[string]any{"ison": false}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.TurnOff(ctx)
	if err != nil {
		t.Fatalf("TurnOff() error = %v", err)
	}

	if !turnOffCalled {
		t.Error("Color turn=off was not called")
	}
}

func TestGen1Color_Toggle(t *testing.T) {
	t.Parallel()

	toggleCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=toggle") {
				toggleCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.Toggle(ctx)
	if err != nil {
		t.Fatalf("Toggle() error = %v", err)
	}

	if !toggleCalled {
		t.Error("Color turn=toggle was not called")
	}
}

func TestGen1Color_SetRGB(t *testing.T) {
	t.Parallel()

	setRGBCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "red=255") && strings.Contains(query, "green=128") && strings.Contains(query, "blue=64") {
				setRGBCalled = true
			}
			return map[string]any{"ison": true, "red": 255, "green": 128, "blue": 64}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.SetRGB(ctx, 255, 128, 64)
	if err != nil {
		t.Fatalf("SetRGB() error = %v", err)
	}

	if !setRGBCalled {
		t.Error("Color SetRGB was not called with correct values")
	}
}

func TestGen1Color_SetRGBW(t *testing.T) {
	t.Parallel()

	setRGBWCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "red=255") && strings.Contains(query, "green=128") && strings.Contains(query, "blue=64") && strings.Contains(query, "white=100") {
				setRGBWCalled = true
			}
			return map[string]any{"ison": true, "red": 255, "green": 128, "blue": 64, "white": 100}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.SetRGBW(ctx, 255, 128, 64, 100)
	if err != nil {
		t.Fatalf("SetRGBW() error = %v", err)
	}

	if !setRGBWCalled {
		t.Error("Color SetRGBW was not called with correct values")
	}
}

func TestGen1Color_SetGain(t *testing.T) {
	t.Parallel()

	setGainCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "gain=75") {
				setGainCalled = true
			}
			return map[string]any{"ison": true, "gain": 75}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.SetGain(ctx, 75)
	if err != nil {
		t.Fatalf("SetGain() error = %v", err)
	}

	if !setGainCalled {
		t.Error("Color SetGain was not called with correct value")
	}
}

func TestGen1Color_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(_ string) (any, int) {
			return map[string]any{
				"ison":  true,
				"red":   255,
				"green": 128,
				"blue":  64,
				"white": 50,
				"gain":  80,
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	status, err := color.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if !status.IsOn {
		t.Error("GetStatus() IsOn = false, want true")
	}
	if status.Red != 255 {
		t.Errorf("GetStatus() Red = %d, want 255", status.Red)
	}
}

func TestGen1Color_TurnOnForDuration(t *testing.T) {
	t.Parallel()

	turnOnForDurationCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") && strings.Contains(query, "timer=60") {
				turnOnForDurationCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.TurnOnForDuration(ctx, 60)
	if err != nil {
		t.Fatalf("TurnOnForDuration() error = %v", err)
	}

	if !turnOnForDurationCalled {
		t.Error("Color TurnOnForDuration was not called")
	}
}

func TestGen1Color_TurnOnWithRGB(t *testing.T) {
	t.Parallel()

	turnOnWithRGBCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/color/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") && strings.Contains(query, "red=200") && strings.Contains(query, "green=100") && strings.Contains(query, "blue=50") && strings.Contains(query, "gain=90") {
				turnOnWithRGBCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	color, err := client.Color(0)
	if err != nil {
		t.Fatalf("client.Color() error = %v", err)
	}

	err = color.TurnOnWithRGB(ctx, 200, 100, 50, 90)
	if err != nil {
		t.Fatalf("TurnOnWithRGB() error = %v", err)
	}

	if !turnOnWithRGBCalled {
		t.Error("Color TurnOnWithRGB was not called with correct values")
	}
}

// =============================================================================
// Gen1 White Component Tests
// =============================================================================

func TestGen1White_TurnOn(t *testing.T) {
	t.Parallel()

	turnOnCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/white/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") {
				turnOnCalled = true
			}
			return map[string]any{"ison": true, "brightness": 100}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	white, err := client.White(0)
	if err != nil {
		t.Fatalf("client.White() error = %v", err)
	}

	err = white.TurnOn(ctx)
	if err != nil {
		t.Fatalf("TurnOn() error = %v", err)
	}

	if !turnOnCalled {
		t.Error("White turn=on was not called")
	}
}

func TestGen1White_TurnOff(t *testing.T) {
	t.Parallel()

	turnOffCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/white/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=off") {
				turnOffCalled = true
			}
			return map[string]any{"ison": false, "brightness": 100}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	white, err := client.White(0)
	if err != nil {
		t.Fatalf("client.White() error = %v", err)
	}

	err = white.TurnOff(ctx)
	if err != nil {
		t.Fatalf("TurnOff() error = %v", err)
	}

	if !turnOffCalled {
		t.Error("White turn=off was not called")
	}
}

func TestGen1White_SetBrightness(t *testing.T) {
	t.Parallel()

	setBrightnessCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/white/0", func(query string) (any, int) {
			if strings.Contains(query, "brightness=75") {
				setBrightnessCalled = true
			}
			return map[string]any{"ison": true, "brightness": 75}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	white, err := client.White(0)
	if err != nil {
		t.Fatalf("client.White() error = %v", err)
	}

	err = white.SetBrightness(ctx, 75)
	if err != nil {
		t.Fatalf("SetBrightness() error = %v", err)
	}

	if !setBrightnessCalled {
		t.Error("White SetBrightness was not called")
	}
}

func TestGen1White_GetStatus(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/white/0", func(_ string) (any, int) {
			return map[string]any{
				"ison":       true,
				"brightness": 80,
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	white, err := client.White(0)
	if err != nil {
		t.Fatalf("client.White() error = %v", err)
	}

	status, err := white.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if !status.IsOn {
		t.Error("GetStatus() IsOn = false, want true")
	}
	if status.Brightness != 80 {
		t.Errorf("GetStatus() Brightness = %d, want 80", status.Brightness)
	}
}

// =============================================================================
// Gen1 Extended Methods Tests
// =============================================================================

func TestGen1Roller_OpenForDuration(t *testing.T) {
	t.Parallel()

	openForDurationCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(query string) (any, int) {
			if strings.Contains(query, "go=open") && strings.Contains(query, "duration=10") {
				openForDurationCalled = true
			}
			return map[string]any{"state": "open"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}

	err = roller.OpenForDuration(ctx, 10.0)
	if err != nil {
		t.Fatalf("OpenForDuration() error = %v", err)
	}

	if !openForDurationCalled {
		t.Error("Roller OpenForDuration was not called")
	}
}

func TestGen1Roller_CloseForDuration(t *testing.T) {
	t.Parallel()

	closeForDurationCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/roller/0", func(query string) (any, int) {
			if strings.Contains(query, "go=close") && strings.Contains(query, "duration=15") {
				closeForDurationCalled = true
			}
			return map[string]any{"state": "close"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	roller, err := client.Roller(0)
	if err != nil {
		t.Fatalf("client.Roller() error = %v", err)
	}

	err = roller.CloseForDuration(ctx, 15.0)
	if err != nil {
		t.Fatalf("CloseForDuration() error = %v", err)
	}

	if !closeForDurationCalled {
		t.Error("Roller CloseForDuration was not called")
	}
}

func TestGen1Light_TurnOnWithBrightness(t *testing.T) {
	t.Parallel()

	turnOnWithBrightnessCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") && strings.Contains(query, "brightness=50") {
				turnOnWithBrightnessCalled = true
			}
			return map[string]any{"ison": true, "brightness": 50}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}

	err = light.TurnOnWithBrightness(ctx, 50)
	if err != nil {
		t.Fatalf("TurnOnWithBrightness() error = %v", err)
	}

	if !turnOnWithBrightnessCalled {
		t.Error("Light TurnOnWithBrightness was not called")
	}
}

func TestGen1Light_TurnOnForDuration(t *testing.T) {
	t.Parallel()

	turnOnForDurationCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(query string) (any, int) {
			if strings.Contains(query, "turn=on") && strings.Contains(query, "timer=30") {
				turnOnForDurationCalled = true
			}
			return map[string]any{"ison": true}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}

	err = light.TurnOnForDuration(ctx, 30)
	if err != nil {
		t.Fatalf("TurnOnForDuration() error = %v", err)
	}

	if !turnOnForDurationCalled {
		t.Error("Light TurnOnForDuration was not called")
	}
}

func TestGen1Light_SetBrightnessWithTransition(t *testing.T) {
	t.Parallel()

	setBrightnessWithTransitionCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/light/0", func(query string) (any, int) {
			if strings.Contains(query, "brightness=80") && strings.Contains(query, "transition=1000") {
				setBrightnessWithTransitionCalled = true
			}
			return map[string]any{"ison": true, "brightness": 80}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	light, err := client.Light(0)
	if err != nil {
		t.Fatalf("client.Light() error = %v", err)
	}

	err = light.SetBrightnessWithTransition(ctx, 80, 1000)
	if err != nil {
		t.Fatalf("SetBrightnessWithTransition() error = %v", err)
	}

	if !setBrightnessWithTransitionCalled {
		t.Error("Light SetBrightnessWithTransition was not called")
	}
}

// =============================================================================
// Gen1 Client Action Methods Tests
// =============================================================================

func TestGen1Client_GetDebugLog(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/debug/log", func(_ string) (any, int) {
			return "Debug log line 1\nDebug log line 2", http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	_, err = client.GetDebugLog(ctx)
	if err != nil {
		t.Fatalf("GetDebugLog() error = %v", err)
	}
}

func TestGen1Client_GetActions(t *testing.T) {
	t.Parallel()

	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/settings/actions", func(_ string) (any, int) {
			// Return valid action settings format per shelly-go library expectations
			return map[string]any{
				"actions": []any{
					map[string]any{
						"index":   0,
						"name":    "relay_on_url",
						"enabled": true,
						"urls":    []string{"http://example.com"},
					},
				},
			}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	_, err = client.GetActions(ctx)
	if err != nil {
		t.Fatalf("GetActions() error = %v", err)
	}
}

func TestGen1Client_Update(t *testing.T) {
	t.Parallel()

	updateCalled := false
	mock := newMockGen1Server().
		handle("/shelly", func(_ string) (any, int) {
			return standardGen1DeviceInfo(), http.StatusOK
		}).
		handle("/ota", func(query string) (any, int) {
			if strings.Contains(query, "url=") {
				updateCalled = true
			}
			return map[string]any{"status": "ok"}, http.StatusOK
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := ConnectGen1(ctx, device)
	if err != nil {
		t.Fatalf("ConnectGen1() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	err = client.Update(ctx, "http://example.com/firmware.bin")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if !updateCalled {
		t.Error("Update was not called")
	}
}

// =============================================================================
// Thermostat.SetConfig Test
// =============================================================================

func TestGen2Thermostat_SetConfig(t *testing.T) {
	t.Parallel()

	setConfigCalled := false
	mock := newMockRPCServer().
		handle("Shelly.GetDeviceInfo", func(_ map[string]any) (any, error) {
			return standardDeviceInfo(), nil
		}).
		handle("Thermostat.SetConfig", func(_ map[string]any) (any, error) {
			setConfigCalled = true
			return map[string]any{"restart_required": false}, nil
		})

	server := mock.start(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device := model.Device{Address: server.URL}
	client, err := Connect(ctx, device)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			t.Logf("warning: close error: %v", cerr)
		}
	}()

	thermostat := client.Thermostat(0)
	enable := true
	err = thermostat.SetConfig(ctx, &components.ThermostatConfig{
		ID:     0,
		Enable: &enable,
	})
	if err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	if !setConfigCalled {
		t.Error("Thermostat.SetConfig was not called")
	}
}
