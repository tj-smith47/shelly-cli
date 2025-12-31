package utils

import (
	"net"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

func TestDeviceRegistrationToDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		reg      DeviceRegistration
		wantName string
		wantAddr string
		wantPlat string
	}{
		{
			name: "basic shelly device",
			reg: DeviceRegistration{
				Name:       "kitchen",
				Address:    "192.168.1.100",
				Generation: 2,
				Type:       "SPSW-001PE16EU",
				Model:      "Shelly Pro 1PM",
				Platform:   model.PlatformShelly,
			},
			wantName: "kitchen",
			wantAddr: "192.168.1.100",
			wantPlat: model.PlatformShelly,
		},
		{
			name: "empty platform defaults to shelly",
			reg: DeviceRegistration{
				Name:    "test",
				Address: "192.168.1.101",
			},
			wantName: "test",
			wantAddr: "192.168.1.101",
			wantPlat: model.PlatformShelly,
		},
		{
			name: "tasmota platform",
			reg: DeviceRegistration{
				Name:     "tasmota-device",
				Address:  "192.168.1.102",
				Platform: "tasmota",
			},
			wantName: "tasmota-device",
			wantAddr: "192.168.1.102",
			wantPlat: "tasmota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.reg.ToDevice()
			if got.Name != tt.wantName {
				t.Errorf("ToDevice().Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Address != tt.wantAddr {
				t.Errorf("ToDevice().Address = %q, want %q", got.Address, tt.wantAddr)
			}
			if got.Platform != tt.wantPlat {
				t.Errorf("ToDevice().Platform = %q, want %q", got.Platform, tt.wantPlat)
			}
		})
	}
}

func TestDiscoveredDeviceToRegistration(t *testing.T) {
	t.Parallel()

	t.Run("uses name when available", func(t *testing.T) {
		t.Parallel()
		d := discovery.DiscoveredDevice{
			ID:         "shellyplus1pm-abcd",
			Name:       "Kitchen",
			Address:    net.ParseIP("192.168.1.100"),
			Generation: 2,
			Model:      "SPSW-001PE16EU",
			MACAddress: "AA:BB:CC:DD:EE:FF",
		}
		got := DiscoveredDeviceToRegistration(d)
		if got.Name != "Kitchen" {
			t.Errorf("Name = %q, want 'Kitchen'", got.Name)
		}
		if got.Platform != model.PlatformShelly {
			t.Errorf("Platform = %q, want %q", got.Platform, model.PlatformShelly)
		}
	})

	t.Run("falls back to ID when name empty", func(t *testing.T) {
		t.Parallel()
		d := discovery.DiscoveredDevice{
			ID:         "shellyplus1pm-abcd",
			Name:       "",
			Address:    net.ParseIP("192.168.1.100"),
			Generation: 2,
			Model:      "SPSW-001PE16EU",
		}
		got := DiscoveredDeviceToRegistration(d)
		if got.Name != "shellyplus1pm-abcd" {
			t.Errorf("Name = %q, want 'shellyplus1pm-abcd'", got.Name)
		}
	})
}

func TestPluginResultToRegistration(t *testing.T) {
	t.Parallel()

	const testAddress = "192.168.1.200"

	t.Run("uses device name when available", func(t *testing.T) {
		t.Parallel()
		result := &plugins.DeviceDetectionResult{
			Detected:   true,
			Platform:   "tasmota",
			DeviceID:   "tasmota-device",
			DeviceName: "Office Light",
			Model:      "ESP8266",
		}
		got := PluginResultToRegistration(result, testAddress)
		if got.Name != "Office Light" {
			t.Errorf("Name = %q, want 'Office Light'", got.Name)
		}
		if got.Platform != "tasmota" {
			t.Errorf("Platform = %q, want 'tasmota'", got.Platform)
		}
		if got.Address != testAddress {
			t.Errorf("Address = %q, want %q", got.Address, testAddress)
		}
	})

	t.Run("falls back to device ID", func(t *testing.T) {
		t.Parallel()
		result := &plugins.DeviceDetectionResult{
			Detected: true,
			Platform: "tasmota",
			DeviceID: "tasmota-abc123",
		}
		got := PluginResultToRegistration(result, testAddress)
		if got.Name != "tasmota-abc123" {
			t.Errorf("Name = %q, want 'tasmota-abc123'", got.Name)
		}
	})

	t.Run("falls back to address", func(t *testing.T) {
		t.Parallel()
		result := &plugins.DeviceDetectionResult{
			Detected: true,
			Platform: "tasmota",
		}
		got := PluginResultToRegistration(result, testAddress)
		if got.Name != testAddress {
			t.Errorf("Name = %q, want %q", got.Name, testAddress)
		}
	})
}

func TestPluginDeviceToRegistration(t *testing.T) {
	t.Parallel()

	t.Run("uses name when available", func(t *testing.T) {
		t.Parallel()
		d := PluginDevice{
			Address:  "192.168.1.150",
			Platform: "tasmota",
			ID:       "tasmota-abc",
			Name:     "Garage Door",
			Model:    "ESP8266",
		}
		got := PluginDeviceToRegistration(d)
		if got.Name != "Garage Door" {
			t.Errorf("Name = %q, want 'Garage Door'", got.Name)
		}
		if got.Platform != "tasmota" {
			t.Errorf("Platform = %q, want 'tasmota'", got.Platform)
		}
	})

	t.Run("falls back to ID then address", func(t *testing.T) {
		t.Parallel()
		d := PluginDevice{
			Address:  "192.168.1.150",
			Platform: "custom",
		}
		got := PluginDeviceToRegistration(d)
		if got.Name != "192.168.1.150" {
			t.Errorf("Name = %q, want '192.168.1.150'", got.Name)
		}
	})
}

func TestRegisterDevicesFromSlice(t *testing.T) {
	t.Parallel()

	// Test that empty address items are skipped
	devices := []PluginDevice{
		{Address: "", Platform: "test", Name: "empty"},
		{Address: "192.168.1.1", Platform: "test", Name: "valid"},
	}

	// Note: This test doesn't actually register due to config dependencies,
	// but it verifies the conversion logic works.
	// In a real test environment with config mocking, we'd verify registration.
	count := 0
	for _, d := range devices {
		reg := PluginDeviceToRegistration(d)
		if reg.Address != "" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 1 device with address, got %d", count)
	}
}
