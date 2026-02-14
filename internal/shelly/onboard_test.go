package shelly

import (
	"encoding/json"
	"testing"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

const testSSID = "MyNetwork"

func TestOnboardSource_Constants(t *testing.T) {
	t.Parallel()

	sources := []OnboardSource{
		OnboardSourceBLE,
		OnboardSourceWiFiAP,
		OnboardSourceMDNS,
		OnboardSourceCoIoT,
		OnboardSourceHTTP,
	}

	for _, s := range sources {
		if s == "" {
			t.Error("source constant is empty")
		}
	}
}

func TestOnboardDevice_Fields(t *testing.T) {
	t.Parallel()

	dev := OnboardDevice{
		Name:        "test-device",
		Model:       "SNSW-001P16EU",
		Address:     "192.168.1.100",
		MACAddress:  "AA:BB:CC:DD:EE:FF",
		SSID:        "ShellyPlus1-AABBCC",
		BLEAddress:  "AA:BB:CC:DD:EE:FF",
		Source:      OnboardSourceBLE,
		Generation:  2,
		RSSI:        -55,
		Registered:  false,
		Provisioned: false,
	}

	if dev.Name != "test-device" {
		t.Errorf("Name = %q, want %q", dev.Name, "test-device")
	}
	if dev.Source != OnboardSourceBLE {
		t.Errorf("Source = %q, want %q", dev.Source, OnboardSourceBLE)
	}
	if dev.Generation != 2 {
		t.Errorf("Generation = %d, want 2", dev.Generation)
	}
}

func TestOnboardWiFiConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := OnboardWiFiConfig{SSID: testSSID, Password: "secret"}
	if cfg.SSID != testSSID {
		t.Errorf("SSID = %q, want %q", cfg.SSID, testSSID)
	}
	if cfg.Password != "secret" {
		t.Errorf("Password = %q, want %q", cfg.Password, "secret")
	}
}

func TestOnboardOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &OnboardOptions{}
	if opts.BLEOnly {
		t.Error("BLEOnly should default to false")
	}
	if opts.APOnly {
		t.Error("APOnly should default to false")
	}
	if opts.NoCloud {
		t.Error("NoCloud should default to false")
	}
}

func TestOnboardResult_Fields(t *testing.T) {
	t.Parallel()

	dev := &OnboardDevice{Name: "test"}
	result := &OnboardResult{
		Device:     dev,
		NewAddress: "192.168.1.50",
		Registered: true,
		Method:     "BLE",
	}

	if result.Device.Name != "test" {
		t.Errorf("Device.Name = %q, want %q", result.Device.Name, "test")
	}
	if result.NewAddress != "192.168.1.50" {
		t.Errorf("NewAddress = %q, want %q", result.NewAddress, "192.168.1.50")
	}
	if result.Method != "BLE" {
		t.Errorf("Method = %q, want %q", result.Method, "BLE")
	}
	if !result.Registered {
		t.Error("Registered should be true")
	}
}

func TestOnboardProgress_Fields(t *testing.T) {
	t.Parallel()

	p := OnboardProgress{Method: "BLE", Found: 3, Done: true}
	if p.Method != "BLE" {
		t.Errorf("Method = %q, want %q", p.Method, "BLE")
	}
	if p.Found != 3 {
		t.Errorf("Found = %d, want 3", p.Found)
	}
	if !p.Done {
		t.Error("Done should be true")
	}
}

func TestSanitizeDeviceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "my-device", "my-device"},
		{"uppercase", "My Device", "my-device"},
		{"spaces", "  test  device  ", "test--device"},
		{"special chars", "device@#$123", "device123"},
		{"underscores", "my_device_01", "my_device_01"},
		{"empty", "", ""},
		{"only special", "!@#$%", ""},
		{"mixed", "Shelly Plus 1PM (Living Room)", "shelly-plus-1pm-living-room"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sanitizeDeviceName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeDeviceName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterUnregistered(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "new-1", Registered: false},
		{Name: "existing-1", Registered: true},
		{Name: "new-2", Registered: false},
		{Name: "existing-2", Registered: true},
	}

	filtered := FilterUnregistered(devices)
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	if filtered[0].Name != "new-1" {
		t.Errorf("filtered[0].Name = %q, want %q", filtered[0].Name, "new-1")
	}
	if filtered[1].Name != "new-2" {
		t.Errorf("filtered[1].Name = %q, want %q", filtered[1].Name, "new-2")
	}
}

func TestFilterUnregistered_AllRegistered(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "dev-1", Registered: true},
		{Name: "dev-2", Registered: true},
	}

	filtered := FilterUnregistered(devices)
	if len(filtered) != 0 {
		t.Errorf("len(filtered) = %d, want 0", len(filtered))
	}
}

func TestFilterUnregistered_NoneRegistered(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "dev-1", Registered: false},
		{Name: "dev-2", Registered: false},
	}

	filtered := FilterUnregistered(devices)
	if len(filtered) != 2 {
		t.Errorf("len(filtered) = %d, want 2", len(filtered))
	}
}

func TestFilterUnregistered_Empty(t *testing.T) {
	t.Parallel()

	filtered := FilterUnregistered(nil)
	if len(filtered) != 0 {
		t.Errorf("len(filtered) = %d, want 0", len(filtered))
	}
}

func TestSplitBySource(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "ble-1", Source: OnboardSourceBLE},
		{Name: "ap-1", Source: OnboardSourceWiFiAP},
		{Name: "mdns-1", Source: OnboardSourceMDNS},
		{Name: "ble-2", Source: OnboardSourceBLE},
		{Name: "coiot-1", Source: OnboardSourceCoIoT},
		{Name: "http-1", Source: OnboardSourceHTTP},
	}

	bleDevs, apDevs, netDevs := SplitBySource(devices)

	if len(bleDevs) != 2 {
		t.Errorf("len(ble) = %d, want 2", len(bleDevs))
	}
	if len(apDevs) != 1 {
		t.Errorf("len(ap) = %d, want 1", len(apDevs))
	}
	if len(netDevs) != 3 {
		t.Errorf("len(network) = %d, want 3", len(netDevs))
	}

	if bleDevs[0].Name != "ble-1" {
		t.Errorf("ble[0].Name = %q, want %q", bleDevs[0].Name, "ble-1")
	}
	if apDevs[0].Name != "ap-1" {
		t.Errorf("ap[0].Name = %q, want %q", apDevs[0].Name, "ap-1")
	}
}

func TestSplitBySource_Empty(t *testing.T) {
	t.Parallel()

	bleDevs, apDevs, netDevs := SplitBySource(nil)
	if len(bleDevs) != 0 || len(apDevs) != 0 || len(netDevs) != 0 {
		t.Error("expected all slices to be empty")
	}
}

func TestSplitBySource_BLEOnly(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "ble-1", Source: OnboardSourceBLE},
		{Name: "ble-2", Source: OnboardSourceBLE},
	}

	bleDevs, apDevs, netDevs := SplitBySource(devices)
	if len(bleDevs) != 2 {
		t.Errorf("len(ble) = %d, want 2", len(bleDevs))
	}
	if len(apDevs) != 0 {
		t.Errorf("len(ap) = %d, want 0", len(apDevs))
	}
	if len(netDevs) != 0 {
		t.Errorf("len(network) = %d, want 0", len(netDevs))
	}
}

func TestDeduplicateOnboardDevices_ByMAC(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "mdns-dev", MACAddress: "AA:BB:CC:DD:EE:FF", Source: OnboardSourceMDNS},
		{Name: "ble-dev", MACAddress: "AA:BB:CC:DD:EE:FF", Source: OnboardSourceBLE},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 1 {
		t.Fatalf("len(deduped) = %d, want 1", len(deduped))
	}
	// BLE should be preferred
	if deduped[0].Source != OnboardSourceBLE {
		t.Errorf("Source = %q, want %q (BLE preferred)", deduped[0].Source, OnboardSourceBLE)
	}
}

func TestDeduplicateOnboardDevices_ByName(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "shelly-plus-1pm", MACAddress: "", Source: OnboardSourceMDNS},
		{Name: "Shelly-Plus-1PM", MACAddress: "", Source: OnboardSourceBLE},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 1 {
		t.Fatalf("len(deduped) = %d, want 1", len(deduped))
	}
	// BLE should be preferred
	if deduped[0].Source != OnboardSourceBLE {
		t.Errorf("Source = %q, want %q", deduped[0].Source, OnboardSourceBLE)
	}
}

func TestDeduplicateOnboardDevices_NoDuplicates(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "dev-1", MACAddress: "AA:BB:CC:DD:EE:01", Source: OnboardSourceBLE},
		{Name: "dev-2", MACAddress: "AA:BB:CC:DD:EE:02", Source: OnboardSourceMDNS},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 2 {
		t.Errorf("len(deduped) = %d, want 2", len(deduped))
	}
}

func TestDeduplicateOnboardDevices_Empty(t *testing.T) {
	t.Parallel()

	deduped := deduplicateOnboardDevices(nil)
	if len(deduped) != 0 {
		t.Errorf("len(deduped) = %d, want 0", len(deduped))
	}
}

func TestDeduplicateOnboardDevices_NoMACOrName(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "", MACAddress: "", Address: "192.168.1.1"},
		{Name: "", MACAddress: "", Address: "192.168.1.2"},
	}

	deduped := deduplicateOnboardDevices(devices)
	// Both should be kept since there's no key to dedup on
	if len(deduped) != 2 {
		t.Errorf("len(deduped) = %d, want 2", len(deduped))
	}
}

func TestDeduplicateOnboardDevices_PrefersBLEOverOthers(t *testing.T) {
	t.Parallel()

	mac := "AA:BB:CC:DD:EE:FF"
	devices := []OnboardDevice{
		{Name: "coiot", MACAddress: mac, Source: OnboardSourceCoIoT, Address: "192.168.1.5"},
		{Name: "ble", MACAddress: mac, Source: OnboardSourceBLE, BLEAddress: "XX:YY:ZZ"},
		{Name: "http", MACAddress: mac, Source: OnboardSourceHTTP, Address: "192.168.1.5"},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 1 {
		t.Fatalf("len(deduped) = %d, want 1", len(deduped))
	}
	if deduped[0].Source != OnboardSourceBLE {
		t.Errorf("Source = %q, want BLE", deduped[0].Source)
	}
	if deduped[0].BLEAddress != "XX:YY:ZZ" {
		t.Errorf("BLEAddress = %q, want %q", deduped[0].BLEAddress, "XX:YY:ZZ")
	}
}

func TestDeduplicateOnboardDevices_MACCaseInsensitive(t *testing.T) {
	t.Parallel()

	devices := []OnboardDevice{
		{Name: "dev-1", MACAddress: "aa:bb:cc:dd:ee:ff", Source: OnboardSourceMDNS},
		{Name: "dev-2", MACAddress: "AA:BB:CC:DD:EE:FF", Source: OnboardSourceCoIoT},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 1 {
		t.Errorf("len(deduped) = %d, want 1 (MAC should be case-insensitive)", len(deduped))
	}
}

func TestRegisterNetworkDevices(t *testing.T) {
	t.Parallel()

	devices := []*OnboardDevice{
		{Name: "dev-1", Address: "192.168.1.50"},
		{Name: "dev-2", Address: ""},
	}

	results := RegisterNetworkDevices(devices)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	// dev-2 has no address, should error
	if results[1].Error == nil {
		t.Error("expected error for device with no address")
	}
	if results[0].Method != "register-only" {
		t.Errorf("Method = %q, want %q", results[0].Method, "register-only")
	}
}

func TestRegisterNetworkDevices_Empty(t *testing.T) {
	t.Parallel()

	results := RegisterNetworkDevices(nil)
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestExtractWiFiFromBackup_Gen1(t *testing.T) {
	t.Parallel()

	// Simulate Gen1 settings with WiFi STA config
	settings := map[string]any{
		"wifi_sta": map[string]any{
			"ssid":    testSSID,
			"key":     "secret123",
			"enabled": true,
		},
	}
	configData, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
			Config:     configData,
		},
	}

	creds := extractWiFiFromBackup(bkp)
	if creds == nil {
		t.Fatal("expected WiFi credentials, got nil")
	}
	if creds.SSID != testSSID {
		t.Errorf("SSID = %q, want %q", creds.SSID, testSSID)
	}
	if creds.Password != "secret123" {
		t.Errorf("Password = %q, want %q", creds.Password, "secret123")
	}
}

func TestExtractWiFiFromBackup_Gen1_NoWiFi(t *testing.T) {
	t.Parallel()

	configData := json.RawMessage(`{}`)
	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
			Config:     configData,
		},
	}

	creds := extractWiFiFromBackup(bkp)
	if creds != nil {
		t.Errorf("expected nil, got %+v", creds)
	}
}

func TestExtractWiFiFromBackup_Gen2_WithSSID(t *testing.T) {
	t.Parallel()

	wifiBlob := json.RawMessage(`{"sta":{"ssid":"` + testSSID + `","is_open":false,"enable":true}}`)

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 2},
			Config:     json.RawMessage(`{}`),
			WiFi:       wifiBlob,
		},
	}

	creds := extractWiFiFromBackup(bkp)
	if creds == nil {
		t.Fatal("expected WiFi credentials, got nil")
	}
	if creds.SSID != testSSID {
		t.Errorf("SSID = %q, want %q", creds.SSID, testSSID)
	}
	// Gen2+ doesn't return password
	if creds.Password != "" {
		t.Errorf("Password should be empty for Gen2+, got %q", creds.Password)
	}
}

func TestExtractWiFiFromBackup_NilConfig(t *testing.T) {
	t.Parallel()

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
		},
	}

	creds := extractWiFiFromBackup(bkp)
	if creds != nil {
		t.Errorf("expected nil for nil config, got %+v", creds)
	}
}

func TestExtractSSIDFromWiFiBlob_Nil(t *testing.T) {
	t.Parallel()

	creds := extractSSIDFromWiFiBlob(nil)
	if creds != nil {
		t.Errorf("expected nil, got %+v", creds)
	}
}

func TestExtractSSIDFromWiFiBlob_InvalidJSON(t *testing.T) {
	t.Parallel()

	creds := extractSSIDFromWiFiBlob(json.RawMessage(`{invalid`))
	if creds != nil {
		t.Errorf("expected nil for invalid JSON, got %+v", creds)
	}
}

func TestExtractSSIDFromWiFiBlob_NoSTA(t *testing.T) {
	t.Parallel()

	creds := extractSSIDFromWiFiBlob(json.RawMessage(`{"ap":{"ssid":"test"}}`))
	if creds != nil {
		t.Errorf("expected nil when no sta key, got %+v", creds)
	}
}

func TestExtractSSIDFromWiFiBlob_EmptySSID(t *testing.T) {
	t.Parallel()

	creds := extractSSIDFromWiFiBlob(json.RawMessage(`{"sta":{"ssid":""}}`))
	if creds != nil {
		t.Errorf("expected nil for empty SSID, got %+v", creds)
	}
}

func TestProvisionSource_Types(t *testing.T) {
	t.Parallel()

	// Test with backup
	source := &ProvisionSource{
		Backup: &backup.DeviceBackup{Backup: &shellybackup.Backup{}},
		WiFi:   &OnboardWiFiConfig{SSID: testSSID, Password: "pass"},
	}
	if source.Backup == nil {
		t.Error("Backup should not be nil")
	}
	if source.WiFi.SSID != testSSID {
		t.Errorf("WiFi.SSID = %q, want %q", source.WiFi.SSID, testSSID)
	}

	// Test with template
	source2 := &ProvisionSource{
		Template: &config.DeviceTemplate{
			Name:  "test-tpl",
			Model: "SHBDUO-1",
		},
	}
	if source2.Template == nil {
		t.Error("Template should not be nil")
	}
	if source2.Template.Name != "test-tpl" {
		t.Errorf("Template.Name = %q, want %q", source2.Template.Name, "test-tpl")
	}
}
