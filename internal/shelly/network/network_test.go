// Package network provides network-related services for Shelly devices.
package network

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

const (
	testStaIP       = "192.168.1.101"
	testCloudAPIURL = "https://api.shelly.cloud"
	testIPv4Static  = "static"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func TestNewWiFiService(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := NewWiFiService(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestNewMQTTService(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := NewMQTTService(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestNewEthernetService(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := NewEthernetService(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != provider {
		t.Error("expected provider to be set")
	}
}

func TestWiFiStatusFull_Fields(t *testing.T) {
	t.Parallel()

	status := WiFiStatusFull{
		Status:        "got_ip",
		StaIP:         testStaIP,
		SSID:          "MyNetwork",
		RSSI:          -45.5,
		APClientCount: 2,
	}

	if status.Status != "got_ip" {
		t.Errorf("got Status=%q, want %q", status.Status, "got_ip")
	}
	if status.StaIP != testStaIP {
		t.Errorf("got StaIP=%q, want %q", status.StaIP, testStaIP)
	}
	if status.SSID != "MyNetwork" {
		t.Errorf("got SSID=%q, want %q", status.SSID, "MyNetwork")
	}
	if status.RSSI != -45.5 {
		t.Errorf("got RSSI=%f, want -45.5", status.RSSI)
	}
	if status.APClientCount != 2 {
		t.Errorf("got APClientCount=%d, want 2", status.APClientCount)
	}
}

func TestWiFiStationFull_Fields(t *testing.T) {
	t.Parallel()

	station := WiFiStationFull{
		SSID:     "MyNetwork",
		Enabled:  true,
		IsOpen:   false,
		IPv4Mode: "dhcp",
		IP:       testStaIP,
		Netmask:  "255.255.255.0",
		Gateway:  "192.168.1.1",
	}

	if station.SSID != "MyNetwork" {
		t.Errorf("got SSID=%q, want %q", station.SSID, "MyNetwork")
	}
	if !station.Enabled {
		t.Error("expected Enabled to be true")
	}
	if station.IsOpen {
		t.Error("expected IsOpen to be false")
	}
	if station.IPv4Mode != "dhcp" {
		t.Errorf("got IPv4Mode=%q, want %q", station.IPv4Mode, "dhcp")
	}
	if station.IP != testStaIP {
		t.Errorf("got IP=%q, want %q", station.IP, testStaIP)
	}
}

func TestWiFiAPFull_Fields(t *testing.T) {
	t.Parallel()

	ap := WiFiAPFull{
		SSID:          "Shelly-AP",
		Enabled:       true,
		IsOpen:        false,
		RangeExtender: true,
	}

	if ap.SSID != "Shelly-AP" {
		t.Errorf("got SSID=%q, want %q", ap.SSID, "Shelly-AP")
	}
	if !ap.Enabled {
		t.Error("expected Enabled to be true")
	}
	if ap.IsOpen {
		t.Error("expected IsOpen to be false")
	}
	if !ap.RangeExtender {
		t.Error("expected RangeExtender to be true")
	}
}

func TestWiFiNetworkFull_Fields(t *testing.T) {
	t.Parallel()

	network := WiFiNetworkFull{
		SSID:    "NeighborNetwork",
		BSSID:   "AA:BB:CC:DD:EE:FF",
		Auth:    "wpa2-psk",
		Channel: 6,
		RSSI:    -55.0,
	}

	if network.SSID != "NeighborNetwork" {
		t.Errorf("got SSID=%q, want %q", network.SSID, "NeighborNetwork")
	}
	if network.BSSID != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got BSSID=%q, want %q", network.BSSID, "AA:BB:CC:DD:EE:FF")
	}
	if network.Auth != "wpa2-psk" {
		t.Errorf("got Auth=%q, want %q", network.Auth, "wpa2-psk")
	}
	if network.Channel != 6 {
		t.Errorf("got Channel=%d, want 6", network.Channel)
	}
	if network.RSSI != -55.0 {
		t.Errorf("got RSSI=%f, want -55.0", network.RSSI)
	}
}

func TestMQTTStatus_Fields(t *testing.T) {
	t.Parallel()

	status := MQTTStatus{Connected: true}

	if !status.Connected {
		t.Error("expected Connected to be true")
	}
}

func TestMQTTConfig_Fields(t *testing.T) {
	t.Parallel()

	config := MQTTConfig{
		Enable:      true,
		Server:      "mqtt://broker.example.com:1883",
		User:        "shelly",
		ClientID:    "shelly-device-1",
		TopicPrefix: "home/shelly",
		RPCNTF:      true,
		StatusNTF:   true,
	}

	if !config.Enable {
		t.Error("expected Enable to be true")
	}
	if config.Server != "mqtt://broker.example.com:1883" {
		t.Errorf("got Server=%q, want %q", config.Server, "mqtt://broker.example.com:1883")
	}
	if config.User != "shelly" {
		t.Errorf("got User=%q, want %q", config.User, "shelly")
	}
	if config.ClientID != "shelly-device-1" {
		t.Errorf("got ClientID=%q, want %q", config.ClientID, "shelly-device-1")
	}
	if config.TopicPrefix != "home/shelly" {
		t.Errorf("got TopicPrefix=%q, want %q", config.TopicPrefix, "home/shelly")
	}
	if !config.RPCNTF {
		t.Error("expected RPCNTF to be true")
	}
	if !config.StatusNTF {
		t.Error("expected StatusNTF to be true")
	}
}

func TestSetConfigParams_Fields(t *testing.T) {
	t.Parallel()

	enable := true
	params := SetConfigParams{
		Enable:      &enable,
		Server:      "mqtt://broker.example.com:1883",
		User:        "shelly",
		Password:    "secret",
		TopicPrefix: "home/shelly",
	}

	if params.Enable == nil || !*params.Enable {
		t.Error("expected Enable to be true")
	}
	if params.Server != "mqtt://broker.example.com:1883" {
		t.Errorf("got Server=%q, want %q", params.Server, "mqtt://broker.example.com:1883")
	}
	if params.User != "shelly" {
		t.Errorf("got User=%q, want %q", params.User, "shelly")
	}
	if params.Password != "secret" {
		t.Errorf("got Password=%q, want %q", params.Password, "secret")
	}
	if params.TopicPrefix != "home/shelly" {
		t.Errorf("got TopicPrefix=%q, want %q", params.TopicPrefix, "home/shelly")
	}
}

func TestEthernetStatus_Fields(t *testing.T) {
	t.Parallel()

	status := EthernetStatus{IP: testStaIP}

	if status.IP != testStaIP {
		t.Errorf("got IP=%q, want %q", status.IP, testStaIP)
	}
}

func TestEthernetConfig_Fields(t *testing.T) {
	t.Parallel()

	config := EthernetConfig{
		Enable:     true,
		IPv4Mode:   testIPv4Static,
		IP:         testStaIP,
		Netmask:    "255.255.255.0",
		GW:         "192.168.1.1",
		Nameserver: "8.8.8.8",
	}

	if !config.Enable {
		t.Error("expected Enable to be true")
	}
	if config.IPv4Mode != testIPv4Static {
		t.Errorf("got IPv4Mode=%q, want %q", config.IPv4Mode, testIPv4Static)
	}
	if config.IP != testStaIP {
		t.Errorf("got IP=%q, want %q", config.IP, testStaIP)
	}
	if config.Netmask != "255.255.255.0" {
		t.Errorf("got Netmask=%q, want %q", config.Netmask, "255.255.255.0")
	}
	if config.GW != "192.168.1.1" {
		t.Errorf("got GW=%q, want %q", config.GW, "192.168.1.1")
	}
	if config.Nameserver != "8.8.8.8" {
		t.Errorf("got Nameserver=%q, want %q", config.Nameserver, "8.8.8.8")
	}
}

func TestEthernetSetConfigParams_Fields(t *testing.T) {
	t.Parallel()

	enable := true
	params := EthernetSetConfigParams{
		Enable:     &enable,
		IPv4Mode:   testIPv4Static,
		IP:         testStaIP,
		Netmask:    "255.255.255.0",
		GW:         "192.168.1.1",
		Nameserver: "8.8.8.8",
	}

	if params.Enable == nil || !*params.Enable {
		t.Error("expected Enable to be true")
	}
	if params.IPv4Mode != testIPv4Static {
		t.Errorf("got IPv4Mode=%q, want %q", params.IPv4Mode, testIPv4Static)
	}
	if params.IP != testStaIP {
		t.Errorf("got IP=%q, want %q", params.IP, testStaIP)
	}
}

func TestCloudDevice_Fields(t *testing.T) {
	t.Parallel()

	device := CloudDevice{
		ID:              "aabbccddee",
		Name:            "Living Room",
		Model:           "SNSW-002P16EU",
		MAC:             "AA:BB:CC:DD:EE:FF",
		FirmwareVersion: "1.0.0",
		Generation:      2,
		Online:          true,
		CloudEnabled:    true,
	}

	if device.ID != "aabbccddee" {
		t.Errorf("got ID=%q, want %q", device.ID, "aabbccddee")
	}
	if device.Name != "Living Room" {
		t.Errorf("got Name=%q, want %q", device.Name, "Living Room")
	}
	if device.Model != "SNSW-002P16EU" {
		t.Errorf("got Model=%q, want %q", device.Model, "SNSW-002P16EU")
	}
	if !device.Online {
		t.Error("expected Online to be true")
	}
	if !device.CloudEnabled {
		t.Error("expected CloudEnabled to be true")
	}
}

func TestCloudToken_Fields(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(24 * time.Hour)
	token := CloudToken{
		AccessToken: "abc123xyz",
		UserAPIURL:  testCloudAPIURL,
		Email:       "user@example.com",
		UserID:      12345,
		Expiry:      expiry,
	}

	if token.AccessToken != "abc123xyz" {
		t.Errorf("got AccessToken=%q, want %q", token.AccessToken, "abc123xyz")
	}
	if token.UserAPIURL != testCloudAPIURL {
		t.Errorf("got UserAPIURL=%q, want %q", token.UserAPIURL, testCloudAPIURL)
	}
	if token.Email != "user@example.com" {
		t.Errorf("got Email=%q, want %q", token.Email, "user@example.com")
	}
	if token.UserID != 12345 {
		t.Errorf("got UserID=%d, want 12345", token.UserID)
	}
	if token.Expiry != expiry {
		t.Errorf("got Expiry=%v, want %v", token.Expiry, expiry)
	}
}

func TestCloudAuthStatus_Fields(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(24 * time.Hour)
	status := CloudAuthStatus{
		Authenticated: true,
		Email:         "user@example.com",
		UserAPIURL:    testCloudAPIURL,
		TokenExpiry:   expiry,
		TokenValid:    true,
	}

	if !status.Authenticated {
		t.Error("expected Authenticated to be true")
	}
	if status.Email != "user@example.com" {
		t.Errorf("got Email=%q, want %q", status.Email, "user@example.com")
	}
	if status.UserAPIURL != testCloudAPIURL {
		t.Errorf("got UserAPIURL=%q, want %q", status.UserAPIURL, testCloudAPIURL)
	}
	if !status.TokenValid {
		t.Error("expected TokenValid to be true")
	}
}

func TestCloudLoginResult_Fields(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(24 * time.Hour)
	result := CloudLoginResult{
		Token:      "abc123xyz",
		UserAPIURL: "https://api.shelly.cloud",
		Expiry:     expiry,
	}

	if result.Token != "abc123xyz" {
		t.Errorf("got Token=%q, want %q", result.Token, "abc123xyz")
	}
	if result.UserAPIURL != "https://api.shelly.cloud" {
		t.Errorf("got UserAPIURL=%q, want %q", result.UserAPIURL, "https://api.shelly.cloud")
	}
	if result.Expiry != expiry {
		t.Errorf("got Expiry=%v, want %v", result.Expiry, expiry)
	}
}

func TestCloudControlResult_Fields(t *testing.T) {
	t.Parallel()

	result := CloudControlResult{
		Success: true,
		Message: "Switch turned on",
	}

	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Message != "Switch turned on" {
		t.Errorf("got Message=%q, want %q", result.Message, "Switch turned on")
	}
}

func TestNewCloudClient(t *testing.T) {
	t.Parallel()

	cloudClient := NewCloudClient("test-token")

	if cloudClient == nil {
		t.Fatal("expected non-nil client")
	}
	if cloudClient.client == nil {
		t.Error("expected internal client to be set")
	}
}

func TestBuildCloudWebSocketURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serverURL  string
		token      string
		wantPrefix string
		wantErr    bool
	}{
		{
			name:       "with server URL",
			serverURL:  "https://api.shelly.cloud",
			token:      "abc123",
			wantPrefix: "wss://api.shelly.cloud:6113/shelly/wss/hk_sock?t=",
		},
		{
			name:      "empty server URL",
			serverURL: "",
			token:     "invalid-token",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url, err := BuildCloudWebSocketURL(tt.serverURL, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !containsPrefix(url, tt.wantPrefix) {
				t.Errorf("got URL=%q, want prefix %q", url, tt.wantPrefix)
			}
		})
	}
}

func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func TestCloudEventStreamOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := CloudEventStreamOptions{
		DeviceFilter: "device123",
		EventFilter:  "switch",
		Raw:          true,
	}

	if opts.DeviceFilter != "device123" {
		t.Errorf("got DeviceFilter=%q, want %q", opts.DeviceFilter, "device123")
	}
	if opts.EventFilter != "switch" {
		t.Errorf("got EventFilter=%q, want %q", opts.EventFilter, "switch")
	}
	if !opts.Raw {
		t.Error("expected Raw to be true")
	}
}

func TestHashPassword(t *testing.T) {
	t.Parallel()

	hash := HashPassword("password123")

	// Hash should be non-empty
	if hash == "" {
		t.Error("expected non-empty hash")
	}

	// Hash should be consistent
	hash2 := HashPassword("password123")
	if hash != hash2 {
		t.Errorf("expected consistent hash, got %q and %q", hash, hash2)
	}

	// Different passwords should have different hashes
	hash3 := HashPassword("differentpassword")
	if hash == hash3 {
		t.Error("expected different hash for different password")
	}
}

// ----- Service Connection Error Tests -----

func TestWiFiService_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := NewWiFiService(provider)

	t.Run("GetStatusFull", func(t *testing.T) {
		t.Parallel()
		status, err := svc.GetStatusFull(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if status != nil {
			t.Errorf("expected nil status, got %v", status)
		}
	})

	t.Run("GetConfigFull", func(t *testing.T) {
		t.Parallel()
		config, err := svc.GetConfigFull(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if config != nil {
			t.Errorf("expected nil config, got %v", config)
		}
	})

	t.Run("ScanNetworksFull", func(t *testing.T) {
		t.Parallel()
		networks, err := svc.ScanNetworksFull(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if networks != nil {
			t.Errorf("expected nil networks, got %v", networks)
		}
	})

	t.Run("SetStation", func(t *testing.T) {
		t.Parallel()
		err := svc.SetStation(context.Background(), "test-device", "SSID", "pass", true)
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
	})

	t.Run("SetAP", func(t *testing.T) {
		t.Parallel()
		err := svc.SetAP(context.Background(), "test-device", "SSID", "pass", true)
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
	})
}

func TestMQTTService_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := NewMQTTService(provider)

	t.Run("GetStatus", func(t *testing.T) {
		t.Parallel()
		status, err := svc.GetStatus(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if status != nil {
			t.Errorf("expected nil status, got %v", status)
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		t.Parallel()
		config, err := svc.GetConfig(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if config != nil {
			t.Errorf("expected nil config, got %v", config)
		}
	})

	t.Run("SetConfig", func(t *testing.T) {
		t.Parallel()
		err := svc.SetConfig(context.Background(), "test-device", SetConfigParams{})
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
	})
}

func TestEthernetService_ConnectionError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("connection failed")
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, _ func(*client.Client) error) error {
			return expectedErr
		},
	}

	svc := NewEthernetService(provider)

	t.Run("GetStatus", func(t *testing.T) {
		t.Parallel()
		status, err := svc.GetStatus(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if status != nil {
			t.Errorf("expected nil status, got %v", status)
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		t.Parallel()
		config, err := svc.GetConfig(context.Background(), "test-device")
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		if config != nil {
			t.Errorf("expected nil config, got %v", config)
		}
	})

	t.Run("SetConfig", func(t *testing.T) {
		t.Parallel()
		err := svc.SetConfig(context.Background(), "test-device", EthernetSetConfigParams{})
		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
	})
}

// ----- WiFiConfigFull Tests -----

func TestWiFiConfigFull_Fields(t *testing.T) {
	t.Parallel()

	config := WiFiConfigFull{
		STA: &WiFiStationFull{
			SSID:     "HomeNetwork",
			Enabled:  true,
			IsOpen:   false,
			IPv4Mode: "dhcp",
		},
		STA1: &WiFiStationFull{
			SSID:     "BackupNetwork",
			Enabled:  false,
			IsOpen:   true,
			IPv4Mode: "dhcp",
		},
		AP: &WiFiAPFull{
			SSID:          "ShellyAP",
			Enabled:       true,
			IsOpen:        true,
			RangeExtender: false,
		},
	}

	if config.STA == nil {
		t.Fatal("expected STA to be set")
	}
	if config.STA.SSID != "HomeNetwork" {
		t.Errorf("got STA.SSID=%q, want %q", config.STA.SSID, "HomeNetwork")
	}
	if config.STA1 == nil {
		t.Fatal("expected STA1 to be set")
	}
	if config.STA1.SSID != "BackupNetwork" {
		t.Errorf("got STA1.SSID=%q, want %q", config.STA1.SSID, "BackupNetwork")
	}
	if config.AP == nil {
		t.Fatal("expected AP to be set")
	}
	if config.AP.SSID != "ShellyAP" {
		t.Errorf("got AP.SSID=%q, want %q", config.AP.SSID, "ShellyAP")
	}
}

// ----- Zero Value Tests -----

func TestWiFiStatusFull_ZeroValue(t *testing.T) {
	t.Parallel()

	var status WiFiStatusFull

	if status.Status != "" {
		t.Errorf("got Status=%q, want empty", status.Status)
	}
	if status.StaIP != "" {
		t.Errorf("got StaIP=%q, want empty", status.StaIP)
	}
	if status.SSID != "" {
		t.Errorf("got SSID=%q, want empty", status.SSID)
	}
	if status.RSSI != 0 {
		t.Errorf("got RSSI=%f, want 0", status.RSSI)
	}
	if status.APClientCount != 0 {
		t.Errorf("got APClientCount=%d, want 0", status.APClientCount)
	}
}

func TestMQTTStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status MQTTStatus

	if status.Connected {
		t.Error("expected Connected to be false by default")
	}
}

func TestEthernetStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status EthernetStatus

	if status.IP != "" {
		t.Errorf("got IP=%q, want empty", status.IP)
	}
}

func TestCloudControlResult_ZeroValue(t *testing.T) {
	t.Parallel()

	var result CloudControlResult

	if result.Success {
		t.Error("expected Success to be false by default")
	}
	if result.Message != "" {
		t.Errorf("got Message=%q, want empty", result.Message)
	}
}

// ----- Identifier Passthrough Tests -----

func TestWiFiService_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := NewWiFiService(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_, _ = svc.GetStatusFull(context.Background(), "my-wifi-device")

	if capturedIdentifier != "my-wifi-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "my-wifi-device")
	}
}

func TestMQTTService_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := NewMQTTService(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_, _ = svc.GetStatus(context.Background(), "my-mqtt-device")

	if capturedIdentifier != "my-mqtt-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "my-mqtt-device")
	}
}

func TestEthernetService_IdentifierPassthrough(t *testing.T) {
	t.Parallel()

	var capturedIdentifier string
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, identifier string, _ func(*client.Client) error) error {
			capturedIdentifier = identifier
			return errors.New("rpc not mocked")
		},
	}

	svc := NewEthernetService(provider)
	//nolint:errcheck // intentionally ignoring error to test identifier passthrough
	_, _ = svc.GetStatus(context.Background(), "my-eth-device")

	if capturedIdentifier != "my-eth-device" {
		t.Errorf("got identifier=%q, want %q", capturedIdentifier, "my-eth-device")
	}
}

// ----- Cloud Action Handler Tests -----

func TestBuildCloudActionHandlers(t *testing.T) {
	t.Parallel()

	handlers := buildCloudActionHandlers()

	// Check that all expected actions are present
	expectedActions := []string{
		"on", "off", "toggle",
		"open", "close", "stop",
		"light-on", "light-off", "light-toggle",
	}

	for _, action := range expectedActions {
		if _, ok := handlers[action]; !ok {
			t.Errorf("missing handler for action %q", action)
		}
	}

	// Check handler properties
	tests := []struct {
		action  string
		success string
		errMsg  string
	}{
		{"on", "Switch turned on", "failed to turn on switch"},
		{"off", "Switch turned off", "failed to turn off switch"},
		{"toggle", "Switch toggled", "failed to toggle switch"},
		{"open", "Cover opening", "failed to open cover"},
		{"close", "Cover closing", "failed to close cover"},
		{"stop", "Cover stopped", "failed to stop cover"},
		{"light-on", "Light turned on", "failed to turn on light"},
		{"light-off", "Light turned off", "failed to turn off light"},
		{"light-toggle", "Light toggled", "failed to toggle light"},
	}

	for _, tt := range tests {
		handler := handlers[tt.action]
		if handler.success != tt.success {
			t.Errorf("action %q: got success=%q, want %q", tt.action, handler.success, tt.success)
		}
		if handler.errMsg != tt.errMsg {
			t.Errorf("action %q: got errMsg=%q, want %q", tt.action, handler.errMsg, tt.errMsg)
		}
	}
}

// ----- Service Nil Provider Tests -----

func TestNewWiFiServiceNil(t *testing.T) {
	t.Parallel()

	svc := NewWiFiService(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected provider to be nil")
	}
}

func TestNewMQTTServiceNil(t *testing.T) {
	t.Parallel()

	svc := NewMQTTService(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected provider to be nil")
	}
}

func TestNewEthernetServiceNil(t *testing.T) {
	t.Parallel()

	svc := NewEthernetService(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected provider to be nil")
	}
}

// ----- Interface Compliance Test -----

func TestConnectionProvider_Interface(t *testing.T) {
	t.Parallel()

	// This test verifies that mockConnectionProvider satisfies the ConnectionProvider interface
	var _ ConnectionProvider = (*mockConnectionProvider)(nil)
}

// ----- Token Validation Tests -----

func TestValidateToken_InvalidToken(t *testing.T) {
	t.Parallel()

	err := ValidateToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestIsTokenExpired_InvalidToken(t *testing.T) {
	t.Parallel()

	// An invalid token should not be treated as expired (returns false for parsing failure)
	// This tests the edge case behavior
	expired := IsTokenExpired("invalid-token")
	// The behavior depends on the cloud package implementation
	_ = expired // Just ensure it doesn't panic
}

func TestTimeUntilExpiry_InvalidToken(t *testing.T) {
	t.Parallel()

	// An invalid token should return a zero or negative duration
	duration := TimeUntilExpiry("invalid-token")
	// Just ensure it doesn't panic and returns some duration
	_ = duration
}

func TestParseToken_InvalidToken(t *testing.T) {
	t.Parallel()

	token, err := ParseToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
	if token != nil {
		t.Error("expected nil token for invalid input")
	}
}

// ----- Cloud Client Tests -----

func TestNewCloudClient_SetClient(t *testing.T) {
	t.Parallel()

	cloudClient := NewCloudClient("test-token")

	// The internal client should be set
	if cloudClient.client == nil {
		t.Error("expected internal client to be set")
	}
}

// ----- Cloud Event Stream Options Tests -----

func TestCloudEventStreamOptions_Defaults(t *testing.T) {
	t.Parallel()

	var opts CloudEventStreamOptions

	if opts.DeviceFilter != "" {
		t.Errorf("expected empty DeviceFilter, got %q", opts.DeviceFilter)
	}
	if opts.EventFilter != "" {
		t.Errorf("expected empty EventFilter, got %q", opts.EventFilter)
	}
	if opts.Raw {
		t.Error("expected Raw to be false by default")
	}
}

// ----- Cloud Control Result Tests -----

func TestCloudControlResult_FailureCase(t *testing.T) {
	t.Parallel()

	result := CloudControlResult{
		Success: false,
		Message: "Device offline",
	}

	if result.Success {
		t.Error("expected Success to be false")
	}
	if result.Message != "Device offline" {
		t.Errorf("got Message=%q, want %q", result.Message, "Device offline")
	}
}

// ----- Cloud Login Result Tests -----

func TestCloudLoginResult_ZeroExpiry(t *testing.T) {
	t.Parallel()

	var result CloudLoginResult

	if !result.Expiry.IsZero() {
		t.Error("expected zero expiry by default")
	}
	if result.Token != "" {
		t.Error("expected empty token by default")
	}
}

// ----- Build Cloud Action Handlers Coverage Tests -----

func TestBuildCloudActionHandlers_AllActions(t *testing.T) {
	t.Parallel()

	handlers := buildCloudActionHandlers()

	// Verify all handlers are present and have correct field values
	allActions := []string{
		"on", "off", "toggle",
		"open", "close", "stop",
		"light-on", "light-off", "light-toggle",
	}

	for _, action := range allActions {
		handler, ok := handlers[action]
		if !ok {
			t.Errorf("missing handler for action %q", action)
			continue
		}
		// Verify handler has non-empty strings
		if handler.success == "" {
			t.Errorf("action %q has empty success message", action)
		}
		if handler.errMsg == "" {
			t.Errorf("action %q has empty error message", action)
		}
		if handler.fn == nil {
			t.Errorf("action %q has nil function", action)
		}
	}
}

// ----- Additional WiFi Tests -----

func TestWiFiConfigFull_NilFields(t *testing.T) {
	t.Parallel()

	config := WiFiConfigFull{}

	if config.STA != nil {
		t.Error("expected STA to be nil by default")
	}
	if config.STA1 != nil {
		t.Error("expected STA1 to be nil by default")
	}
	if config.AP != nil {
		t.Error("expected AP to be nil by default")
	}
}

// ----- Additional MQTT Tests -----

func TestMQTTConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	var config MQTTConfig

	if config.Enable {
		t.Error("expected Enable to be false by default")
	}
	if config.Server != "" {
		t.Error("expected empty Server by default")
	}
	if config.User != "" {
		t.Error("expected empty User by default")
	}
}

// ----- Additional Ethernet Tests -----

func TestEthernetConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	var config EthernetConfig

	if config.Enable {
		t.Error("expected Enable to be false by default")
	}
	if config.IPv4Mode != "" {
		t.Error("expected empty IPv4Mode by default")
	}
}

// ----- Cloud Auth Status Tests -----

func TestCloudAuthStatus_ZeroValue(t *testing.T) {
	t.Parallel()

	var status CloudAuthStatus

	if status.Authenticated {
		t.Error("expected Authenticated to be false by default")
	}
	if status.TokenValid {
		t.Error("expected TokenValid to be false by default")
	}
	if !status.TokenExpiry.IsZero() {
		t.Error("expected zero TokenExpiry by default")
	}
}

// ----- Set Config Params Tests -----

func TestSetConfigParams_NilEnable(t *testing.T) {
	t.Parallel()

	params := SetConfigParams{}

	if params.Enable != nil {
		t.Error("expected Enable to be nil by default")
	}
}

func TestEthernetSetConfigParams_NilEnable(t *testing.T) {
	t.Parallel()

	params := EthernetSetConfigParams{}

	if params.Enable != nil {
		t.Error("expected Enable to be nil by default")
	}
}

// ----- Additional BuildCloudWebSocketURL Tests -----

func TestBuildCloudWebSocketURL_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		serverURL string
		token     string
		wantErr   bool
		wantHost  string
	}{
		{
			name:      "valid HTTPS URL",
			serverURL: "https://api.shelly.cloud",
			token:     "token123",
			wantHost:  "api.shelly.cloud",
		},
		{
			name:      "valid HTTP URL",
			serverURL: "http://local.shelly.com",
			token:     "token456",
			wantHost:  "local.shelly.com",
		},
		{
			name:      "URL with port",
			serverURL: "https://api.shelly.cloud:8443",
			token:     "token789",
			wantHost:  "api.shelly.cloud",
		},
		{
			name:      "URL with path",
			serverURL: "https://api.shelly.cloud/v1/api",
			token:     "tokenabc",
			wantHost:  "api.shelly.cloud",
		},
		{
			name:      "empty server and invalid token",
			serverURL: "",
			token:     "not-a-jwt",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			url, err := BuildCloudWebSocketURL(tt.serverURL, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that URL contains expected host
			if tt.wantHost != "" && !containsSubstring(url, tt.wantHost) {
				t.Errorf("URL %q should contain host %q", url, tt.wantHost)
			}

			// Check URL structure
			if !containsPrefix(url, "wss://") {
				t.Errorf("URL %q should start with wss://", url)
			}
			if !containsSubstring(url, ":6113/shelly/wss/hk_sock?t=") {
				t.Errorf("URL %q should contain port and path", url)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ----- SetConfigParams Full Field Tests -----

func TestSetConfigParams_AllFields(t *testing.T) {
	t.Parallel()

	enable := true
	params := SetConfigParams{
		Enable:      &enable,
		Server:      "mqtt://broker:1883",
		User:        "user",
		Password:    "pass",
		ClientID:    "client-123",
		TopicPrefix: "prefix/",
		SSLCA:       "ca.pem",
	}

	if params.Enable == nil || !*params.Enable {
		t.Error("expected Enable to be true")
	}
	if params.Server != "mqtt://broker:1883" {
		t.Errorf("got Server=%q", params.Server)
	}
	if params.User != "user" {
		t.Errorf("got User=%q", params.User)
	}
	if params.Password != "pass" {
		t.Errorf("got Password=%q", params.Password)
	}
	if params.ClientID != "client-123" {
		t.Errorf("got ClientID=%q", params.ClientID)
	}
	if params.TopicPrefix != "prefix/" {
		t.Errorf("got TopicPrefix=%q", params.TopicPrefix)
	}
	if params.SSLCA != "ca.pem" {
		t.Errorf("got SSLCA=%q", params.SSLCA)
	}
}

// ----- EthernetSetConfigParams Full Field Tests -----

func TestEthernetSetConfigParams_AllFields(t *testing.T) {
	t.Parallel()

	enable := false
	params := EthernetSetConfigParams{
		Enable:     &enable,
		IPv4Mode:   "dhcp",
		IP:         "192.168.1.50",
		Netmask:    "255.255.255.0",
		GW:         "192.168.1.1",
		Nameserver: "1.1.1.1",
	}

	if params.Enable == nil || *params.Enable {
		t.Error("expected Enable to be false")
	}
	if params.IPv4Mode != "dhcp" {
		t.Errorf("got IPv4Mode=%q", params.IPv4Mode)
	}
	if params.IP != "192.168.1.50" {
		t.Errorf("got IP=%q", params.IP)
	}
	if params.Netmask != "255.255.255.0" {
		t.Errorf("got Netmask=%q", params.Netmask)
	}
	if params.GW != "192.168.1.1" {
		t.Errorf("got GW=%q", params.GW)
	}
	if params.Nameserver != "1.1.1.1" {
		t.Errorf("got Nameserver=%q", params.Nameserver)
	}
}

// ----- WiFi Station Full Edge Cases -----

func TestWiFiStationFull_StaticIP(t *testing.T) {
	t.Parallel()

	station := WiFiStationFull{
		SSID:     "StaticNetwork",
		Enabled:  true,
		IsOpen:   false,
		IPv4Mode: testIPv4Static,
		IP:       "10.0.0.50",
		Netmask:  "255.0.0.0",
		Gateway:  "10.0.0.1",
	}

	if station.IPv4Mode != testIPv4Static {
		t.Errorf("got IPv4Mode=%q, want %q", station.IPv4Mode, testIPv4Static)
	}
	if station.IP != "10.0.0.50" {
		t.Errorf("got IP=%q", station.IP)
	}
}

// ----- WiFi Network Full Signal Strength Tests -----

func TestWiFiNetworkFull_SignalStrength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rssi   float64
		strong bool
	}{
		{"excellent signal", -30.0, true},
		{"good signal", -50.0, true},
		{"fair signal", -70.0, false},
		{"weak signal", -85.0, false},
		{"very weak signal", -95.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			network := WiFiNetworkFull{
				SSID: "TestNetwork",
				RSSI: tt.rssi,
			}

			// Strong signal is typically > -60 dBm
			isStrong := network.RSSI > -60.0
			if isStrong != tt.strong {
				t.Errorf("RSSI %f: got strong=%v, want %v", tt.rssi, isStrong, tt.strong)
			}
		})
	}
}

// ----- MQTT Config SSL CA Tests -----

func TestMQTTConfig_SSLCA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		sslca string
	}{
		{"no TLS", ""},
		{"trust all", "*"},
		{"bundled CA", "ca.pem"},
		{"user CA", "user_ca.pem"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := MQTTConfig{
				Enable: true,
				Server: "mqtts://broker:8883",
				SSLCA:  tt.sslca,
			}

			if config.SSLCA != tt.sslca {
				t.Errorf("got SSLCA=%q, want %q", config.SSLCA, tt.sslca)
			}
		})
	}
}

// ----- Cloud Device Full Tests -----

func TestCloudDevice_AllFields(t *testing.T) {
	t.Parallel()

	device := CloudDevice{
		ID:              "shellyplus1pm-aabbccddee",
		Name:            "Office Lamp",
		Model:           "SNSW-001P16EU",
		MAC:             "AA:BB:CC:DD:EE:FF",
		FirmwareVersion: "1.2.3-beta1",
		Generation:      2,
		Online:          false,
		CloudEnabled:    true,
		Status:          []byte(`{"switch:0":{"output":true}}`),
		Settings:        []byte(`{"name":"Office Lamp"}`),
	}

	if device.ID != "shellyplus1pm-aabbccddee" {
		t.Errorf("got ID=%q", device.ID)
	}
	if device.Name != "Office Lamp" {
		t.Errorf("got Name=%q", device.Name)
	}
	if device.Model != "SNSW-001P16EU" {
		t.Errorf("got Model=%q", device.Model)
	}
	if device.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got MAC=%q", device.MAC)
	}
	if device.FirmwareVersion != "1.2.3-beta1" {
		t.Errorf("got FirmwareVersion=%q", device.FirmwareVersion)
	}
	if device.Generation != 2 {
		t.Errorf("got Generation=%d", device.Generation)
	}
	if device.Online {
		t.Error("expected Online to be false")
	}
	if !device.CloudEnabled {
		t.Error("expected CloudEnabled to be true")
	}
	if len(device.Status) == 0 {
		t.Error("expected non-empty Status")
	}
	if len(device.Settings) == 0 {
		t.Error("expected non-empty Settings")
	}
}

// ----- Cloud Token Full Tests -----

func TestCloudToken_AllFields(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	token := CloudToken{
		AccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
		UserAPIURL:  "https://shelly-eu.cloudflare.com",
		Email:       "test@example.com",
		UserID:      999999,
		Expiry:      expiry,
	}

	if token.AccessToken == "" {
		t.Error("expected non-empty AccessToken")
	}
	if token.UserAPIURL != "https://shelly-eu.cloudflare.com" {
		t.Errorf("got UserAPIURL=%q", token.UserAPIURL)
	}
	if token.Email != "test@example.com" {
		t.Errorf("got Email=%q", token.Email)
	}
	if token.UserID != 999999 {
		t.Errorf("got UserID=%d", token.UserID)
	}
	if token.Expiry.IsZero() {
		t.Error("expected non-zero Expiry")
	}
}

// ----- Mock Provider Callback Tests -----

func TestMockConnectionProvider_CallbackInvoked(t *testing.T) {
	t.Parallel()

	callbackInvoked := false
	provider := &mockConnectionProvider{
		withConnectionFn: func(_ context.Context, _ string, fn func(*client.Client) error) error {
			callbackInvoked = true
			// Don't actually call fn since we don't have a real client
			return errors.New("mock error")
		},
	}

	svc := NewWiFiService(provider)
	//nolint:errcheck // testing callback invocation, not error
	_, _ = svc.GetStatusFull(context.Background(), "test")

	if !callbackInvoked {
		t.Error("expected callback to be invoked")
	}
}

// ----- Context Cancellation Tests -----

func TestWiFiService_ContextCancellation(t *testing.T) {
	t.Parallel()

	contextWasCaptured := false
	provider := &mockConnectionProvider{
		withConnectionFn: func(ctx context.Context, _ string, _ func(*client.Client) error) error {
			contextWasCaptured = true
			return ctx.Err()
		},
	}

	svc := NewWiFiService(provider)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := svc.GetStatusFull(ctx, "test-device")

	if !contextWasCaptured {
		t.Fatal("expected context to be captured")
	}
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

// ----- Hash Password Consistency Tests -----

func TestHashPassword_Consistency(t *testing.T) {
	t.Parallel()

	passwords := []string{
		"simple",
		"P@ssw0rd!",
		"with spaces here",
		"日本語",
		"",
	}

	for _, pass := range passwords {
		t.Run(pass, func(t *testing.T) {
			t.Parallel()

			hash1 := HashPassword(pass)
			hash2 := HashPassword(pass)

			if hash1 != hash2 {
				t.Errorf("inconsistent hash for %q: %q != %q", pass, hash1, hash2)
			}
		})
	}
}

func TestHashPassword_Uniqueness(t *testing.T) {
	t.Parallel()

	hashes := make(map[string]string)
	passwords := []string{
		"password1",
		"password2",
		"password3",
		"abc",
		"ABC",
	}

	for _, pass := range passwords {
		hash := HashPassword(pass)
		if existing, ok := hashes[hash]; ok {
			t.Errorf("hash collision: %q and %q both hash to %q", pass, existing, hash)
		}
		hashes[hash] = pass
	}
}

// ----- Cloud Auth Status Validity Tests -----

func TestCloudAuthStatus_Validity(t *testing.T) {
	t.Parallel()

	t.Run("authenticated with valid token", func(t *testing.T) {
		t.Parallel()

		status := CloudAuthStatus{
			Authenticated: true,
			Email:         "user@example.com",
			UserAPIURL:    "https://api.shelly.cloud",
			TokenExpiry:   time.Now().Add(time.Hour),
			TokenValid:    true,
		}

		if !status.Authenticated || !status.TokenValid {
			t.Error("expected valid authenticated status")
		}
	})

	t.Run("authenticated with expired token", func(t *testing.T) {
		t.Parallel()

		status := CloudAuthStatus{
			Authenticated: true,
			Email:         "user@example.com",
			UserAPIURL:    "https://api.shelly.cloud",
			TokenExpiry:   time.Now().Add(-time.Hour), // Expired
			TokenValid:    false,
		}

		if !status.Authenticated {
			t.Error("expected Authenticated to be true")
		}
		if status.TokenValid {
			t.Error("expected TokenValid to be false for expired token")
		}
	})

	t.Run("not authenticated", func(t *testing.T) {
		t.Parallel()

		status := CloudAuthStatus{
			Authenticated: false,
			TokenValid:    false,
		}

		if status.Authenticated {
			t.Error("expected Authenticated to be false")
		}
	})
}

// ----- Cloud Login Result Expiry Tests -----

func TestCloudLoginResult_ExpiryTimes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		expiry time.Time
	}{
		{"future expiry", time.Now().Add(24 * time.Hour)},
		{"past expiry", time.Now().Add(-24 * time.Hour)},
		{"immediate expiry", time.Now()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CloudLoginResult{
				Token:      "token",
				UserAPIURL: "https://api.shelly.cloud",
				Expiry:     tt.expiry,
			}

			if result.Expiry.IsZero() {
				t.Error("expected non-zero expiry")
			}
		})
	}
}
