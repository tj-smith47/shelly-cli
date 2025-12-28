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

	registry := plugins.NewRegistryWithDir("/tmp/test-plugins")
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

func TestGenerateSubnetAddresses(t *testing.T) {
	t.Parallel()

	t.Run("valid /24 subnet", func(t *testing.T) {
		t.Parallel()

		addresses := generateSubnetAddresses("192.168.1.0/24")

		// /24 has 256 addresses, minus network (0) and broadcast (255) = 254
		if len(addresses) != 254 {
			t.Errorf("expected 254 addresses, got %d", len(addresses))
		}

		// First should be .1
		if addresses[0] != "192.168.1.1" {
			t.Errorf("expected first address '192.168.1.1', got %q", addresses[0])
		}

		// Last should be .254
		if addresses[len(addresses)-1] != "192.168.1.254" {
			t.Errorf("expected last address '192.168.1.254', got %q", addresses[len(addresses)-1])
		}
	})

	t.Run("valid /30 subnet", func(t *testing.T) {
		t.Parallel()

		addresses := generateSubnetAddresses("192.168.1.0/30")

		// /30 has 4 addresses, minus network and broadcast = 2
		if len(addresses) != 2 {
			t.Errorf("expected 2 addresses, got %d", len(addresses))
		}
	})

	t.Run("invalid subnet", func(t *testing.T) {
		t.Parallel()

		addresses := generateSubnetAddresses("invalid")

		if len(addresses) != 0 {
			t.Errorf("expected 0 addresses for invalid subnet, got %d", len(addresses))
		}
	})

	t.Run("empty subnet", func(t *testing.T) {
		t.Parallel()

		addresses := generateSubnetAddresses("")

		if len(addresses) != 0 {
			t.Errorf("expected 0 addresses for empty subnet, got %d", len(addresses))
		}
	})
}

func TestIncIP(t *testing.T) {
	t.Parallel()

	t.Run("simple increment", func(t *testing.T) {
		t.Parallel()

		ip := net.ParseIP("192.168.1.1").To4()
		incIP(ip)

		if ip.String() != "192.168.1.2" {
			t.Errorf("expected '192.168.1.2', got %q", ip.String())
		}
	})

	t.Run("octet rollover", func(t *testing.T) {
		t.Parallel()

		ip := net.ParseIP("192.168.1.255").To4()
		incIP(ip)

		if ip.String() != "192.168.2.0" {
			t.Errorf("expected '192.168.2.0', got %q", ip.String())
		}
	})

	t.Run("multiple octet rollover", func(t *testing.T) {
		t.Parallel()

		ip := net.ParseIP("192.168.255.255").To4()
		incIP(ip)

		if ip.String() != "192.169.0.0" {
			t.Errorf("expected '192.169.0.0', got %q", ip.String())
		}
	})
}

func TestIsNetworkOrBroadcast(t *testing.T) {
	t.Parallel()

	t.Run("/24 network address", func(t *testing.T) {
		t.Parallel()

		//nolint:errcheck // test uses known-valid CIDR
		_, ipNet, _ := net.ParseCIDR("192.168.1.0/24")
		ip := net.ParseIP("192.168.1.0").To4()

		if !isNetworkOrBroadcast(ip, ipNet) {
			t.Error("expected .0 to be network address")
		}
	})

	t.Run("/24 broadcast address", func(t *testing.T) {
		t.Parallel()

		//nolint:errcheck // test uses known-valid CIDR
		_, ipNet, _ := net.ParseCIDR("192.168.1.0/24")
		ip := net.ParseIP("192.168.1.255").To4()

		if !isNetworkOrBroadcast(ip, ipNet) {
			t.Error("expected .255 to be broadcast address")
		}
	})

	t.Run("/24 regular address", func(t *testing.T) {
		t.Parallel()

		//nolint:errcheck // test uses known-valid CIDR
		_, ipNet, _ := net.ParseCIDR("192.168.1.0/24")
		ip := net.ParseIP("192.168.1.100").To4()

		if isNetworkOrBroadcast(ip, ipNet) {
			t.Error("expected .100 to NOT be network or broadcast")
		}
	})

	t.Run("/31 has no network/broadcast", func(t *testing.T) {
		t.Parallel()

		//nolint:errcheck // test uses known-valid CIDR
		_, ipNet, _ := net.ParseCIDR("192.168.1.0/31")
		ip := net.ParseIP("192.168.1.0").To4()

		if isNetworkOrBroadcast(ip, ipNet) {
			t.Error("/31 should not exclude any addresses")
		}
	})

	t.Run("/32 has no network/broadcast", func(t *testing.T) {
		t.Parallel()

		//nolint:errcheck // test uses known-valid CIDR
		_, ipNet, _ := net.ParseCIDR("192.168.1.1/32")
		ip := net.ParseIP("192.168.1.1").To4()

		if isNetworkOrBroadcast(ip, ipNet) {
			t.Error("/32 should not exclude any addresses")
		}
	})
}
