package shelly

import (
	"net"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

const (
	testPluginDeviceID   = "device-123"
	testPluginDeviceName = "Test Device"
)

func TestPluginDiscoverer_New(t *testing.T) {
	t.Parallel()

	// Use a virtual path - NewRegistryWithDir doesn't create the directory
	registry := plugins.NewRegistryWithDir("/test/plugins")
	discoverer := NewPluginDiscoverer(registry)

	if discoverer == nil {
		t.Fatal("expected non-nil discoverer")
	}
	if discoverer.registry != registry {
		t.Error("expected registry to be set")
	}
}

func TestPluginDetectionResult_Fields(t *testing.T) {
	t.Parallel()

	result := PluginDetectionResult{
		Detection: &plugins.DeviceDetectionResult{
			Detected:   true,
			DeviceID:   testPluginDeviceID,
			DeviceName: testPluginDeviceName,
			Model:      "TestModel",
			Platform:   "test-platform",
			Firmware:   "1.0.0",
			Components: []plugins.ComponentInfo{
				{Type: "switch", ID: 0, Name: "Switch 0"},
			},
		},
		Plugin:  &plugins.Plugin{Name: "test-plugin"},
		Address: "192.168.1.100",
	}

	if result.Detection.DeviceID != testPluginDeviceID {
		t.Errorf("expected DeviceID %q, got %q", testPluginDeviceID, result.Detection.DeviceID)
	}
	if result.Address != "192.168.1.100" {
		t.Errorf("expected Address '192.168.1.100', got %q", result.Address)
	}
	if result.Plugin.Name != "test-plugin" {
		t.Errorf("expected Plugin name 'test-plugin', got %q", result.Plugin.Name)
	}
}

func TestPluginDiscoveredDevice_Fields(t *testing.T) {
	t.Parallel()

	device := PluginDiscoveredDevice{
		ID:         testPluginDeviceID,
		Name:       testPluginDeviceName,
		Model:      "TestModel",
		Address:    net.ParseIP("192.168.1.100"),
		Platform:   "test-platform",
		Generation: 0,
		Firmware:   "1.0.0",
		AuthEn:     false,
		Added:      true,
		Components: []plugins.ComponentInfo{
			{Type: "switch", ID: 0, Name: "Switch 0"},
		},
	}

	if device.ID != testPluginDeviceID {
		t.Errorf("expected ID %q, got %q", testPluginDeviceID, device.ID)
	}
	if device.Name != testPluginDeviceName {
		t.Errorf("expected Name %q, got %q", testPluginDeviceName, device.Name)
	}
	if device.Platform != "test-platform" {
		t.Errorf("expected Platform 'test-platform', got %q", device.Platform)
	}
	if !device.Added {
		t.Error("expected Added to be true")
	}
	if len(device.Components) != 1 {
		t.Errorf("expected 1 component, got %d", len(device.Components))
	}
}

func TestToPluginDiscoveredDevice(t *testing.T) {
	t.Parallel()

	t.Run("converts all fields", func(t *testing.T) {
		t.Parallel()

		result := &PluginDetectionResult{
			Detection: &plugins.DeviceDetectionResult{
				Detected:   true,
				DeviceID:   testPluginDeviceID,
				DeviceName: testPluginDeviceName,
				Model:      "TestModel",
				Platform:   "test-platform",
				Firmware:   "1.0.0",
				Components: []plugins.ComponentInfo{
					{Type: "switch", ID: 0, Name: "Switch 0"},
					{Type: "light", ID: 1, Name: "Light 1"},
				},
			},
			Plugin:  &plugins.Plugin{Name: "test-plugin"},
			Address: "192.168.1.100",
		}

		device := ToPluginDiscoveredDevice(result, true)

		if device.ID != testPluginDeviceID {
			t.Errorf("expected ID %q, got %q", testPluginDeviceID, device.ID)
		}
		if device.Name != testPluginDeviceName {
			t.Errorf("expected Name %q, got %q", testPluginDeviceName, device.Name)
		}
		if device.Model != "TestModel" {
			t.Errorf("expected Model 'TestModel', got %q", device.Model)
		}
		if device.Platform != "test-platform" {
			t.Errorf("expected Platform 'test-platform', got %q", device.Platform)
		}
		if device.Firmware != "1.0.0" {
			t.Errorf("expected Firmware '1.0.0', got %q", device.Firmware)
		}
		if device.Generation != 0 {
			t.Errorf("expected Generation 0, got %d", device.Generation)
		}
		if device.AuthEn {
			t.Error("expected AuthEn to be false")
		}
		if !device.Added {
			t.Error("expected Added to be true")
		}
		if len(device.Components) != 2 {
			t.Errorf("expected 2 components, got %d", len(device.Components))
		}
	})

	t.Run("not added", func(t *testing.T) {
		t.Parallel()

		result := &PluginDetectionResult{
			Detection: &plugins.DeviceDetectionResult{
				DeviceID: "device-456",
			},
			Address: "192.168.1.101",
		}

		device := ToPluginDiscoveredDevice(result, false)

		if device.Added {
			t.Error("expected Added to be false")
		}
	})
}

func TestGenerateSubnetsAddresses(t *testing.T) {
	t.Parallel()

	t.Run("multiple subnets cover all address spaces", func(t *testing.T) {
		t.Parallel()

		addresses := generateSubnetsAddresses([]string{"192.168.1.0/30", "10.0.0.0/30"})

		// Each /30 yields 2 usable hosts, so two subnets must produce 4.
		if len(addresses) != 4 {
			t.Fatalf("expected 4 addresses across both subnets, got %d: %v", len(addresses), addresses)
		}

		has := func(want string) bool {
			for _, a := range addresses {
				if a == want {
					return true
				}
			}
			return false
		}
		// Addresses from the SECOND subnet must be present — the multi-subnet
		// fix exists precisely so discovery does not stop at subnets[0].
		if !has("10.0.0.1") || !has("10.0.0.2") {
			t.Errorf("expected addresses from second subnet 10.0.0.0/30, got %v", addresses)
		}
		if !has("192.168.1.1") || !has("192.168.1.2") {
			t.Errorf("expected addresses from first subnet 192.168.1.0/30, got %v", addresses)
		}
	})

	t.Run("/24 excludes network and broadcast", func(t *testing.T) {
		t.Parallel()

		// Host enumeration is owned by discovery.GenerateSubnetAddresses; this
		// guards that plugin discovery still skips .0/.255 for a /24 via the SDK.
		addresses := generateSubnetsAddresses([]string{"192.168.1.0/24"})
		if len(addresses) != 254 {
			t.Fatalf("expected 254 addresses for /24, got %d", len(addresses))
		}
		if addresses[0] != "192.168.1.1" {
			t.Errorf("expected first address '192.168.1.1', got %q", addresses[0])
		}
		if addresses[len(addresses)-1] != "192.168.1.254" {
			t.Errorf("expected last address '192.168.1.254', got %q", addresses[len(addresses)-1])
		}
	})

	t.Run("invalid subnet yields no addresses", func(t *testing.T) {
		t.Parallel()

		if addresses := generateSubnetsAddresses([]string{"invalid"}); len(addresses) != 0 {
			t.Errorf("expected 0 addresses for invalid subnet, got %d", len(addresses))
		}
	})

	t.Run("empty slice yields no addresses", func(t *testing.T) {
		t.Parallel()

		if addresses := generateSubnetsAddresses(nil); len(addresses) != 0 {
			t.Errorf("expected 0 addresses for nil subnets, got %d", len(addresses))
		}
	})
}
