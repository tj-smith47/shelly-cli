package shelly

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testIP  = "192.168.1.100"
	testMAC = "AA:BB:CC:DD:EE:FF"
)

func TestIsConfigFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"json extension", "config.json", true},
		{"yaml extension", "config.yaml", true},
		{"yml extension", "config.yml", true},
		{"no extension", "somedevice", false},
		{"device ip", "192.168.1.1", false},
		{"device hostname", "shellyplus1pm-aabbcc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsConfigFile(tt.path)
			if got != tt.want {
				t.Errorf("IsConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsConfigFile_ExistingFile(t *testing.T) {
	t.Parallel()

	// Create a temp file without config extension
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existing-file")
	if err := os.WriteFile(tmpFile, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if !IsConfigFile(tmpFile) {
		t.Error("IsConfigFile should return true for existing file")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "config.json")
		content := `{"switch:0":{"name":"test"}}`
		if err := os.WriteFile(tmpFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		cfg, name, err := LoadConfigFromFile(tmpFile)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if name != tmpFile {
			t.Errorf("expected name %q, got %q", tmpFile, name)
		}
		if cfg["switch:0"] == nil {
			t.Error("config should contain switch:0")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		t.Parallel()

		_, _, err := LoadConfigFromFile("/nonexistent/file.json")

		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0o600); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, _, err := LoadConfigFromFile(tmpFile)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestDedupeWiFiNetworks(t *testing.T) {
	t.Parallel()

	t.Run("deduplicates by SSID", func(t *testing.T) {
		t.Parallel()

		results := []WiFiScanResult{
			{SSID: "Network1", RSSI: -60},
			{SSID: "Network1", RSSI: -50}, // Stronger signal
			{SSID: "Network2", RSSI: -70},
		}

		deduped := DedupeWiFiNetworks(results)

		if len(deduped) != 2 {
			t.Errorf("expected 2 networks, got %d", len(deduped))
		}

		// Find Network1 and verify it has the stronger signal
		for _, n := range deduped {
			if n.SSID == "Network1" && n.RSSI != -50 {
				t.Error("should keep the stronger signal (-50)")
			}
		}
	})

	t.Run("skips empty SSIDs", func(t *testing.T) {
		t.Parallel()

		results := []WiFiScanResult{
			{SSID: "", RSSI: -40},
			{SSID: "ValidNetwork", RSSI: -60},
		}

		deduped := DedupeWiFiNetworks(results)

		if len(deduped) != 1 {
			t.Errorf("expected 1 network, got %d", len(deduped))
		}
	})

	t.Run("sorts by signal strength", func(t *testing.T) {
		t.Parallel()

		results := []WiFiScanResult{
			{SSID: "Weak", RSSI: -80},
			{SSID: "Strong", RSSI: -40},
			{SSID: "Medium", RSSI: -60},
		}

		deduped := DedupeWiFiNetworks(results)

		if len(deduped) != 3 {
			t.Errorf("expected 3 networks, got %d", len(deduped))
		}
		if deduped[0].SSID != "Strong" {
			t.Error("first network should be Strong")
		}
		if deduped[2].SSID != "Weak" {
			t.Error("last network should be Weak")
		}
	})

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		deduped := DedupeWiFiNetworks([]WiFiScanResult{})

		if len(deduped) != 0 {
			t.Errorf("expected 0 networks, got %d", len(deduped))
		}
	})
}

func TestWiFiStatus_Fields(t *testing.T) {
	t.Parallel()

	status := WiFiStatus{
		StaIP:   testIP,
		Status:  "got ip",
		SSID:    "MyNetwork",
		RSSI:    -55,
		APCount: 2,
	}

	if status.StaIP != testIP {
		t.Error("unexpected StaIP")
	}
	if status.Status != "got ip" {
		t.Error("unexpected Status")
	}
	if status.RSSI != -55 {
		t.Error("unexpected RSSI")
	}
}

func TestWiFiScanResult_Fields(t *testing.T) {
	t.Parallel()

	result := WiFiScanResult{
		SSID:    "TestNetwork",
		BSSID:   testMAC,
		RSSI:    -65,
		Channel: 6,
		Auth:    "WPA2-PSK",
	}

	if result.SSID != "TestNetwork" {
		t.Error("unexpected SSID")
	}
	if result.Channel != 6 {
		t.Error("unexpected Channel")
	}
	if result.Auth != "WPA2-PSK" {
		t.Error("unexpected Auth")
	}
}

func TestCloudStatus_Fields(t *testing.T) {
	t.Parallel()

	status := CloudStatus{Connected: true}

	if !status.Connected {
		t.Error("expected Connected to be true")
	}
}

func TestWebhookInfo_ConfigTest_Fields(t *testing.T) {
	t.Parallel()

	info := WebhookInfo{
		ID:     1,
		Name:   "Test Hook",
		Event:  "switch.on",
		Enable: true,
		URLs:   []string{"http://example.com/hook"},
		Cid:    0,
	}

	if info.ID != 1 {
		t.Error("unexpected ID")
	}
	if info.Event != "switch.on" {
		t.Error("unexpected Event")
	}
	if len(info.URLs) != 1 {
		t.Error("unexpected URLs count")
	}
}

func TestCreateWebhookParams_Fields(t *testing.T) {
	t.Parallel()

	params := CreateWebhookParams{
		Event:  "switch.off",
		URLs:   []string{"http://example.com"},
		Name:   "My Hook",
		Enable: true,
		Cid:    0,
	}

	if params.Event != "switch.off" {
		t.Error("unexpected Event")
	}
	if !params.Enable {
		t.Error("expected Enable to be true")
	}
}

func TestBLEConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := BLEConfig{
		Enable:       true,
		RPCEnabled:   true,
		ObserverMode: false,
	}

	if !cfg.Enable {
		t.Error("expected Enable to be true")
	}
	if !cfg.RPCEnabled {
		t.Error("expected RPCEnabled to be true")
	}
	if cfg.ObserverMode {
		t.Error("expected ObserverMode to be false")
	}
}

func TestWebSocketInfo_Fields(t *testing.T) {
	t.Parallel()

	info := WebSocketInfo{
		Config: map[string]any{"enable": true},
		Status: map[string]any{"connected": true},
	}

	if info.Config["enable"] != true {
		t.Error("unexpected Config")
	}
	if info.Status["connected"] != true {
		t.Error("unexpected Status")
	}
}

func TestWiFiAPClient_Fields(t *testing.T) {
	t.Parallel()

	client := WiFiAPClient{
		MAC:   testMAC,
		IP:    "192.168.33.100",
		Since: 1700000000,
	}

	if client.MAC != testMAC {
		t.Error("unexpected MAC")
	}
	if client.IP != "192.168.33.100" {
		t.Error("unexpected IP")
	}
}

func TestTUIMatterStatus_Fields(t *testing.T) {
	t.Parallel()

	status := TUIMatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   1,
	}

	if !status.Enabled {
		t.Error("expected Enabled")
	}
	if status.Commissionable {
		t.Error("expected not Commissionable")
	}
	if status.FabricsCount != 1 {
		t.Error("unexpected FabricsCount")
	}
}

func TestTUIZigbeeStatus_Fields(t *testing.T) {
	t.Parallel()

	status := TUIZigbeeStatus{
		Enabled:      true,
		NetworkState: "joined",
		Channel:      15,
		PANID:        0x1234,
	}

	if !status.Enabled {
		t.Error("expected Enabled")
	}
	if status.NetworkState != "joined" {
		t.Error("unexpected NetworkState")
	}
}

func TestTUILoRaStatus_Fields(t *testing.T) {
	t.Parallel()

	status := TUILoRaStatus{
		Enabled:   true,
		Frequency: 868000000,
		TxPower:   14,
		RSSI:      -90,
		SNR:       5.5,
	}

	if !status.Enabled {
		t.Error("expected Enabled")
	}
	if status.Frequency != 868000000 {
		t.Error("unexpected Frequency")
	}
}

func TestTUISecurityStatus_Fields(t *testing.T) {
	t.Parallel()

	status := TUISecurityStatus{
		AuthEnabled:  true,
		EcoMode:      false,
		Discoverable: true,
		DebugMQTT:    false,
		DebugWS:      true,
		DebugUDP:     true,
		DebugUDPAddr: testIP + ":9999",
	}

	if !status.AuthEnabled {
		t.Error("expected AuthEnabled")
	}
	if !status.Discoverable {
		t.Error("expected Discoverable")
	}
	if !status.DebugUDP {
		t.Error("expected DebugUDP")
	}
}
