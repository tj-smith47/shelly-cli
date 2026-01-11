// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/firmware"
)

// mockConnectionProvider is a test double for ConnectionProvider.
type mockConnectionProvider struct {
	withConnectionFunc     func(ctx context.Context, identifier string, fn func(*client.Client) error) error
	withGen1ConnectionFunc func(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
}

func (m *mockConnectionProvider) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFunc != nil {
		return m.withConnectionFunc(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	if m.withGen1ConnectionFunc != nil {
		return m.withGen1ConnectionFunc(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionProvider) ResolveWithGeneration(_ context.Context, identifier string) (model.Device, error) {
	return model.Device{Name: identifier, Address: identifier}, nil
}

func (m *mockConnectionProvider) GetCachedFirmware(_ context.Context, _ string, _ time.Duration) *firmware.CacheEntry {
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	svc := New(nil)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != nil {
		t.Error("expected parent to be nil")
	}
}

func TestInfo_Fields(t *testing.T) {
	t.Parallel()

	info := Info{
		ID:         "shellyplus2pm-123456",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Model:      "SNSW-002P16EU",
		Generation: 2,
		Firmware:   "1.0.0",
		App:        "Plus2PM",
		AuthEn:     true,
	}

	if info.ID != "shellyplus2pm-123456" {
		t.Errorf("got ID=%q, want %q", info.ID, "shellyplus2pm-123456")
	}
	if info.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got MAC=%q, want %q", info.MAC, "AA:BB:CC:DD:EE:FF")
	}
	if info.Model != "SNSW-002P16EU" {
		t.Errorf("got Model=%q, want %q", info.Model, "SNSW-002P16EU")
	}
	if info.Generation != 2 {
		t.Errorf("got Generation=%d, want 2", info.Generation)
	}
	if info.Firmware != "1.0.0" {
		t.Errorf("got Firmware=%q, want %q", info.Firmware, "1.0.0")
	}
	if info.App != "Plus2PM" {
		t.Errorf("got App=%q, want %q", info.App, "Plus2PM")
	}
	if !info.AuthEn {
		t.Error("expected AuthEn to be true")
	}
}

func TestStatus_Fields(t *testing.T) {
	t.Parallel()

	info := &Info{
		ID:    "test-device",
		Model: "SNSW-102",
	}

	status := Status{
		Info: info,
		Status: map[string]any{
			"uptime": 86400,
			"ram":    50000,
		},
	}

	if status.Info == nil {
		t.Fatal("expected Info to be set")
	}
	if status.Info.ID != "test-device" {
		t.Errorf("got Info.ID=%q, want %q", status.Info.ID, "test-device")
	}
	if status.Status == nil {
		t.Fatal("expected Status map to be set")
	}
	if status.Status["uptime"] != 86400 {
		t.Errorf("got uptime=%v, want 86400", status.Status["uptime"])
	}
}

func TestSysStatus_Fields(t *testing.T) {
	t.Parallel()

	status := SysStatus{
		MAC:             "AA:BB:CC:DD:EE:FF",
		Uptime:          86400,
		Time:            "14:30:00",
		Unixtime:        1700000000,
		RAMFree:         50000,
		RAMSize:         98304,
		FSFree:          25000,
		FSSize:          65536,
		RestartRequired: true,
		CfgRev:          5,
		UpdateAvailable: "1.2.3",
	}

	if status.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("got MAC=%q, want %q", status.MAC, "AA:BB:CC:DD:EE:FF")
	}
	if status.Uptime != 86400 {
		t.Errorf("got Uptime=%d, want 86400", status.Uptime)
	}
	if status.Time != "14:30:00" {
		t.Errorf("got Time=%q, want %q", status.Time, "14:30:00")
	}
	if status.Unixtime != 1700000000 {
		t.Errorf("got Unixtime=%d, want 1700000000", status.Unixtime)
	}
	if status.RAMFree != 50000 {
		t.Errorf("got RAMFree=%d, want 50000", status.RAMFree)
	}
	if status.RAMSize != 98304 {
		t.Errorf("got RAMSize=%d, want 98304", status.RAMSize)
	}
	if status.FSFree != 25000 {
		t.Errorf("got FSFree=%d, want 25000", status.FSFree)
	}
	if status.FSSize != 65536 {
		t.Errorf("got FSSize=%d, want 65536", status.FSSize)
	}
	if !status.RestartRequired {
		t.Error("expected RestartRequired to be true")
	}
	if status.CfgRev != 5 {
		t.Errorf("got CfgRev=%d, want 5", status.CfgRev)
	}
	if status.UpdateAvailable != "1.2.3" {
		t.Errorf("got UpdateAvailable=%q, want %q", status.UpdateAvailable, "1.2.3")
	}
}

func TestSysConfig_Fields(t *testing.T) {
	t.Parallel()

	cfg := SysConfig{
		Name:         "Living Room Device",
		Timezone:     "America/New_York",
		Lat:          40.7128,
		Lng:          -74.006,
		EcoMode:      true,
		Discoverable: true,
		Profile:      "switch",
		SNTPServer:   "pool.ntp.org",
	}

	if cfg.Name != "Living Room Device" {
		t.Errorf("got Name=%q, want %q", cfg.Name, "Living Room Device")
	}
	if cfg.Timezone != "America/New_York" {
		t.Errorf("got Timezone=%q, want %q", cfg.Timezone, "America/New_York")
	}
	if cfg.Lat != 40.7128 {
		t.Errorf("got Lat=%f, want 40.7128", cfg.Lat)
	}
	if cfg.Lng != -74.006 {
		t.Errorf("got Lng=%f, want -74.006", cfg.Lng)
	}
	if !cfg.EcoMode {
		t.Error("expected EcoMode to be true")
	}
	if !cfg.Discoverable {
		t.Error("expected Discoverable to be true")
	}
	if cfg.Profile != "switch" {
		t.Errorf("got Profile=%q, want %q", cfg.Profile, "switch")
	}
	if cfg.SNTPServer != "pool.ntp.org" {
		t.Errorf("got SNTPServer=%q, want %q", cfg.SNTPServer, "pool.ntp.org")
	}
}

func TestListFilterOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := ListFilterOptions{
		Generation: 2,
		DeviceType: "switch",
		Platform:   "plus",
	}

	if opts.Generation != 2 {
		t.Errorf("got Generation=%d, want 2", opts.Generation)
	}
	if opts.DeviceType != "switch" {
		t.Errorf("got DeviceType=%q, want %q", opts.DeviceType, "switch")
	}
	if opts.Platform != "plus" {
		t.Errorf("got Platform=%q, want %q", opts.Platform, "plus")
	}
}

func TestFilterList(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"dev1": {Name: "dev1", Address: "192.168.1.1", Generation: 2, Platform: "plus", Model: "SNSW-102", Type: "switch"},
		"dev2": {Name: "dev2", Address: "192.168.1.2", Generation: 1, Platform: "shelly1", Model: "SHSW-1", Type: "relay"},
		"dev3": {Name: "dev3", Address: "192.168.1.3", Generation: 2, Platform: "plus", Model: "SNDM-0013US", Type: "dimmer"},
	}

	tests := []struct {
		name         string
		opts         ListFilterOptions
		wantCount    int
		wantPlatform []string
	}{
		{
			name:         "no filter",
			opts:         ListFilterOptions{},
			wantCount:    3,
			wantPlatform: []string{"plus", "shelly1"},
		},
		{
			name:         "filter by generation 2",
			opts:         ListFilterOptions{Generation: 2},
			wantCount:    2,
			wantPlatform: []string{"plus"},
		},
		{
			name:         "filter by generation 1",
			opts:         ListFilterOptions{Generation: 1},
			wantCount:    1,
			wantPlatform: []string{"shelly1"},
		},
		{
			name:         "filter by platform",
			opts:         ListFilterOptions{Platform: "plus"},
			wantCount:    2,
			wantPlatform: []string{"plus"},
		},
		{
			name:         "filter by device type",
			opts:         ListFilterOptions{DeviceType: "switch"},
			wantCount:    1,
			wantPlatform: []string{"plus"},
		},
		{
			name:         "combined filters",
			opts:         ListFilterOptions{Generation: 2, DeviceType: "dimmer"},
			wantCount:    1,
			wantPlatform: []string{"plus"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filtered, platforms := FilterList(devices, tt.opts)

			if len(filtered) != tt.wantCount {
				t.Errorf("got %d devices, want %d", len(filtered), tt.wantCount)
			}

			for _, p := range tt.wantPlatform {
				if _, ok := platforms[p]; !ok {
					t.Errorf("expected platform %q in result", p)
				}
			}
		})
	}
}

func TestSortList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		devices      []model.DeviceListItem
		updatesFirst bool
		wantFirst    string
	}{
		{
			name: "sort by name",
			devices: []model.DeviceListItem{
				{Name: "Zebra"},
				{Name: "Apple"},
				{Name: "Mango"},
			},
			updatesFirst: false,
			wantFirst:    "Apple",
		},
		{
			name: "updates first",
			devices: []model.DeviceListItem{
				{Name: "Apple", HasUpdate: false},
				{Name: "Zebra", HasUpdate: true},
				{Name: "Mango", HasUpdate: false},
			},
			updatesFirst: true,
			wantFirst:    "Zebra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			devices := make([]model.DeviceListItem, len(tt.devices))
			copy(devices, tt.devices)

			SortList(devices, tt.updatesFirst)

			if devices[0].Name != tt.wantFirst {
				t.Errorf("got first=%q, want %q", devices[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestMatchesFilters(t *testing.T) {
	t.Parallel()

	dev := model.Device{
		Name:       "Kitchen Switch",
		Generation: 2,
		Platform:   "plus",
		Model:      "SNSW-102",
		Type:       "switch",
	}

	tests := []struct {
		name string
		opts ListFilterOptions
		want bool
	}{
		{
			name: "no filters match",
			opts: ListFilterOptions{},
			want: true,
		},
		{
			name: "generation matches",
			opts: ListFilterOptions{Generation: 2},
			want: true,
		},
		{
			name: "generation does not match",
			opts: ListFilterOptions{Generation: 1},
			want: false,
		},
		{
			name: "platform matches",
			opts: ListFilterOptions{Platform: "plus"},
			want: true,
		},
		{
			name: "platform does not match",
			opts: ListFilterOptions{Platform: "pro"},
			want: false,
		},
		{
			name: "device type matches",
			opts: ListFilterOptions{DeviceType: "switch"},
			want: true,
		},
		{
			name: "device type does not match",
			opts: ListFilterOptions{DeviceType: "light"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := matchesFilters(dev, tt.opts)

			if got != tt.want {
				t.Errorf("matchesFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConnectionProvider_Interface verifies that mockConnectionProvider satisfies ConnectionProvider.
func TestConnectionProvider_Interface(t *testing.T) {
	t.Parallel()
	// This compile-time check ensures mockConnectionProvider implements ConnectionProvider
	var _ ConnectionProvider = (*mockConnectionProvider)(nil)
}

func TestGetFullStatus_ServiceCreation(t *testing.T) {
	t.Parallel()

	provider := &mockConnectionProvider{}
	svc := New(provider)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.parent != provider {
		t.Error("expected parent to be the provider")
	}
}

func TestGetFullStatus_ResponseParsing(t *testing.T) {
	t.Parallel()

	// Simulate a Gen2 Shelly.GetStatus response structure
	testResponse := map[string]any{
		"switch:0": map[string]any{
			"id":     0,
			"output": true,
		},
		"wifi": map[string]any{
			"sta_ip": "192.168.1.100",
			"rssi":   -45,
		},
		"sys": map[string]any{
			"uptime": 86400,
		},
	}

	// Marshal and unmarshal to simulate what GetFullStatus does
	data, err := json.Marshal(testResponse)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal to RawMessage map: %v", err)
	}

	// Verify we got all components
	if len(result) != 3 {
		t.Errorf("got %d components, want 3", len(result))
	}

	// Verify switch:0 is present
	if _, ok := result["switch:0"]; !ok {
		t.Error("expected switch:0 to be present")
	}

	// Verify wifi is present
	if _, ok := result["wifi"]; !ok {
		t.Error("expected wifi to be present")
	}

	// Verify sys is present
	if _, ok := result["sys"]; !ok {
		t.Error("expected sys to be present")
	}

	// Verify we can parse a component
	var switchStatus map[string]any
	if err := json.Unmarshal(result["switch:0"], &switchStatus); err != nil {
		t.Fatalf("failed to parse switch:0: %v", err)
	}

	if switchStatus["output"] != true {
		t.Errorf("got output=%v, want true", switchStatus["output"])
	}
}

func TestGetFullStatusGen1_ResponseParsing(t *testing.T) {
	t.Parallel()

	// Simulate a Gen1 /status response structure
	gen1Response := `{
		"wifi_sta": {"connected": true, "ssid": "TestNetwork", "ip": "192.168.1.100"},
		"relays": [{"ison": true, "source": "input"}],
		"meters": [{"power": 100.5, "is_valid": true}],
		"uptime": 86400
	}`

	var result map[string]json.RawMessage
	if err := json.Unmarshal([]byte(gen1Response), &result); err != nil {
		t.Fatalf("failed to unmarshal Gen1 status: %v", err)
	}

	// Verify we got expected fields
	if _, ok := result["wifi_sta"]; !ok {
		t.Error("expected wifi_sta to be present")
	}
	if _, ok := result["relays"]; !ok {
		t.Error("expected relays to be present")
	}
	if _, ok := result["meters"]; !ok {
		t.Error("expected meters to be present")
	}
	if _, ok := result["uptime"]; !ok {
		t.Error("expected uptime to be present")
	}

	// Verify we can parse a component
	var relays []map[string]any
	if err := json.Unmarshal(result["relays"], &relays); err != nil {
		t.Fatalf("failed to parse relays: %v", err)
	}

	if len(relays) != 1 {
		t.Errorf("got %d relays, want 1", len(relays))
	}
	if relays[0]["ison"] != true {
		t.Errorf("got ison=%v, want true", relays[0]["ison"])
	}
}
