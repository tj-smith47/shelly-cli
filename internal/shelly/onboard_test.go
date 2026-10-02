package shelly

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"
	"github.com/tj-smith47/shelly-go/discovery"
	"github.com/tj-smith47/shelly-go/reprovision"

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

func TestDeduplicateOnboardDevices_BLEAndAPCollapseBySuffix(t *testing.T) {
	t.Parallel()

	// One physical device advertising both BLE (LocalName) and a WiFi AP (SSID).
	// Their MACs differ — the BLE entry carries the BT radio MAC while the AP
	// entry carries the WiFi BSSID — so only the shared Shelly device-ID suffix
	// can collapse them.
	devices := []OnboardDevice{
		{
			Name:       "shellyplus1pm-aabbcc",
			SSID:       "shellyplus1pm-aabbcc",
			MACAddress: "11:22:33:44:55:66", // WiFi BSSID
			Source:     OnboardSourceWiFiAP,
			Address:    "192.168.33.1",
		},
		{
			Name:       "ShellyPlus1PM-AABBCC",
			MACAddress: "AA:BB:CC:00:11:22", // BT radio MAC (distinct from BSSID)
			BLEAddress: "AA:BB:CC:00:11:22",
			Source:     OnboardSourceBLE,
		},
	}

	deduped := deduplicateOnboardDevices(devices)
	if len(deduped) != 1 {
		t.Fatalf("len(deduped) = %d, want 1 (BLE+AP rows for one device must collapse)", len(deduped))
	}
	if deduped[0].Source != OnboardSourceBLE {
		t.Errorf("Source = %q, want %q (BLE preferred)", deduped[0].Source, OnboardSourceBLE)
	}
}

func TestShellyDeviceIDSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ble local name", "ShellyPlus1PM-AABBCC", "aabbcc"},
		{"ap ssid lowercase", "shellyplus1pm-aabbcc", "aabbcc"},
		{"non-shelly name returns empty", "kitchen-light", ""},
		{"empty returns empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := shellyDeviceIDSuffix(tt.input); got != tt.want {
				t.Errorf("shellyDeviceIDSuffix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
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

func TestProvisionWiFi_Gen1(t *testing.T) {
	t.Parallel()

	// Gen1 WiFi blob as produced by marshalGen1WiFi: {"sta": {WiFiStaSettings}}.
	wifiBlob := json.RawMessage(`{"sta":{"ssid":"` + testSSID + `","key":"secret123","enabled":true}}`)

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
			WiFi:       wifiBlob,
		},
	}

	creds := provisionWiFi(bkp)
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

func TestProvisionWiFi_Gen1_FallbackToConfig(t *testing.T) {
	t.Parallel()

	// No WiFi blob, but Config has the full gen1.Settings with wifi_sta.
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

	creds := provisionWiFi(bkp)
	if creds == nil {
		t.Fatal("expected WiFi credentials from Config fallback, got nil")
	}
	if creds.SSID != testSSID {
		t.Errorf("SSID = %q, want %q", creds.SSID, testSSID)
	}
	if creds.Password != "secret123" {
		t.Errorf("Password = %q, want %q", creds.Password, "secret123")
	}
}

func TestProvisionWiFi_Gen1_NoWiFi(t *testing.T) {
	t.Parallel()

	configData := json.RawMessage(`{}`)
	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
			Config:     configData,
		},
	}

	creds := provisionWiFi(bkp)
	if creds != nil {
		t.Errorf("expected nil, got %+v", creds)
	}
}

func TestProvisionWiFi_Gen2_WithSSID(t *testing.T) {
	t.Parallel()

	wifiBlob := json.RawMessage(`{"sta":{"ssid":"` + testSSID + `","is_open":false,"enable":true}}`)

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 2},
			Config:     json.RawMessage(`{}`),
			WiFi:       wifiBlob,
		},
	}

	creds := provisionWiFi(bkp)
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
	// A missing key is not an open network.
	if creds.Open {
		t.Error("a secured station with no key in the backup must not be marked open")
	}
}

func TestProvisionWiFi_OpenStation(t *testing.T) {
	t.Parallel()

	bkp := &backup.DeviceBackup{Backup: &shellybackup.Backup{
		DeviceInfo: &shellybackup.DeviceInfo{Generation: 2},
		WiFi:       json.RawMessage(`{"sta":{"ssid":"` + testSSID + `","is_open":true,"enable":true}}`),
	}}

	creds := provisionWiFi(bkp)
	want := OnboardWiFiConfig{SSID: testSSID, Open: true}
	if creds == nil || *creds != want {
		t.Errorf("creds = %+v, want %+v", creds, want)
	}
}

func TestProvisionWiFi_NilConfig(t *testing.T) {
	t.Parallel()

	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{
			DeviceInfo: &shellybackup.DeviceInfo{Generation: 1},
		},
	}

	creds := provisionWiFi(bkp)
	if creds != nil {
		t.Errorf("expected nil for nil config, got %+v", creds)
	}
}

func TestProvisionWiFi_LeavesOutStaticAddress(t *testing.T) {
	t.Parallel()

	wifiBlob := json.RawMessage(`{"sta":{"ssid":"` + testSSID + `","key":"secret123","enabled":true,` +
		`"ipv4_method":"static","ip":"10.0.0.5","gw":"10.0.0.1","mask":"255.255.255.0"}}`)
	bkp := &backup.DeviceBackup{
		Backup: &shellybackup.Backup{DeviceInfo: &shellybackup.DeviceInfo{Generation: 1}, WiFi: wifiBlob},
	}

	creds := provisionWiFi(bkp)
	want := OnboardWiFiConfig{SSID: testSSID, Password: "secret123"}
	if creds == nil || *creds != want {
		t.Errorf("creds = %+v, want %+v", creds, want)
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

// recordingScanner is an in-memory WiFi backend that records every call, so a
// test can prove it is the backend a service method used.
type recordingScanner struct {
	mu        sync.Mutex
	calls     []string
	current   string
	networks  []discovery.WiFiNetwork
	passwords map[string]string
}

func (r *recordingScanner) record(call string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, call)
}

func (r *recordingScanner) called() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.calls...)
}

func (r *recordingScanner) Scan(context.Context) ([]discovery.WiFiNetwork, error) {
	r.record("scan")
	return r.networks, nil
}

func (r *recordingScanner) Connect(_ context.Context, ssid, password string) error {
	r.record("connect:" + ssid + ":" + password)
	return errors.New("connect refused in tests")
}

func (r *recordingScanner) Disconnect(context.Context) error {
	r.record("disconnect")
	return errors.New("disconnect refused in tests")
}

func (r *recordingScanner) CurrentNetwork(context.Context) (*discovery.WiFiNetwork, error) {
	r.record("current")
	if r.current == "" {
		return nil, errors.New("not on WiFi")
	}
	return &discovery.WiFiNetwork{SSID: r.current}, nil
}

func (r *recordingScanner) HostNetworkPassword(_ context.Context, ssid string) (string, error) {
	r.record("password:" + ssid)
	if pw, ok := r.passwords[ssid]; ok {
		return pw, nil
	}
	return "", errors.New("no stored passphrase")
}

func TestHostWiFiPassword_UsesInjectedScanner(t *testing.T) {
	t.Parallel()

	scanner := &recordingScanner{passwords: map[string]string{"iot": "stored"}}
	svc := New(NewConfigResolver(), WithWiFiScanner(scanner))

	pw, err := svc.HostWiFiPassword(context.Background(), "iot")
	if err != nil || pw != "stored" {
		t.Errorf("HostWiFiPassword = %q, %v; want the injected scanner's passphrase", pw, err)
	}
	if _, err := svc.HostWiFiPassword(context.Background(), "guest"); err == nil {
		t.Error("an unknown network must be an error")
	}
	if _, err := svc.HostWiFiPassword(context.Background(), ""); err == nil {
		t.Error("no network named must be an error")
	}
	if got := strings.Join(scanner.called(), ","); got != "password:iot,password:guest" {
		t.Errorf("scanner calls = %s", got)
	}

	noProvider := New(NewConfigResolver(), WithWiFiScanner(fakeScanner{}))
	if _, err := noProvider.HostWiFiPassword(context.Background(), "iot"); err == nil {
		t.Error("a backend without passphrase recovery must be an error")
	}
}

func TestHostWiFiCredentials_UsesInjectedScanner(t *testing.T) {
	t.Parallel()

	scanner := &recordingScanner{current: "iot", passwords: map[string]string{"iot": "stored"}}
	svc := New(NewConfigResolver(), WithWiFiScanner(scanner))

	got := svc.HostWiFiCredentials(context.Background())
	if got == nil || *got != (OnboardWiFiConfig{SSID: "iot", Password: "stored"}) {
		t.Errorf("HostWiFiCredentials = %+v", got)
	}
	if calls := strings.Join(scanner.called(), ","); calls != "current,password:iot" {
		t.Errorf("scanner calls = %s", calls)
	}
}

func TestDiscoverWiFiAPForOnboard_UsesInjectedScanner(t *testing.T) {
	t.Parallel()

	scanner := &recordingScanner{networks: []discovery.WiFiNetwork{{SSID: testAPSSID, Signal: -40}}}
	svc := New(NewConfigResolver(), WithWiFiScanner(scanner))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	devices, err := svc.discoverWiFiAPForOnboard(ctx)
	if err != nil {
		t.Fatalf("discoverWiFiAPForOnboard: %v", err)
	}
	if len(devices) != 1 || devices[0].SSID != testAPSSID {
		t.Errorf("devices = %+v, want the injected scanner's AP", devices)
	}
	for _, call := range scanner.called() {
		if call != "scan" {
			t.Errorf("discovery made a %s call", call)
		}
	}
}

func TestJoinPassword(t *testing.T) {
	t.Parallel()

	scanner := &recordingScanner{passwords: map[string]string{"iot": "stored"}}
	svc := New(NewConfigResolver(), WithWiFiScanner(scanner))
	tests := []struct {
		name    string
		wifi    OnboardWiFiConfig
		want    OnboardWiFiConfig
		wantErr string
	}{
		{name: "password kept", wifi: OnboardWiFiConfig{SSID: "iot", Password: "given"},
			want: OnboardWiFiConfig{SSID: "iot", Password: "given"}},
		{name: "open kept", wifi: OnboardWiFiConfig{SSID: "guest", Open: true},
			want: OnboardWiFiConfig{SSID: "guest", Open: true}},
		{name: "host passphrase filled in", wifi: OnboardWiFiConfig{SSID: "iot"},
			want: OnboardWiFiConfig{SSID: "iot", Password: "stored"}},
		{name: "none found is refused", wifi: OnboardWiFiConfig{SSID: "guest"},
			wantErr: `no WiFi passphrase for "guest":`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := svc.joinPassword(context.Background(), &tt.wifi)
			if tt.wantErr != "" {
				if err == nil || !strings.HasPrefix(err.Error(), tt.wantErr) || !errors.Is(err, reprovision.ErrNoPassphrase) {
					t.Errorf("err = %v, want prefix %s", err, tt.wantErr)
				}
				return
			}
			if err != nil || *got != tt.want {
				t.Errorf("joinPassword = %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}

// TestOnboardBLEParallel_RefusesUnknownPassword proves a BLE onboard never writes
// a secured network as an open one: with no password, no Open and none stored on
// the host, every device fails with the passphrase error before BLE is touched.
func TestOnboardBLEParallel_RefusesUnknownPassword(t *testing.T) {
	t.Parallel()

	svc := New(NewConfigResolver(), WithWiFiScanner(&recordingScanner{}))
	devices := []*OnboardDevice{{Name: "ShellyPlus1-AABBCC"}, {Name: "ShellyPlus1-DDEEFF"}}

	results := svc.OnboardBLEParallel(context.Background(), devices, &OnboardWiFiConfig{SSID: "iot"}, &OnboardOptions{})

	if len(results) != len(devices) {
		t.Fatalf("results = %d, want %d", len(results), len(devices))
	}
	for i, r := range results {
		if r.Device != devices[i] || !errors.Is(r.Error, reprovision.ErrNoPassphrase) {
			t.Errorf("result %d = %+v, want the passphrase error", i, r)
		}
	}
}

func TestOnboardViaBLE_RefusesUnknownPassword(t *testing.T) {
	t.Parallel()

	svc := New(NewConfigResolver(), WithWiFiScanner(&recordingScanner{}))
	dev := &OnboardDevice{Name: "ShellyPlus1-AABBCC"}

	r := svc.OnboardViaBLE(context.Background(), dev, &OnboardWiFiConfig{SSID: "iot"}, &OnboardOptions{})

	if r.Device != dev || r.Method != string(OnboardSourceBLE) || !errors.Is(r.Error, reprovision.ErrNoPassphrase) {
		t.Errorf("result = %+v, want the passphrase error", r)
	}
}

func TestScanner_OfflineInTests(t *testing.T) {
	t.Parallel()
	if _, ok := (&Service{}).scanner().(OfflineWiFiScanner); !ok {
		t.Error("a service built without a scanner reaches the platform WiFi in a test binary")
	}
	scanner := &recordingScanner{}
	if (&Service{wifiScanner: scanner}).scanner() != scanner {
		t.Error("an injected scanner is not used")
	}
}
