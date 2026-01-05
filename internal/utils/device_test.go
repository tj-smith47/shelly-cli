// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

const testUsername = "admin"

func TestParseDeviceSpec_ValidFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     string
		wantName string
		wantAddr string
		wantUser string
		wantPass string
	}{
		{
			name:     "simple name=ip",
			spec:     "kitchen=192.168.1.100",
			wantName: "kitchen",
			wantAddr: "192.168.1.100",
		},
		{
			name:     "with whitespace",
			spec:     "  living  =  10.0.0.50  ",
			wantName: "living",
			wantAddr: "10.0.0.50",
		},
		{
			name:     "with auth",
			spec:     "secure=192.168.1.102:admin:secret",
			wantName: "secure",
			wantAddr: "192.168.1.102",
			wantUser: "admin",
			wantPass: "secret",
		},
		{
			name:     "with port (ip:port format)",
			spec:     "custom=192.168.1.103:8080",
			wantName: "custom",
			wantAddr: "192.168.1.103:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			device, err := ParseDeviceSpec(tt.spec)
			if err != nil {
				t.Fatalf("ParseDeviceSpec(%q) error = %v", tt.spec, err)
			}

			if device.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", device.Name, tt.wantName)
			}
			if device.Address != tt.wantAddr {
				t.Errorf("Address = %q, want %q", device.Address, tt.wantAddr)
			}
			if device.Username != tt.wantUser {
				t.Errorf("Username = %q, want %q", device.Username, tt.wantUser)
			}
			if device.Password != tt.wantPass {
				t.Errorf("Password = %q, want %q", device.Password, tt.wantPass)
			}
		})
	}
}

func TestParseDeviceSpec_InvalidFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		spec string
	}{
		{"no equals sign", "kitchen192.168.1.100"},
		{"empty name", "=192.168.1.100"},
		{"empty address", "kitchen="},
		{"empty string", ""},
		{"only equals", "="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseDeviceSpec(tt.spec)
			if err == nil {
				t.Errorf("ParseDeviceSpec(%q) should return error", tt.spec)
			}
		})
	}
}

func TestParseDevicesJSON_Array(t *testing.T) {
	t.Parallel()

	input := `[{"name":"dev1","address":"192.168.1.1"},{"name":"dev2","address":"192.168.1.2"}]`

	devices, err := ParseDevicesJSON(input)
	if err != nil {
		t.Fatalf("ParseDevicesJSON() error = %v", err)
	}

	if len(devices) != 2 {
		t.Fatalf("len(devices) = %d, want 2", len(devices))
	}
	if devices[0].Name != "dev1" {
		t.Errorf("devices[0].Name = %q, want 'dev1'", devices[0].Name)
	}
	if devices[0].Address != "192.168.1.1" {
		t.Errorf("devices[0].Address = %q, want '192.168.1.1'", devices[0].Address)
	}
	if devices[1].Name != "dev2" {
		t.Errorf("devices[1].Name = %q, want 'dev2'", devices[1].Name)
	}
}

func TestParseDevicesJSON_SingleObject(t *testing.T) {
	t.Parallel()

	input := `{"name":"kitchen","address":"192.168.1.100","username":"admin","password":"secret"}`

	devices, err := ParseDevicesJSON(input)
	if err != nil {
		t.Fatalf("ParseDevicesJSON() error = %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if devices[0].Name != "kitchen" {
		t.Errorf("Name = %q, want 'kitchen'", devices[0].Name)
	}
	if devices[0].Username != testUsername {
		t.Errorf("Username = %q, want %q", devices[0].Username, testUsername)
	}
	if devices[0].Password != "secret" {
		t.Errorf("Password = %q, want 'secret'", devices[0].Password)
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestParseDevicesJSON_FromFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create a file with device JSON
	tmpFile := "/test/devices.json"
	content := `[{"name":"filedev","address":"10.0.0.1"}]`
	if err := afero.WriteFile(fs, tmpFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	devices, err := ParseDevicesJSON(tmpFile)
	if err != nil {
		t.Fatalf("ParseDevicesJSON() error = %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("len(devices) = %d, want 1", len(devices))
	}
	if devices[0].Name != "filedev" {
		t.Errorf("Name = %q, want 'filedev'", devices[0].Name)
	}
}

func TestParseDevicesJSON_EmptyInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			devices, err := ParseDevicesJSON(tt.input)
			if err != nil {
				t.Fatalf("ParseDevicesJSON(%q) error = %v", tt.input, err)
			}
			if len(devices) != 0 {
				t.Errorf("len(devices) = %d, want 0", len(devices))
			}
		})
	}
}

func TestParseDevicesJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"invalid array", "[{invalid}]"},
		{"invalid object", "{invalid}"},
		{"not JSON", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseDevicesJSON(tt.input)
			if err == nil {
				t.Errorf("ParseDevicesJSON(%q) should return error", tt.input)
			}
		})
	}
}

func TestPluginDetectionResultToConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		result    *plugins.DeviceDetectionResult
		address   string
		wantName  string
		wantModel string
		wantPlatf string
	}{
		{
			name: "with device name",
			result: &plugins.DeviceDetectionResult{
				Detected:   true,
				Platform:   "shelly",
				DeviceID:   "shelly123",
				DeviceName: "Kitchen Light",
				Model:      "SHSW-1",
			},
			address:   "192.168.1.50",
			wantName:  "Kitchen Light",
			wantPlatf: "shelly",
		},
		{
			name: "without device name - uses ID",
			result: &plugins.DeviceDetectionResult{
				Detected:   true,
				Platform:   "tasmota",
				DeviceID:   "tasmota123",
				DeviceName: "",
				Model:      "Generic",
			},
			address:   "192.168.1.51",
			wantName:  "tasmota123",
			wantPlatf: "tasmota",
		},
		{
			name: "without name or ID - uses address",
			result: &plugins.DeviceDetectionResult{
				Detected:   true,
				Platform:   "esphome",
				DeviceID:   "",
				DeviceName: "",
				Model:      "ESP32",
			},
			address:   "192.168.1.52",
			wantName:  "192.168.1.52",
			wantPlatf: "esphome",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			device := PluginDetectionResultToConfig(tt.result, tt.address)

			if device.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", device.Name, tt.wantName)
			}
			if device.Address != tt.address {
				t.Errorf("Address = %q, want %q", device.Address, tt.address)
			}
			if device.Platform != tt.wantPlatf {
				t.Errorf("Platform = %q, want %q", device.Platform, tt.wantPlatf)
			}
			if device.Generation != 0 {
				t.Errorf("Generation = %d, want 0 (plugin devices don't have generations)", device.Generation)
			}
		})
	}
}

func TestDeviceSpec_Struct(t *testing.T) {
	t.Parallel()

	spec := DeviceSpec{
		Name:     "test",
		Address:  "192.168.1.1",
		Username: testUsername,
		Password: "secret",
	}

	if spec.Name != "test" {
		t.Errorf("Name = %q, want 'test'", spec.Name)
	}
	if spec.Address != "192.168.1.1" {
		t.Errorf("Address = %q, want '192.168.1.1'", spec.Address)
	}
	if spec.Username != testUsername {
		t.Errorf("Username = %q, want %q", spec.Username, testUsername)
	}
	if spec.Password != "secret" {
		t.Errorf("Password = %q, want 'secret'", spec.Password)
	}
}

func TestJSONDevice_Struct(t *testing.T) {
	t.Parallel()

	dev := JSONDevice{
		Name:     "kitchen",
		Address:  "192.168.1.100",
		Username: "user",
		Password: "pass",
	}

	if dev.Name != "kitchen" {
		t.Errorf("Name = %q, want 'kitchen'", dev.Name)
	}
	if dev.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want '192.168.1.100'", dev.Address)
	}
	if dev.Username != "user" {
		t.Errorf("Username = %q, want 'user'", dev.Username)
	}
	if dev.Password != "pass" {
		t.Errorf("Password = %q, want 'pass'", dev.Password)
	}
}

func TestPluginDevice_Struct(t *testing.T) {
	t.Parallel()

	dev := PluginDevice{
		Address:  "192.168.1.200",
		Platform: "esphome",
		ID:       "esp123",
		Name:     "Sensor",
		Model:    "ESP32",
		Firmware: "1.0.0",
	}

	if dev.Address != "192.168.1.200" {
		t.Errorf("Address = %q, want '192.168.1.200'", dev.Address)
	}
	if dev.Platform != "esphome" {
		t.Errorf("Platform = %q, want 'esphome'", dev.Platform)
	}
	if dev.ID != "esp123" {
		t.Errorf("ID = %q, want 'esp123'", dev.ID)
	}
	if dev.Name != "Sensor" {
		t.Errorf("Name = %q, want 'Sensor'", dev.Name)
	}
	if dev.Model != "ESP32" {
		t.Errorf("Model = %q, want 'ESP32'", dev.Model)
	}
	if dev.Firmware != "1.0.0" {
		t.Errorf("Firmware = %q, want '1.0.0'", dev.Firmware)
	}
}

// errorOnReadFs wraps an afero.Fs and returns an error for specific paths.
type errorOnReadFs struct {
	afero.Fs
	errorPath string
}

func (e *errorOnReadFs) Open(name string) (afero.File, error) {
	if name == e.errorPath {
		return nil, fmt.Errorf("simulated read error for %s", name)
	}
	return e.Fs.Open(name)
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestParseDevicesJSON_FileReadError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	// Create a file that "exists" for afero.Exists check
	if err := afero.WriteFile(memFs, "/test/errorfile.json", []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Wrap with error-injecting fs
	errorFs := &errorOnReadFs{Fs: memFs, errorPath: "/test/errorfile.json"}
	config.SetFs(errorFs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Should error when file read fails
	_, err := ParseDevicesJSON("/test/errorfile.json")
	if err == nil {
		t.Error("ParseDevicesJSON() should error when file read fails")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestParseDevicesJSON_EmptyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create an empty file
	tmpFile := "/test/empty.json"
	if err := afero.WriteFile(fs, tmpFile, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	devices, err := ParseDevicesJSON(tmpFile)
	if err != nil {
		t.Fatalf("ParseDevicesJSON() error = %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("len(devices) = %d, want 0 for empty file", len(devices))
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestParseDevicesJSON_WhitespaceOnlyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Create a file with only whitespace
	tmpFile := "/test/whitespace.json"
	if err := afero.WriteFile(fs, tmpFile, []byte("   \n\t  \n"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	devices, err := ParseDevicesJSON(tmpFile)
	if err != nil {
		t.Fatalf("ParseDevicesJSON() error = %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("len(devices) = %d, want 0 for whitespace file", len(devices))
	}
}

func TestParseDeviceSpec_ComplexAuth(t *testing.T) {
	t.Parallel()

	// Test password with special characters
	spec := "secure=192.168.1.100:admin:p@ss:w0rd"
	device, err := ParseDeviceSpec(spec)
	if err != nil {
		t.Fatalf("ParseDeviceSpec(%q) error = %v", spec, err)
	}

	// With 3 colons after the IP, we get ip:user:pass
	if device.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want '192.168.1.100'", device.Address)
	}
	if device.Username != testUsername {
		t.Errorf("Username = %q, want %q", device.Username, testUsername)
	}
	// Password captures everything after the second colon
	if device.Password != "p@ss:w0rd" {
		t.Errorf("Password = %q, want 'p@ss:w0rd'", device.Password)
	}
}

func TestListRegisteredDevices(t *testing.T) {
	t.Parallel()

	// Just verify it returns a map (actual contents depend on config state)
	devices := ListRegisteredDevices()
	if devices == nil {
		t.Error("ListRegisteredDevices() returned nil, want non-nil map")
	}
}

func TestIsPluginDeviceRegistered(t *testing.T) {
	t.Parallel()

	// Test with an address that's unlikely to be registered
	registered := IsPluginDeviceRegistered("192.168.255.255")
	// We can't know if it's registered without controlling config state,
	// but we can verify the function doesn't panic
	_ = registered
}

func TestUnmarshalJSON_NilTarget(t *testing.T) {
	t.Parallel()

	input := map[string]any{"key": "value"}

	// Unmarshal to nil target should not panic
	err := UnmarshalJSON(input, nil)
	if err == nil {
		t.Error("UnmarshalJSON(data, nil) should return error")
	}
}

func TestUnmarshalJSON_MarshalError(t *testing.T) {
	t.Parallel()

	// Create an unmarshalable value (channel)
	input := make(chan int)

	var result map[string]any
	err := UnmarshalJSON(input, &result)
	if err == nil {
		t.Error("UnmarshalJSON() should return error for unmarshalable input")
	}
}

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
