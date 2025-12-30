package model

import (
	"testing"
	"time"
)

// TestAuditResult tests AuditResult struct fields.
func TestAuditResult(t *testing.T) {
	t.Parallel()

	result := AuditResult{
		Device:    "kitchen-switch",
		Address:   "192.168.1.100",
		Issues:    []string{"no auth"},
		Warnings:  []string{"firmware outdated"},
		InfoItems: []string{"Gen2 device"},
		Reachable: true,
		AuthStatus: &AuthAudit{
			AuthEnabled: false,
		},
		CloudAudit: &CloudAudit{
			Connected: true,
		},
		FWAudit: &FirmwareAudit{
			Current:   "1.0.0",
			Available: "1.1.0",
			HasUpdate: true,
		},
	}

	if result.Device != "kitchen-switch" {
		t.Errorf("Device = %q, want %q", result.Device, "kitchen-switch")
	}
	if result.Address != "192.168.1.100" { //nolint:goconst // test data in different file
		t.Errorf("Address = %q, want %q", result.Address, "192.168.1.100")
	}
	if len(result.Issues) != 1 {
		t.Errorf("Issues len = %d, want 1", len(result.Issues))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Warnings len = %d, want 1", len(result.Warnings))
	}
	if len(result.InfoItems) != 1 {
		t.Errorf("InfoItems len = %d, want 1", len(result.InfoItems))
	}
	if !result.Reachable {
		t.Error("Reachable = false, want true")
	}
	if result.AuthStatus == nil || result.AuthStatus.AuthEnabled {
		t.Error("AuthStatus should have AuthEnabled = false")
	}
	if result.CloudAudit == nil || !result.CloudAudit.Connected {
		t.Error("CloudAudit should have Connected = true")
	}
	if result.FWAudit == nil || !result.FWAudit.HasUpdate {
		t.Error("FWAudit should have HasUpdate = true")
	}
}

// TestAuthAudit tests AuthAudit struct.
func TestAuthAudit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authEnabled bool
	}{
		{"auth enabled", true},
		{"auth disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			audit := AuthAudit{AuthEnabled: tt.authEnabled}
			if audit.AuthEnabled != tt.authEnabled {
				t.Errorf("AuthEnabled = %v, want %v", audit.AuthEnabled, tt.authEnabled)
			}
		})
	}
}

// TestCloudAudit tests CloudAudit struct.
func TestCloudAudit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		connected bool
	}{
		{"connected", true},
		{"disconnected", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			audit := CloudAudit{Connected: tt.connected}
			if audit.Connected != tt.connected {
				t.Errorf("Connected = %v, want %v", audit.Connected, tt.connected)
			}
		})
	}
}

// TestFirmwareAudit tests FirmwareAudit struct.
func TestFirmwareAudit(t *testing.T) {
	t.Parallel()

	audit := FirmwareAudit{
		Current:   "1.0.0",
		Available: "1.1.0",
		HasUpdate: true,
	}

	if audit.Current != "1.0.0" {
		t.Errorf("Current = %q, want %q", audit.Current, "1.0.0")
	}
	if audit.Available != "1.1.0" {
		t.Errorf("Available = %q, want %q", audit.Available, "1.1.0")
	}
	if !audit.HasUpdate {
		t.Error("HasUpdate = false, want true")
	}
}

// TestBatchRPCResult tests BatchRPCResult struct.
func TestBatchRPCResult(t *testing.T) {
	t.Parallel()

	// Test success case
	result := BatchRPCResult{
		Device:   "device-1",
		Response: map[string]any{"output": true},
	}

	if result.Device != "device-1" {
		t.Errorf("Device = %q, want %q", result.Device, "device-1")
	}
	if result.Response == nil {
		t.Error("Response should not be nil")
	}
	if result.Error != "" {
		t.Errorf("Error = %q, want empty", result.Error)
	}

	// Test error case
	errResult := BatchRPCResult{
		Device: "device-2",
		Error:  "connection failed",
	}

	if errResult.Error != "connection failed" {
		t.Errorf("Error = %q, want %q", errResult.Error, "connection failed")
	}
}

// TestBenchmarkResult tests BenchmarkResult struct.
func TestBenchmarkResult(t *testing.T) {
	t.Parallel()

	now := time.Now()
	result := BenchmarkResult{
		Device:     "test-device",
		Iterations: 100,
		PingLatency: LatencyStats{
			Min:    5 * time.Millisecond,
			Max:    50 * time.Millisecond,
			Avg:    15 * time.Millisecond,
			P50:    12 * time.Millisecond,
			P95:    35 * time.Millisecond,
			P99:    45 * time.Millisecond,
			Errors: 0,
		},
		RPCLatency: LatencyStats{
			Min:    10 * time.Millisecond,
			Max:    100 * time.Millisecond,
			Avg:    30 * time.Millisecond,
			P50:    25 * time.Millisecond,
			P95:    70 * time.Millisecond,
			P99:    90 * time.Millisecond,
			Errors: 2,
		},
		Summary:   "100 iterations completed",
		Timestamp: now,
	}

	if result.Device != "test-device" {
		t.Errorf("Device = %q, want %q", result.Device, "test-device")
	}
	if result.Iterations != 100 {
		t.Errorf("Iterations = %d, want 100", result.Iterations)
	}
	if result.PingLatency.Errors != 0 {
		t.Errorf("PingLatency.Errors = %d, want 0", result.PingLatency.Errors)
	}
	if result.RPCLatency.Errors != 2 {
		t.Errorf("RPCLatency.Errors = %d, want 2", result.RPCLatency.Errors)
	}
}

// TestLatencyStats tests LatencyStats struct.
func TestLatencyStats(t *testing.T) {
	t.Parallel()

	stats := LatencyStats{
		Min:    1 * time.Millisecond,
		Max:    100 * time.Millisecond,
		Avg:    25 * time.Millisecond,
		P50:    20 * time.Millisecond,
		P95:    80 * time.Millisecond,
		P99:    95 * time.Millisecond,
		Errors: 5,
	}

	if stats.Min != 1*time.Millisecond {
		t.Errorf("Min = %v, want 1ms", stats.Min)
	}
	if stats.Max != 100*time.Millisecond {
		t.Errorf("Max = %v, want 100ms", stats.Max)
	}
	if stats.Avg != 25*time.Millisecond {
		t.Errorf("Avg = %v, want 25ms", stats.Avg)
	}
	if stats.P50 != 20*time.Millisecond {
		t.Errorf("P50 = %v, want 20ms", stats.P50)
	}
	if stats.P95 != 80*time.Millisecond {
		t.Errorf("P95 = %v, want 80ms", stats.P95)
	}
	if stats.P99 != 95*time.Millisecond {
		t.Errorf("P99 = %v, want 95ms", stats.P99)
	}
	if stats.Errors != 5 {
		t.Errorf("Errors = %d, want 5", stats.Errors)
	}
}

// TestBTHomeDeviceInfo tests BTHomeDeviceInfo struct.
func TestBTHomeDeviceInfo(t *testing.T) {
	t.Parallel()

	rssi := -55
	battery := 85
	info := BTHomeDeviceInfo{
		ID:         0,
		Name:       "Sensor1",
		Addr:       "AA:BB:CC:DD:EE:FF",
		RSSI:       &rssi,
		Battery:    &battery,
		LastUpdate: 1700000000.0,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if info.Name != "Sensor1" {
		t.Errorf("Name = %q, want %q", info.Name, "Sensor1")
	}
	if info.Addr != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Addr = %q, want %q", info.Addr, "AA:BB:CC:DD:EE:FF")
	}
	if info.RSSI == nil || *info.RSSI != -55 {
		t.Errorf("RSSI = %v, want -55", info.RSSI)
	}
	if info.Battery == nil || *info.Battery != 85 {
		t.Errorf("Battery = %v, want 85", info.Battery)
	}
}

// TestBTHomeComponentStatus tests BTHomeComponentStatus struct.
func TestBTHomeComponentStatus(t *testing.T) {
	t.Parallel()

	status := BTHomeComponentStatus{
		Discovery: &BTHomeDiscoveryStatus{
			StartedAt: 1700000000.0,
			Duration:  30,
		},
		Errors: []string{"scan timeout"},
	}

	if status.Discovery == nil {
		t.Fatal("Discovery should not be nil")
	}
	if status.Discovery.Duration != 30 {
		t.Errorf("Discovery.Duration = %d, want 30", status.Discovery.Duration)
	}
	if len(status.Errors) != 1 {
		t.Errorf("Errors len = %d, want 1", len(status.Errors))
	}
}

// TestBTHomeDeviceStatus tests BTHomeDeviceStatus struct.
func TestBTHomeDeviceStatus(t *testing.T) {
	t.Parallel()

	rssi := -60
	packetID := 42
	component := "temperature:0"
	status := BTHomeDeviceStatus{
		ID:           0,
		Name:         "TempSensor",
		Addr:         "11:22:33:44:55:66",
		RSSI:         &rssi,
		PacketID:     &packetID,
		LastUpdateTS: 1700000000.0,
		KnownObjects: []BTHomeKnownObj{
			{
				ObjID:     0x02,
				Idx:       0,
				Component: &component,
			},
		},
	}

	if status.Name != "TempSensor" {
		t.Errorf("Name = %q, want %q", status.Name, "TempSensor")
	}
	if len(status.KnownObjects) != 1 {
		t.Fatalf("KnownObjects len = %d, want 1", len(status.KnownObjects))
	}
	if status.KnownObjects[0].ObjID != 0x02 {
		t.Errorf("KnownObjects[0].ObjID = %d, want 2", status.KnownObjects[0].ObjID)
	}
}

// TestBTHomeKnownObj tests BTHomeKnownObj struct.
func TestBTHomeKnownObj(t *testing.T) {
	t.Parallel()

	component := "humidity:0"
	obj := BTHomeKnownObj{
		ObjID:     0x03,
		Idx:       0,
		Component: &component,
	}

	if obj.ObjID != 0x03 {
		t.Errorf("ObjID = %d, want 3", obj.ObjID)
	}
	if obj.Idx != 0 {
		t.Errorf("Idx = %d, want 0", obj.Idx)
	}
	if obj.Component == nil || *obj.Component != "humidity:0" {
		t.Errorf("Component = %v, want %q", obj.Component, "humidity:0")
	}
}

// TestCertInstallData tests CertInstallData struct.
func TestCertInstallData(t *testing.T) {
	t.Parallel()

	data := CertInstallData{
		CAData:   []byte("-----BEGIN CERTIFICATE-----\nCA\n-----END CERTIFICATE-----"),
		CertData: []byte("-----BEGIN CERTIFICATE-----\nCERT\n-----END CERTIFICATE-----"),
		KeyData:  []byte("-----BEGIN PRIVATE KEY-----\nKEY\n-----END PRIVATE KEY-----"),
	}

	if len(data.CAData) == 0 {
		t.Error("CAData should not be empty")
	}
	if len(data.CertData) == 0 {
		t.Error("CertData should not be empty")
	}
	if len(data.KeyData) == 0 {
		t.Error("KeyData should not be empty")
	}
}

// TestDeviceListItem tests DeviceListItem struct.
func TestDeviceListItem(t *testing.T) {
	t.Parallel()

	item := DeviceListItem{
		Name:             "Kitchen Switch",
		Address:          "192.168.1.100",
		Platform:         PlatformShelly,
		Model:            "Plus 1PM",
		Type:             "SHSW-PM",
		Generation:       2,
		Auth:             true,
		CurrentVersion:   "1.0.0",
		AvailableVersion: "1.1.0",
		HasUpdate:        true,
	}

	if item.Name != "Kitchen Switch" {
		t.Errorf("Name = %q, want %q", item.Name, "Kitchen Switch")
	}
	if item.Platform != PlatformShelly {
		t.Errorf("Platform = %q, want %q", item.Platform, PlatformShelly)
	}
	if item.Generation != 2 {
		t.Errorf("Generation = %d, want 2", item.Generation)
	}
	if !item.Auth {
		t.Error("Auth = false, want true")
	}
	if !item.HasUpdate {
		t.Error("HasUpdate = false, want true")
	}
}

// TestGroupInfo tests GroupInfo struct.
func TestGroupInfo(t *testing.T) {
	t.Parallel()

	info := GroupInfo{
		Name:        "Living Room",
		DeviceCount: 3,
		Devices:     []string{"switch1", "switch2", "light1"},
	}

	if info.Name != "Living Room" {
		t.Errorf("Name = %q, want %q", info.Name, "Living Room")
	}
	if info.DeviceCount != 3 {
		t.Errorf("DeviceCount = %d, want 3", info.DeviceCount)
	}
	if len(info.Devices) != 3 {
		t.Errorf("Devices len = %d, want 3", len(info.Devices))
	}
}

// TestThermostatInfo tests ThermostatInfo struct.
func TestThermostatInfo(t *testing.T) {
	t.Parallel()

	info := ThermostatInfo{
		ID:      0,
		Enabled: true,
		Mode:    "heat",
		TargetC: 22.5,
	}

	if info.ID != 0 {
		t.Errorf("ID = %d, want 0", info.ID)
	}
	if !info.Enabled {
		t.Error("Enabled = false, want true")
	}
	if info.Mode != "heat" {
		t.Errorf("Mode = %q, want %q", info.Mode, "heat")
	}
	if info.TargetC != 22.5 {
		t.Errorf("TargetC = %f, want 22.5", info.TargetC)
	}
}

// TestZigbeeDevice tests ZigbeeDevice struct.
func TestZigbeeDevice(t *testing.T) {
	t.Parallel()

	device := ZigbeeDevice{
		Name:         "Zigbee Hub",
		Address:      "192.168.1.50",
		Model:        "Plus Zigbee",
		Enabled:      true,
		NetworkState: "running",
		EUI64:        "00:11:22:33:44:55:66:77",
	}

	if device.Name != "Zigbee Hub" {
		t.Errorf("Name = %q, want %q", device.Name, "Zigbee Hub")
	}
	if !device.Enabled {
		t.Error("Enabled = false, want true")
	}
	if device.NetworkState != "running" {
		t.Errorf("NetworkState = %q, want %q", device.NetworkState, "running")
	}
}

// TestZigbeeStatus tests ZigbeeStatus struct.
func TestZigbeeStatus(t *testing.T) {
	t.Parallel()

	status := ZigbeeStatus{
		Enabled:          true,
		NetworkState:     "joined",
		EUI64:            "00:11:22:33:44:55:66:77",
		PANID:            0x1234,
		Channel:          15,
		CoordinatorEUI64: "AA:BB:CC:DD:EE:FF:00:11",
	}

	if !status.Enabled {
		t.Error("Enabled = false, want true")
	}
	if status.PANID != 0x1234 {
		t.Errorf("PANID = %d, want %d", status.PANID, 0x1234)
	}
	if status.Channel != 15 {
		t.Errorf("Channel = %d, want 15", status.Channel)
	}
}

// TestLoRaFullStatus tests LoRaFullStatus struct.
func TestLoRaFullStatus(t *testing.T) {
	t.Parallel()

	status := LoRaFullStatus{
		Config: &LoRaConfig{
			ID:   0,
			Freq: 868000000,
			BW:   125,
			DR:   5,
			TxP:  14,
		},
		Status: &LoRaStatus{
			ID:   0,
			RSSI: -75,
			SNR:  7.5,
		},
	}

	if status.Config == nil {
		t.Fatal("Config should not be nil")
	}
	if status.Config.Freq != 868000000 {
		t.Errorf("Config.Freq = %d, want 868000000", status.Config.Freq)
	}
	if status.Status == nil {
		t.Fatal("Status should not be nil")
	}
	if status.Status.RSSI != -75 {
		t.Errorf("Status.RSSI = %d, want -75", status.Status.RSSI)
	}
	if status.Status.SNR != 7.5 {
		t.Errorf("Status.SNR = %f, want 7.5", status.Status.SNR)
	}
}

// TestLoRaConfig tests LoRaConfig struct.
func TestLoRaConfig(t *testing.T) {
	t.Parallel()

	cfg := LoRaConfig{
		ID:   0,
		Freq: 915000000,
		BW:   250,
		DR:   7,
		TxP:  20,
	}

	if cfg.Freq != 915000000 {
		t.Errorf("Freq = %d, want 915000000", cfg.Freq)
	}
	if cfg.BW != 250 {
		t.Errorf("BW = %d, want 250", cfg.BW)
	}
	if cfg.DR != 7 {
		t.Errorf("DR = %d, want 7", cfg.DR)
	}
	if cfg.TxP != 20 {
		t.Errorf("TxP = %d, want 20", cfg.TxP)
	}
}

// TestLoRaStatus tests LoRaStatus struct.
func TestLoRaStatus(t *testing.T) {
	t.Parallel()

	status := LoRaStatus{
		ID:   0,
		RSSI: -80,
		SNR:  5.0,
	}

	if status.ID != 0 {
		t.Errorf("ID = %d, want 0", status.ID)
	}
	if status.RSSI != -80 {
		t.Errorf("RSSI = %d, want -80", status.RSSI)
	}
	if status.SNR != 5.0 {
		t.Errorf("SNR = %f, want 5.0", status.SNR)
	}
}

// TestMatterStatus tests MatterStatus struct.
func TestMatterStatus(t *testing.T) {
	t.Parallel()

	status := MatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   2,
	}

	if !status.Enabled {
		t.Error("Enabled = false, want true")
	}
	if !status.Commissionable {
		t.Error("Commissionable = false, want true")
	}
	if status.FabricsCount != 2 {
		t.Errorf("FabricsCount = %d, want 2", status.FabricsCount)
	}
}

// TestCommissioningInfo tests CommissioningInfo struct.
func TestCommissioningInfo(t *testing.T) {
	t.Parallel()

	info := CommissioningInfo{
		ManualCode:    "12345678901",
		QRCode:        "MT:Y.K900.ABC12",
		Discriminator: 1234,
		SetupPINCode:  12345678,
		Available:     true,
	}

	if info.ManualCode != "12345678901" {
		t.Errorf("ManualCode = %q, want %q", info.ManualCode, "12345678901")
	}
	if info.QRCode != "MT:Y.K900.ABC12" {
		t.Errorf("QRCode = %q, want %q", info.QRCode, "MT:Y.K900.ABC12")
	}
	if info.Discriminator != 1234 {
		t.Errorf("Discriminator = %d, want 1234", info.Discriminator)
	}
	if info.SetupPINCode != 12345678 {
		t.Errorf("SetupPINCode = %d, want 12345678", info.SetupPINCode)
	}
	if !info.Available {
		t.Error("Available = false, want true")
	}
}

// TestDeviceQRInfo tests DeviceQRInfo struct.
func TestDeviceQRInfo(t *testing.T) {
	t.Parallel()

	info := DeviceQRInfo{
		Device:    "kitchen-switch",
		IP:        "192.168.1.100",
		MAC:       "AA:BB:CC:DD:EE:FF",
		Model:     "Plus 1PM",
		Firmware:  "1.0.0",
		WebURL:    "http://192.168.1.100",
		WiFiSSID:  "MyNetwork",
		QRContent: "http://192.168.1.100",
	}

	if info.Device != "kitchen-switch" {
		t.Errorf("Device = %q, want %q", info.Device, "kitchen-switch")
	}
	if info.IP != "192.168.1.100" {
		t.Errorf("IP = %q, want %q", info.IP, "192.168.1.100")
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("MAC = %q, want %q", info.MAC, "AA:BB:CC:DD:EE:FF")
	}
	if info.WebURL != "http://192.168.1.100" {
		t.Errorf("WebURL = %q, want %q", info.WebURL, "http://192.168.1.100")
	}
}

// TestBulkProvisionConfig tests BulkProvisionConfig struct.
func TestBulkProvisionConfig(t *testing.T) {
	t.Parallel()

	cfg := BulkProvisionConfig{
		WiFi: &ProvisionWiFiConfig{
			SSID:     "MyNetwork",
			Password: "secret123",
		},
		Devices: []DeviceProvisionConfig{
			{
				Name:    "device1",
				Address: "192.168.1.100",
				WiFi: &ProvisionWiFiConfig{
					SSID:     "DeviceNetwork",
					Password: "devicepass",
				},
				DevName: "Kitchen Switch",
			},
			{
				Name:    "device2",
				Address: "192.168.1.101",
			},
		},
	}

	if cfg.WiFi == nil {
		t.Fatal("WiFi should not be nil")
	}
	if cfg.WiFi.SSID != "MyNetwork" {
		t.Errorf("WiFi.SSID = %q, want %q", cfg.WiFi.SSID, "MyNetwork")
	}
	if len(cfg.Devices) != 2 {
		t.Fatalf("Devices len = %d, want 2", len(cfg.Devices))
	}
	if cfg.Devices[0].DevName != "Kitchen Switch" {
		t.Errorf("Devices[0].DevName = %q, want %q", cfg.Devices[0].DevName, "Kitchen Switch")
	}
}

// TestProvisionWiFiConfig tests ProvisionWiFiConfig struct.
func TestProvisionWiFiConfig(t *testing.T) {
	t.Parallel()

	cfg := ProvisionWiFiConfig{
		SSID:     "TestNetwork",
		Password: "testpass123",
	}

	if cfg.SSID != "TestNetwork" {
		t.Errorf("SSID = %q, want %q", cfg.SSID, "TestNetwork")
	}
	if cfg.Password != "testpass123" {
		t.Errorf("Password = %q, want %q", cfg.Password, "testpass123")
	}
}

// TestDeviceProvisionConfig tests DeviceProvisionConfig struct.
func TestDeviceProvisionConfig(t *testing.T) {
	t.Parallel()

	cfg := DeviceProvisionConfig{
		Name:    "switch1",
		Address: "192.168.1.100",
		WiFi: &ProvisionWiFiConfig{
			SSID:     "Override",
			Password: "override123",
		},
		DevName: "Living Room Switch",
	}

	if cfg.Name != "switch1" {
		t.Errorf("Name = %q, want %q", cfg.Name, "switch1")
	}
	if cfg.Address != "192.168.1.100" {
		t.Errorf("Address = %q, want %q", cfg.Address, "192.168.1.100")
	}
	if cfg.WiFi == nil {
		t.Fatal("WiFi should not be nil")
	}
	if cfg.DevName != "Living Room Switch" {
		t.Errorf("DevName = %q, want %q", cfg.DevName, "Living Room Switch")
	}
}

// TestProvisionResult tests ProvisionResult struct.
func TestProvisionResult(t *testing.T) {
	t.Parallel()

	// Success case
	result := ProvisionResult{
		Device: "device1",
		Err:    nil,
	}

	if result.Device != "device1" {
		t.Errorf("Device = %q, want %q", result.Device, "device1")
	}
	if result.Err != nil {
		t.Errorf("Err = %v, want nil", result.Err)
	}

	// Error case
	errResult := ProvisionResult{
		Device: "device2",
		Err:    ErrConnectionFailed,
	}

	if errResult.Err == nil {
		t.Error("Err should not be nil")
	}
}

// TestBackupFileInfo tests BackupFileInfo struct.
func TestBackupFileInfo(t *testing.T) {
	t.Parallel()

	now := time.Now()
	info := BackupFileInfo{
		Filename:    "backup_20240101.json",
		DeviceID:    "shellyplus1pm-123456",
		DeviceModel: "Plus 1PM",
		FWVersion:   "1.0.0",
		CreatedAt:   now,
		Encrypted:   true,
		Size:        1024,
	}

	if info.Filename != "backup_20240101.json" {
		t.Errorf("Filename = %q, want %q", info.Filename, "backup_20240101.json")
	}
	if info.DeviceID != "shellyplus1pm-123456" {
		t.Errorf("DeviceID = %q, want %q", info.DeviceID, "shellyplus1pm-123456")
	}
	if info.DeviceModel != "Plus 1PM" {
		t.Errorf("DeviceModel = %q, want %q", info.DeviceModel, "Plus 1PM")
	}
	if !info.Encrypted {
		t.Error("Encrypted = false, want true")
	}
	if info.Size != 1024 {
		t.Errorf("Size = %d, want 1024", info.Size)
	}
}

// TestComponentRGBW tests ComponentRGBW constant.
func TestComponentRGBW_Constant(t *testing.T) {
	t.Parallel()

	if ComponentRGBW != "rgbw" {
		t.Errorf("ComponentRGBW = %q, want %q", ComponentRGBW, "rgbw")
	}
}

// TestRGBWStatus tests RGBWStatus struct.
func TestRGBWStatus(t *testing.T) {
	t.Parallel()

	brightness := 80
	white := 50
	power := 15.0
	status := RGBWStatus{
		ID:         0,
		Output:     true,
		Brightness: &brightness,
		RGB: &RGBColor{
			Red:   255,
			Green: 200,
			Blue:  100,
		},
		White:     &white,
		Source:    "app",
		Power:     &power,
		Overtemp:  false,
		Overpower: false,
	}

	if !status.Output {
		t.Error("Output = false, want true")
	}
	if status.Brightness == nil || *status.Brightness != 80 {
		t.Errorf("Brightness = %v, want 80", status.Brightness)
	}
	if status.White == nil || *status.White != 50 {
		t.Errorf("White = %v, want 50", status.White)
	}
	if status.RGB == nil {
		t.Fatal("RGB should not be nil")
	}
}

// TestRGBWConfig tests RGBWConfig struct.
func TestRGBWConfig(t *testing.T) {
	t.Parallel()

	name := "LED Strip RGBW"
	cfg := RGBWConfig{
		ID:              0,
		Name:            &name,
		InitialState:    "off",
		AutoOn:          false,
		AutoOnDelay:     0,
		AutoOff:         true,
		AutoOffDelay:    3600,
		DefaultBright:   75,
		DefaultWhite:    50,
		NightModeEnable: true,
		NightModeBright: 10,
		NightModeWhite:  5,
	}

	if cfg.Name == nil || *cfg.Name != "LED Strip RGBW" {
		t.Errorf("Name = %v, want %q", cfg.Name, "LED Strip RGBW")
	}
	if cfg.DefaultBright != 75 {
		t.Errorf("DefaultBright = %d, want 75", cfg.DefaultBright)
	}
	if cfg.DefaultWhite != 50 {
		t.Errorf("DefaultWhite = %d, want 50", cfg.DefaultWhite)
	}
	if cfg.NightModeWhite != 5 {
		t.Errorf("NightModeWhite = %d, want 5", cfg.NightModeWhite)
	}
}

// TestComponentListItem tests ComponentListItem struct.
func TestComponentListItem(t *testing.T) {
	t.Parallel()

	item := ComponentListItem{
		ID:   0,
		Type: string(ComponentSwitch),
	}

	if item.ID != 0 {
		t.Errorf("ID = %d, want 0", item.ID)
	}
	if item.Type != string(ComponentSwitch) {
		t.Errorf("Type = %q, want %q", item.Type, ComponentSwitch)
	}
}
