// Package firmware provides firmware management for Shelly devices.
package firmware

import (
	"context"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

const (
	testDeviceNameDevice1 = "device1"
	testDeviceNameBeta    = "beta"
	testStageStable       = "stable"
)

// mockConnectionHandler is a test double for ConnectionHandler.
type mockConnectionHandler struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

func (m *mockConnectionHandler) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func TestNewService(t *testing.T) {
	t.Parallel()

	handler := &mockConnectionHandler{}
	svc := NewService(handler)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.connHandler != handler {
		t.Error("expected connHandler to be set")
	}
	if svc.cache == nil {
		t.Error("expected cache to be initialized")
	}
}

func TestService_Cache(t *testing.T) {
	t.Parallel()

	svc := NewService(&mockConnectionHandler{})
	cache := svc.Cache()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache != svc.cache {
		t.Error("expected cache to match service cache")
	}
}

func TestInfo_Fields(t *testing.T) {
	t.Parallel()

	info := Info{
		Current:     "1.0.0",
		Available:   "1.1.0",
		Beta:        "1.2.0-beta",
		HasUpdate:   true,
		DeviceModel: "SNSW-002P16EU",
		DeviceID:    "shellyplus2pm-123456",
		Generation:  2,
		Platform:    "shelly",
	}

	if info.Current != "1.0.0" {
		t.Errorf("got Current=%q, want %q", info.Current, "1.0.0")
	}
	if info.Available != "1.1.0" {
		t.Errorf("got Available=%q, want %q", info.Available, "1.1.0")
	}
	if info.Beta != "1.2.0-beta" {
		t.Errorf("got Beta=%q, want %q", info.Beta, "1.2.0-beta")
	}
	if !info.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
	if info.DeviceModel != "SNSW-002P16EU" {
		t.Errorf("got DeviceModel=%q, want %q", info.DeviceModel, "SNSW-002P16EU")
	}
	if info.DeviceID != "shellyplus2pm-123456" {
		t.Errorf("got DeviceID=%q, want %q", info.DeviceID, "shellyplus2pm-123456")
	}
	if info.Generation != 2 {
		t.Errorf("got Generation=%d, want 2", info.Generation)
	}
	if info.Platform != "shelly" {
		t.Errorf("got Platform=%q, want %q", info.Platform, "shelly")
	}
}

func TestStatus_Fields(t *testing.T) {
	t.Parallel()

	status := Status{
		Status:      "idle",
		HasUpdate:   true,
		NewVersion:  "1.1.0",
		Progress:    0,
		CanRollback: true,
	}

	if status.Status != "idle" {
		t.Errorf("got Status=%q, want %q", status.Status, "idle")
	}
	if !status.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
	if status.NewVersion != "1.1.0" {
		t.Errorf("got NewVersion=%q, want %q", status.NewVersion, "1.1.0")
	}
	if status.Progress != 0 {
		t.Errorf("got Progress=%d, want 0", status.Progress)
	}
	if !status.CanRollback {
		t.Error("expected CanRollback to be true")
	}
}

func TestCheckResult_Fields(t *testing.T) {
	t.Parallel()

	info := &Info{Current: "1.0.0", Available: "1.1.0", HasUpdate: true}

	result := CheckResult{
		Name: testDeviceNameDevice1,
		Info: info,
		Err:  nil,
	}

	if result.Name != testDeviceNameDevice1 {
		t.Errorf("got Name=%q, want %q", result.Name, testDeviceNameDevice1)
	}
	if result.Info != info {
		t.Error("expected Info to match")
	}
	if result.Err != nil {
		t.Errorf("expected Err to be nil, got %v", result.Err)
	}
}

func TestDeviceUpdateStatus_Fields(t *testing.T) {
	t.Parallel()

	info := &Info{Current: "1.0.0", Available: "1.1.0", HasUpdate: true}

	status := DeviceUpdateStatus{
		Name:      testDeviceNameDevice1,
		Info:      info,
		HasUpdate: true,
	}

	if status.Name != testDeviceNameDevice1 {
		t.Errorf("got Name=%q, want %q", status.Name, testDeviceNameDevice1)
	}
	if status.Info != info {
		t.Error("expected Info to match")
	}
	if !status.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
}

func TestUpdateOpts_Fields(t *testing.T) {
	t.Parallel()

	opts := UpdateOpts{
		Beta:        true,
		CustomURL:   "http://example.com/firmware.zip",
		Parallelism: 5,
	}

	if !opts.Beta {
		t.Error("expected Beta to be true")
	}
	if opts.CustomURL != "http://example.com/firmware.zip" {
		t.Errorf("got CustomURL=%q, want %q", opts.CustomURL, "http://example.com/firmware.zip")
	}
	if opts.Parallelism != 5 {
		t.Errorf("got Parallelism=%d, want 5", opts.Parallelism)
	}
}

func TestUpdateResult_Fields(t *testing.T) {
	t.Parallel()

	result := UpdateResult{
		Name:    testDeviceNameDevice1,
		Success: true,
		Err:     nil,
	}

	if result.Name != testDeviceNameDevice1 {
		t.Errorf("got Name=%q, want %q", result.Name, testDeviceNameDevice1)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Err != nil {
		t.Errorf("expected Err to be nil, got %v", result.Err)
	}
}

func TestUpdateEntry_Fields(t *testing.T) {
	t.Parallel()

	info := &Info{Current: "1.0.0", Available: "1.1.0", Beta: "1.2.0-beta", HasUpdate: true}
	device := model.Device{Address: "192.168.1.101", Model: "SNSW-002P16EU"}

	entry := UpdateEntry{
		Name:      testDeviceNameDevice1,
		Device:    device,
		FwInfo:    info,
		HasUpdate: true,
		HasBeta:   true,
		Error:     nil,
	}

	if entry.Name != testDeviceNameDevice1 {
		t.Errorf("got Name=%q, want %q", entry.Name, testDeviceNameDevice1)
	}
	if entry.Device.Address != device.Address {
		t.Errorf("got Device.Address=%q, want %q", entry.Device.Address, device.Address)
	}
	if entry.FwInfo != info {
		t.Error("expected FwInfo to match")
	}
	if !entry.HasUpdate {
		t.Error("expected HasUpdate to be true")
	}
	if !entry.HasBeta {
		t.Error("expected HasBeta to be true")
	}
}

func TestNewCache(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.entries == nil {
		t.Error("expected entries map to be initialized")
	}
}

func TestCache_GetSet(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	// Get non-existent entry
	entry, ok := cache.Get("device1")
	if ok {
		t.Error("expected entry not to exist")
	}
	if entry != nil {
		t.Error("expected nil entry")
	}

	// Set entry
	newEntry := &CacheEntry{
		DeviceName:  "device1",
		Address:     "192.168.1.101",
		Info:        &Info{Current: "1.0.0"},
		LastChecked: time.Now(),
	}
	cache.Set("device1", newEntry)

	// Get existing entry
	entry, ok = cache.Get("device1")
	if !ok {
		t.Error("expected entry to exist")
	}
	if entry != newEntry {
		t.Error("expected entry to match")
	}
}

func TestCache_All(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	// Empty cache
	entries := cache.All()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	// Add entries
	cache.Set("device1", &CacheEntry{DeviceName: "device1"})
	cache.Set("device2", &CacheEntry{DeviceName: "device2"})

	entries = cache.All()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestCache_AllSorted(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	// Add entries with updates
	cache.Set("zebra", &CacheEntry{DeviceName: "zebra", Info: &Info{HasUpdate: false}})
	cache.Set("alpha", &CacheEntry{DeviceName: "alpha", Info: &Info{HasUpdate: true}})
	cache.Set(testDeviceNameBeta, &CacheEntry{DeviceName: testDeviceNameBeta, Info: &Info{HasUpdate: false}})
	cache.Set("gamma", &CacheEntry{DeviceName: "gamma", Info: &Info{HasUpdate: true}})

	entries := cache.AllSorted()
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Entries with updates should come first
	if !entries[0].Info.HasUpdate || !entries[1].Info.HasUpdate {
		t.Error("expected first two entries to have updates")
	}
	if entries[2].Info.HasUpdate || entries[3].Info.HasUpdate {
		t.Error("expected last two entries to not have updates")
	}

	// Alphabetical within groups
	if entries[0].DeviceName != "alpha" || entries[1].DeviceName != "gamma" {
		t.Error("expected alphabetical sort within update group")
	}
	if entries[2].DeviceName != testDeviceNameBeta || entries[3].DeviceName != "zebra" {
		t.Error("expected alphabetical sort within no-update group")
	}
}

func TestCache_DevicesWithUpdates(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	cache.Set("device1", &CacheEntry{DeviceName: "device1", Info: &Info{HasUpdate: true}})
	cache.Set("device2", &CacheEntry{DeviceName: "device2", Info: &Info{HasUpdate: false}})
	cache.Set("device3", &CacheEntry{DeviceName: "device3", Info: &Info{HasUpdate: true}})
	cache.Set("device4", &CacheEntry{DeviceName: "device4", Info: nil}) // nil info

	withUpdates := cache.DevicesWithUpdates()
	if len(withUpdates) != 2 {
		t.Errorf("expected 2 devices with updates, got %d", len(withUpdates))
	}
}

func TestCache_UpdateCount(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	if cache.UpdateCount() != 0 {
		t.Errorf("expected 0 update count, got %d", cache.UpdateCount())
	}

	cache.Set("device1", &CacheEntry{DeviceName: "device1", Info: &Info{HasUpdate: true}})
	cache.Set("device2", &CacheEntry{DeviceName: "device2", Info: &Info{HasUpdate: false}})

	if cache.UpdateCount() != 1 {
		t.Errorf("expected 1 update count, got %d", cache.UpdateCount())
	}
}

func TestCache_Clear(t *testing.T) {
	t.Parallel()

	cache := NewCache()

	cache.Set("device1", &CacheEntry{DeviceName: "device1"})
	cache.Set("device2", &CacheEntry{DeviceName: "device2"})

	cache.Clear()

	if len(cache.All()) != 0 {
		t.Error("expected cache to be cleared")
	}
}

func TestCacheEntry_IsStale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		lastChecked time.Time
		maxAge      time.Duration
		want        bool
	}{
		{
			name:        "fresh entry",
			lastChecked: time.Now(),
			maxAge:      5 * time.Minute,
			want:        false,
		},
		{
			name:        "stale entry",
			lastChecked: time.Now().Add(-10 * time.Minute),
			maxAge:      5 * time.Minute,
			want:        true,
		},
		{
			name:        "just under max age",
			lastChecked: time.Now().Add(-4*time.Minute - 59*time.Second),
			maxAge:      5 * time.Minute,
			want:        false, // < maxAge is not stale
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entry := &CacheEntry{LastChecked: tt.lastChecked}
			got := entry.IsStale(tt.maxAge)
			if got != tt.want {
				t.Errorf("IsStale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildUpdateList(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"device1": {Address: "192.168.1.101", Model: "SNSW-002P16EU"},
		"device2": {Address: "192.168.1.102", Model: "SHSW-1"},
		"device3": {Address: "192.168.1.103", Model: "SNSW-002P16EU"},
	}

	results := []CheckResult{
		{Name: "device1", Info: &Info{Current: "1.0.0", Available: "1.1.0", HasUpdate: true}},
		{Name: "device2", Info: &Info{Current: "1.0.0", Available: "", HasUpdate: false}},
		{Name: "device3", Info: &Info{Current: "1.0.0", Available: "1.1.0", Beta: "1.2.0-beta", HasUpdate: true}},
	}

	entries := BuildUpdateList(results, devices)

	if len(entries) != 2 {
		t.Errorf("expected 2 entries with updates, got %d", len(entries))
	}

	// Should be sorted by name
	if entries[0].Name != "device1" {
		t.Errorf("expected first entry to be device1, got %s", entries[0].Name)
	}
	if entries[1].Name != "device3" {
		t.Errorf("expected second entry to be device3, got %s", entries[1].Name)
	}
}

func TestFilterDevicesByNameAndPlatform(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"device1": {Address: "192.168.1.101", Platform: "shelly"},
		"device2": {Address: "192.168.1.102", Platform: "tasmota"},
		"device3": {Address: "192.168.1.103", Platform: "shelly"},
		"device4": {Address: "192.168.1.104", Platform: ""}, // empty defaults to shelly
	}

	tests := []struct {
		name      string
		devices   string
		platform  string
		wantCount int
	}{
		{
			name:      "no filters",
			wantCount: 4,
		},
		{
			name:      "filter by shelly platform",
			platform:  "shelly",
			wantCount: 3, // device1, device3, device4
		},
		{
			name:      "filter by tasmota platform",
			platform:  "tasmota",
			wantCount: 1,
		},
		{
			name:      "filter by device list",
			devices:   "device1,device2",
			wantCount: 2,
		},
		{
			name:      "filter by both",
			devices:   "device1,device2,device3",
			platform:  "shelly",
			wantCount: 2, // device1, device3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FilterDevicesByNameAndPlatform(devices, tt.devices, tt.platform)
			if len(result) != tt.wantCount {
				t.Errorf("got %d devices, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestFilterEntriesByStage(t *testing.T) {
	t.Parallel()

	entries := []UpdateEntry{
		{Name: "device1", HasUpdate: true, HasBeta: false},
		{Name: "device2", HasUpdate: true, HasBeta: true},
		{Name: "device3", HasUpdate: false, HasBeta: true},
		{Name: "device4", HasUpdate: false, HasBeta: false},
	}

	// Filter for stable - gets entries with HasUpdate=true
	stableEntries := FilterEntriesByStage(entries, false)
	if len(stableEntries) != 2 { // device1, device2 have stable updates
		t.Errorf("expected 2 stable entries, got %d", len(stableEntries))
	}

	// Filter for beta - gets entries with HasBeta=true OR HasUpdate=true
	betaEntries := FilterEntriesByStage(entries, true)
	if len(betaEntries) != 3 { // device1 (HasUpdate), device2 (HasBeta), device3 (HasBeta)
		t.Errorf("expected 3 beta entries, got %d", len(betaEntries))
	}
}

func TestAnyHasBeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []UpdateEntry
		want    bool
	}{
		{
			name:    "empty list",
			entries: []UpdateEntry{},
			want:    false,
		},
		{
			name: "no beta",
			entries: []UpdateEntry{
				{Name: "device1", HasBeta: false},
				{Name: "device2", HasBeta: false},
			},
			want: false,
		},
		{
			name: "has beta",
			entries: []UpdateEntry{
				{Name: "device1", HasBeta: false},
				{Name: "device2", HasBeta: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := AnyHasBeta(tt.entries)
			if got != tt.want {
				t.Errorf("AnyHasBeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectEntriesByStage(t *testing.T) {
	t.Parallel()

	entries := []UpdateEntry{
		{Name: testDeviceNameDevice1, HasUpdate: true, HasBeta: false},
		{Name: "device2", HasUpdate: true, HasBeta: true},
		{Name: "device3", HasUpdate: false, HasBeta: true},
	}

	// Select stable - gets entries with HasUpdate=true
	indices, stage := SelectEntriesByStage(entries, false)
	if stage != testStageStable {
		t.Errorf("got stage=%q, want %q", stage, testStageStable)
	}
	if len(indices) != 2 {
		t.Errorf("expected 2 indices for stable, got %d", len(indices))
	}

	// Select beta - gets entries with HasBeta=true OR HasUpdate=true
	indices, stage = SelectEntriesByStage(entries, true)
	if stage != testDeviceNameBeta {
		t.Errorf("got stage=%q, want %q", stage, testDeviceNameBeta)
	}
	if len(indices) != 3 {
		t.Errorf("expected 3 indices for beta, got %d", len(indices))
	}
}

func TestGetEntriesByIndices(t *testing.T) {
	t.Parallel()

	entries := []UpdateEntry{
		{Name: "device0"},
		{Name: "device1"},
		{Name: "device2"},
		{Name: "device3"},
	}

	result := GetEntriesByIndices(entries, []int{0, 2})
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Name != "device0" || result[1].Name != "device2" {
		t.Error("unexpected entries returned")
	}

	// Test with out of bounds index
	result = GetEntriesByIndices(entries, []int{0, 10})
	if len(result) != 1 {
		t.Errorf("expected 1 entry (ignoring out of bounds), got %d", len(result))
	}
}

func TestToDeviceUpdateStatuses(t *testing.T) {
	t.Parallel()

	entries := []UpdateEntry{
		{Name: "device1", FwInfo: &Info{Current: "1.0.0"}, HasUpdate: true},
		{Name: "device2", FwInfo: &Info{Current: "1.1.0"}, HasUpdate: false},
	}

	statuses := ToDeviceUpdateStatuses(entries)

	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].Name != "device1" || !statuses[0].HasUpdate {
		t.Error("first status incorrect")
	}
	if statuses[1].Name != "device2" || statuses[1].HasUpdate {
		t.Error("second status incorrect")
	}
}
