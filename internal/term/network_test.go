package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

func TestDisplayWiFiStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &shelly.WiFiStatus{
		Status:  "connected",
		SSID:    "MyNetwork",
		StaIP:   "192.168.1.100",
		RSSI:    -55,
		APCount: 2,
	}

	DisplayWiFiStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "WiFi Status") {
		t.Error("output should contain 'WiFi Status'")
	}
	if !strings.Contains(output, "MyNetwork") {
		t.Error("output should contain SSID")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("output should contain IP address")
	}
	if !strings.Contains(output, "-55") {
		t.Error("output should contain RSSI")
	}
}

func TestDisplayWiFiAPClients(t *testing.T) {
	t.Parallel()

	t.Run("with clients", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		clients := []shelly.WiFiAPClient{
			{MAC: "AA:BB:CC:DD:EE:FF", IP: "192.168.1.10"},
			{MAC: "11:22:33:44:55:66", IP: "192.168.1.11"},
		}

		DisplayWiFiAPClients(ios, clients)

		output := out.String()
		if !strings.Contains(output, "Connected Clients") {
			t.Error("output should contain 'Connected Clients'")
		}
		if !strings.Contains(output, "AA:BB:CC:DD:EE:FF") {
			t.Error("output should contain MAC address")
		}
		if !strings.Contains(output, "2 client(s) connected") {
			t.Error("output should contain client count")
		}
	})

	t.Run("client without IP", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		clients := []shelly.WiFiAPClient{
			{MAC: "AA:BB:CC:DD:EE:FF", IP: ""},
		}

		DisplayWiFiAPClients(ios, clients)

		output := out.String()
		if !strings.Contains(output, "<no IP>") {
			t.Error("output should contain '<no IP>' for clients without IP")
		}
	})
}

func TestDisplayWiFiScanResults(t *testing.T) {
	t.Parallel()

	t.Run("with networks", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		results := []shelly.WiFiScanResult{
			{SSID: "Network1", RSSI: -45, Channel: 6, Auth: "WPA2"},
			{SSID: "Network2", RSSI: -70, Channel: 11, Auth: "Open"},
		}

		DisplayWiFiScanResults(ios, results)

		output := out.String()
		if !strings.Contains(output, "Available WiFi Networks") {
			t.Error("output should contain 'Available WiFi Networks'")
		}
		if !strings.Contains(output, "Network1") {
			t.Error("output should contain 'Network1'")
		}
		if !strings.Contains(output, "2 network(s) found") {
			t.Error("output should contain network count")
		}
	})

	t.Run("hidden network", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		results := []shelly.WiFiScanResult{
			{SSID: "", RSSI: -60, Channel: 1, Auth: "WPA3"},
		}

		DisplayWiFiScanResults(ios, results)

		output := out.String()
		if !strings.Contains(output, "<hidden>") {
			t.Error("output should contain '<hidden>' for networks without SSID")
		}
	})
}

func TestFormatWiFiSignal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rssi     int
		wantBars int
	}{
		{"excellent signal", -45, 4},
		{"good signal", -55, 3},
		{"fair signal", -65, 2},
		{"weak signal", -80, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatWiFiSignal(tt.rssi)
			if result == "" {
				t.Error("formatWiFiSignal should return non-empty string")
			}
			if !strings.Contains(result, "dBm") {
				t.Error("result should contain 'dBm'")
			}
		})
	}
}

func TestDisplayEthernetStatus(t *testing.T) {
	t.Parallel()

	t.Run("connected", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.EthernetStatus{
			IP: "192.168.1.200",
		}

		DisplayEthernetStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Ethernet Status") {
			t.Error("output should contain 'Ethernet Status'")
		}
		if !strings.Contains(output, "Connected") {
			t.Error("output should contain 'Connected'")
		}
		if !strings.Contains(output, "192.168.1.200") {
			t.Error("output should contain IP address")
		}
	})

	t.Run("not connected", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.EthernetStatus{
			IP: "",
		}

		DisplayEthernetStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "Not connected") {
			t.Error("output should contain 'Not connected'")
		}
	})
}

func TestDisplayMQTTStatus(t *testing.T) {
	t.Parallel()

	t.Run("connected", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.MQTTStatus{
			Connected: true,
		}

		DisplayMQTTStatus(ios, status)

		output := out.String()
		if !strings.Contains(output, "MQTT Status") {
			t.Error("output should contain 'MQTT Status'")
		}
	})

	t.Run("disconnected", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		status := &shelly.MQTTStatus{
			Connected: false,
		}

		DisplayMQTTStatus(ios, status)

		output := out.String()
		if output == "" {
			t.Error("DisplayMQTTStatus should produce output")
		}
	})
}

func TestDisplayCloudConnectionStatus(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := &shelly.CloudStatus{
		Connected: true,
	}

	DisplayCloudConnectionStatus(ios, status)

	output := out.String()
	if !strings.Contains(output, "Cloud Status") {
		t.Error("output should contain 'Cloud Status'")
	}
}

func TestDisplayCloudDevices(t *testing.T) {
	t.Parallel()

	t.Run("with devices", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []network.CloudDevice{
			{ID: "device1", Model: "Shelly Pro 1PM", Generation: 2, Online: true},
			{ID: "device2", Model: "Shelly Plus 1", Generation: 2, Online: false},
		}

		DisplayCloudDevices(ios, devices)

		output := out.String()
		if !strings.Contains(output, "Found 2 device(s)") {
			t.Error("output should contain device count")
		}
		if !strings.Contains(output, "device1") {
			t.Error("output should contain 'device1'")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()

		DisplayCloudDevices(ios, []network.CloudDevice{})

		output := out.String()
		if !strings.Contains(output, "No devices found") {
			t.Error("output should contain 'No devices found'")
		}
	})

	t.Run("device without model", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		devices := []network.CloudDevice{
			{ID: "device1", Model: "", Generation: 0, Online: true},
		}

		DisplayCloudDevices(ios, devices)

		output := out.String()
		if output == "" {
			t.Error("DisplayCloudDevices should produce output")
		}
	})
}

func TestDisplayCloudDevice(t *testing.T) {
	t.Parallel()

	t.Run("basic device", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		device := &network.CloudDevice{
			ID:              "shellypro1pm-123456",
			Model:           "Shelly Pro 1PM",
			Generation:      2,
			MAC:             "AA:BB:CC:DD:EE:FF",
			FirmwareVersion: "1.0.0",
			Online:          true,
		}

		DisplayCloudDevice(ios, device, false)

		output := out.String()
		if !strings.Contains(output, "Cloud Device") {
			t.Error("output should contain 'Cloud Device'")
		}
		if !strings.Contains(output, "shellypro1pm-123456") {
			t.Error("output should contain device ID")
		}
	})

	t.Run("with status", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		device := &network.CloudDevice{
			ID:     "device1",
			Online: true,
			Status: []byte(`{"switch:0": {"output": true}}`),
		}

		DisplayCloudDevice(ios, device, true)

		output := out.String()
		if !strings.Contains(output, "Device Status") {
			t.Error("output should contain 'Device Status' when showStatus is true")
		}
	})
}

func TestTokenStatusInfo_Fields(t *testing.T) {
	t.Parallel()

	info := TokenStatusInfo{
		Display: "valid",
		Warning: "",
	}

	if info.Display != "valid" {
		t.Errorf("Display = %q, want 'valid'", info.Display)
	}
	if info.Warning != "" {
		t.Errorf("Warning = %q, want empty", info.Warning)
	}
}

func TestDisplayTLSConfig(t *testing.T) {
	t.Parallel()

	t.Run("with MQTT SSL CA", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		config := map[string]any{
			"mqtt": map[string]any{
				"server": "mqtt.example.com:8883",
				"ssl_ca": "/path/to/ca.crt",
			},
		}

		hasCustomCA := DisplayTLSConfig(ios, config)

		if !hasCustomCA {
			t.Error("DisplayTLSConfig should return true when custom CA is configured")
		}
		output := out.String()
		if !strings.Contains(output, "MQTT") {
			t.Error("output should contain 'MQTT'")
		}
	})

	t.Run("no custom CA", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		config := map[string]any{
			"mqtt": map[string]any{
				"server": "mqtt.example.com:1883",
			},
		}

		hasCustomCA := DisplayTLSConfig(ios, config)

		if hasCustomCA {
			t.Error("DisplayTLSConfig should return false when no custom CA is configured")
		}
		output := out.String()
		if output == "" {
			t.Error("DisplayTLSConfig should produce output")
		}
	})

	t.Run("with cloud config", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		config := map[string]any{
			"cloud": map[string]any{
				"enable": true,
			},
		}

		DisplayTLSConfig(ios, config)

		output := out.String()
		if !strings.Contains(output, "Cloud") {
			t.Error("output should contain 'Cloud'")
		}
	})

	t.Run("with WebSocket SSL CA", func(t *testing.T) {
		t.Parallel()
		ios, out, _ := testIOStreams()
		config := map[string]any{
			"ws": map[string]any{
				"ssl_ca": "/path/to/ws-ca.crt",
			},
		}

		hasCustomCA := DisplayTLSConfig(ios, config)

		if !hasCustomCA {
			t.Error("DisplayTLSConfig should return true when WS custom CA is configured")
		}
		output := out.String()
		if !strings.Contains(output, "WebSocket") {
			t.Error("output should contain 'WebSocket'")
		}
	})
}
