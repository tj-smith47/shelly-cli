package term

import (
	"net"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/discovery"
)

func TestDisplayDiscoveredDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDiscoveredDevices(ios, []discovery.DiscoveredDevice{})

	output := out.String()
	if !strings.Contains(output, "devices") {
		t.Error("expected NoResults message")
	}
}

func TestDisplayDiscoveredDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []discovery.DiscoveredDevice{
		{
			Name:    "shellyplus1-abc123",
			Model:   "SNSW-001X16EU",
			Address: net.ParseIP("192.168.1.100"),
		},
		{
			Name:    "shellyplus2pm-def456",
			Model:   "SNSW-002P16EU",
			Address: net.ParseIP("192.168.1.101"),
		},
	}
	DisplayDiscoveredDevices(ios, devices)

	output := out.String()
	if output == "" {
		t.Error("expected output")
	}
	// Count should be displayed
	if !strings.Contains(output, "2") {
		t.Error("expected device count")
	}
}

func TestDisplayBLEDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayBLEDevices(ios, []discovery.BLEDiscoveredDevice{})

	if out.String() != "" {
		t.Error("expected no output for empty devices")
	}
}

func TestDisplayBLEDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	mac, err := net.ParseMAC("AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("ParseMAC error: %v", err)
	}
	devices := []discovery.BLEDiscoveredDevice{
		{
			DiscoveredDevice: discovery.DiscoveredDevice{
				ID:    "device1",
				Model: "Shelly BLU",
			},
			LocalName:   "BTHome Sensor",
			RSSI:        -45,
			Connectable: true,
			BTHomeData:  &discovery.BTHomeData{},
		},
		{
			DiscoveredDevice: discovery.DiscoveredDevice{
				ID:    "device2",
				Model: "Unknown",
			},
			LocalName:   "",
			RSSI:        -75,
			Connectable: false,
			BTHomeData:  nil,
		},
		{
			DiscoveredDevice: discovery.DiscoveredDevice{
				ID:    "device3",
				Model: "Sensor",
			},
			LocalName:   "Weak Signal",
			RSSI:        -85,
			Connectable: true,
		},
	}
	_ = mac // Address is net.IP in embedded struct, not used for BLE
	DisplayBLEDevices(ios, devices)

	output := out.String()
	if !strings.Contains(output, "BTHome Sensor") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "dBm") {
		t.Error("expected RSSI unit")
	}
	if !strings.Contains(output, "BLE device") {
		t.Error("expected count suffix")
	}
}

func TestDisplayPluginDiscoveredDevices_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayPluginDiscoveredDevices(ios, []PluginDiscoveredDevice{})

	if out.String() != "" {
		t.Error("expected no output for empty devices")
	}
}

func TestDisplayPluginDiscoveredDevices_WithDevices(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []PluginDiscoveredDevice{
		{
			ID:       "tasmota-abc123",
			Name:     "Living Room Light",
			Model:    "Generic",
			Address:  "192.168.1.50",
			Platform: "tasmota",
			Firmware: "12.0.0",
			Components: []PluginComponentInfo{
				{Type: "switch", ID: 0, Name: "Light"},
				{Type: "sensor", ID: 1, Name: ""},
			},
		},
	}
	DisplayPluginDiscoveredDevices(ios, devices)

	output := out.String()
	if !strings.Contains(output, "Plugin-managed Devices") {
		t.Error("expected title")
	}
	if !strings.Contains(output, "Living Room Light") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "192.168.1.50") {
		t.Error("expected address")
	}
	if !strings.Contains(output, "tasmota") {
		t.Error("expected platform")
	}
	if !strings.Contains(output, "12.0.0") {
		t.Error("expected firmware")
	}
	if !strings.Contains(output, "switch:0") {
		t.Error("expected component type and ID")
	}
	if !strings.Contains(output, "(Light)") {
		t.Error("expected component name")
	}
}

func TestDisplayPluginDiscoveredDevices_NoName(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []PluginDiscoveredDevice{
		{
			ID:       "device-id",
			Name:     "",
			Address:  "192.168.1.60",
			Platform: "esphome",
		},
	}
	DisplayPluginDiscoveredDevices(ios, devices)

	output := out.String()
	// Should use ID when name is empty
	if !strings.Contains(output, "device-id") {
		t.Error("expected ID as fallback name")
	}
}

func TestDisplayPluginDiscoveredDevices_OnlyAddress(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	devices := []PluginDiscoveredDevice{
		{
			ID:       "",
			Name:     "",
			Address:  "192.168.1.70",
			Platform: "custom",
		},
	}
	DisplayPluginDiscoveredDevices(ios, devices)

	output := out.String()
	// Should use address when both ID and name are empty
	if !strings.Contains(output, "192.168.1.70") {
		t.Error("expected address as fallback name")
	}
}
