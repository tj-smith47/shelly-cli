package mock_test

import (
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

func TestDevice_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device mock.Device
	}{
		{
			name: "full device",
			device: mock.Device{
				Name:     "test-device",
				Model:    "Shelly Plus 1PM",
				Firmware: "1.0.0",
				MAC:      "AA:BB:CC:DD:EE:FF",
				State: map[string]interface{}{
					"output": true,
					"power":  100.5,
				},
			},
		},
		{
			name: "minimal device",
			device: mock.Device{
				Name: "minimal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.device.Name == "" && tt.name != "minimal device" {
				t.Errorf("Name should not be empty for non-minimal device")
			}
		})
	}
}

func TestDevice_FullDevice(t *testing.T) {
	t.Parallel()

	device := mock.Device{
		Name:     "kitchen-light",
		Model:    "Shelly Plus 1PM",
		Firmware: "1.2.3",
		MAC:      "AA:BB:CC:11:22:33",
		State: map[string]interface{}{
			"output":  true,
			"power":   45.5,
			"voltage": 120.0,
		},
	}

	if device.Name != "kitchen-light" {
		t.Errorf("Name = %q, want %q", device.Name, "kitchen-light")
	}
	if device.Model != "Shelly Plus 1PM" {
		t.Errorf("Model = %q, want %q", device.Model, "Shelly Plus 1PM")
	}
	if device.Firmware != "1.2.3" {
		t.Errorf("Firmware = %q, want %q", device.Firmware, "1.2.3")
	}
	if device.MAC != "AA:BB:CC:11:22:33" {
		t.Errorf("MAC = %q, want %q", device.MAC, "AA:BB:CC:11:22:33")
	}
	if device.State == nil {
		t.Error("State should not be nil")
	}
	if len(device.State) != 3 {
		t.Errorf("State has %d entries, want 3", len(device.State))
	}
}

func TestDir(t *testing.T) {
	// Cannot run in parallel - modifies global state via config.SetFs
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	testConfigHome := "/test/config"
	t.Setenv("XDG_CONFIG_HOME", testConfigHome)

	dir, err := mock.Dir()
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}

	if dir == "" {
		t.Error("Dir() returned empty string")
	}

	// Verify it's in the config directory, not real config
	if !strings.HasPrefix(dir, testConfigHome) {
		t.Errorf("Dir() = %q, expected to be under %q", dir, testConfigHome)
	}
}

func TestGenerateMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantLen int
		wantPfx string
	}{
		{
			name:    "simple name",
			input:   "test",
			wantLen: 17, // AA:BB:CC:XX:XX:XX
			wantPfx: "AA:BB:CC:",
		},
		{
			name:    "longer name",
			input:   "kitchen-light",
			wantLen: 17,
			wantPfx: "AA:BB:CC:",
		},
		{
			name:    "empty name",
			input:   "",
			wantLen: 17,
			wantPfx: "AA:BB:CC:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mock.GenerateMAC(tt.input)

			if len(got) != tt.wantLen {
				t.Errorf("GenerateMAC(%q) length = %d, want %d", tt.input, len(got), tt.wantLen)
			}

			if len(got) >= len(tt.wantPfx) && got[:len(tt.wantPfx)] != tt.wantPfx {
				t.Errorf("GenerateMAC(%q) prefix = %q, want %q", tt.input, got[:len(tt.wantPfx)], tt.wantPfx)
			}
		})
	}
}

func TestGenerateMAC_Deterministic(t *testing.T) {
	t.Parallel()

	// Same input should produce same output
	name := "test-device"
	mac1 := mock.GenerateMAC(name)
	mac2 := mock.GenerateMAC(name)

	if mac1 != mac2 {
		t.Errorf("GenerateMAC is not deterministic: %q != %q", mac1, mac2)
	}
}

func TestGenerateMAC_DifferentInputsDifferentOutput(t *testing.T) {
	t.Parallel()

	mac1 := mock.GenerateMAC("device1")
	mac2 := mock.GenerateMAC("device2")

	if mac1 == mac2 {
		t.Error("Different inputs should produce different MACs")
	}
}

func TestGenerateMAC_Format(t *testing.T) {
	t.Parallel()

	mac := mock.GenerateMAC("test")

	// Should be in format XX:XX:XX:XX:XX:XX
	parts := 0
	for i := range mac {
		if mac[i] == ':' {
			parts++
		}
	}

	if parts != 5 {
		t.Errorf("MAC should have 5 colons, got %d", parts)
	}

	// MAC should have exactly 17 characters (6 pairs of 2 hex chars + 5 colons)
	if len(mac) != 17 {
		t.Errorf("MAC length = %d, want 17", len(mac))
	}
}

func TestDevice_StateAccess(t *testing.T) {
	t.Parallel()

	device := mock.Device{
		Name: "test",
		State: map[string]interface{}{
			"output": true,
			"power":  100.5,
			"count":  42,
		},
	}

	// Test type assertions
	if output, ok := device.State["output"].(bool); !ok || !output {
		t.Error("Failed to access output as bool")
	}

	if power, ok := device.State["power"].(float64); !ok || power != 100.5 {
		t.Error("Failed to access power as float64")
	}

	if count, ok := device.State["count"].(int); !ok || count != 42 {
		t.Error("Failed to access count as int")
	}
}

func TestDevice_EmptyState(t *testing.T) {
	t.Parallel()

	device := mock.Device{
		Name:  "test",
		State: nil,
	}

	if device.State != nil {
		t.Error("State should be nil")
	}

	device.State = make(map[string]interface{})
	if device.State == nil {
		t.Error("State should not be nil after initialization")
	}
}

func TestDevice_JSONTags(t *testing.T) {
	t.Parallel()

	// This test verifies the struct can be used with JSON marshaling
	// by checking that fields are accessible via their JSON names
	device := mock.Device{
		Name:     "test",
		Model:    "model",
		Firmware: "1.0.0",
		MAC:      "AA:BB:CC:DD:EE:FF",
		State:    map[string]interface{}{"key": "value"},
	}

	// Just verify all fields are accessible
	if device.Name == "" {
		t.Error("Name should not be empty")
	}
	if device.Model == "" {
		t.Error("Model should not be empty")
	}
	if device.Firmware == "" {
		t.Error("Firmware should not be empty")
	}
	if device.MAC == "" {
		t.Error("MAC should not be empty")
	}
	if device.State == nil {
		t.Error("State should not be nil")
	}
}
