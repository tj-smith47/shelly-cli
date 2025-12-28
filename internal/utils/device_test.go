// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"os"
	"path/filepath"
	"testing"

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

func TestParseDevicesJSON_FromFile(t *testing.T) {
	t.Parallel()

	// Create a temp file with device JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "devices.json")

	content := `[{"name":"filedev","address":"10.0.0.1"}]`
	if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
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

func TestParseDevicesJSON_FileReadError(t *testing.T) {
	t.Parallel()

	// Test with a directory path (cannot be read as file)
	tmpDir := t.TempDir()

	_, err := ParseDevicesJSON(tmpDir)
	if err == nil {
		t.Error("ParseDevicesJSON() should error when reading a directory")
	}
}

func TestParseDevicesJSON_EmptyFile(t *testing.T) {
	t.Parallel()

	// Create an empty temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.json")

	if err := os.WriteFile(tmpFile, []byte(""), 0o600); err != nil {
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

func TestParseDevicesJSON_WhitespaceOnlyFile(t *testing.T) {
	t.Parallel()

	// Create a file with only whitespace
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "whitespace.json")

	if err := os.WriteFile(tmpFile, []byte("   \n\t  \n"), 0o600); err != nil {
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
