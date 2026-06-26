// Package firmware provides firmware management for Shelly devices.
package firmware

import (
	"context"
	"errors"
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

// mockConnectionHandler is a test double for ConnectionHandler. By default it
// reports Gen2 (isGen1Fn nil) so existing Gen2 tests are unaffected; a test
// exercising the Gen1 path sets isGen1Fn and gen1ConnectionFn.
type mockConnectionHandler struct {
	withConnectionFn func(ctx context.Context, identifier string, fn func(*client.Client) error) error
	gen1ConnectionFn func(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error
	isGen1Fn         func(ctx context.Context, identifier string) (bool, model.Device, error)
}

func (m *mockConnectionHandler) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	if m.withConnectionFn != nil {
		return m.withConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionHandler) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	if m.gen1ConnectionFn != nil {
		return m.gen1ConnectionFn(ctx, identifier, fn)
	}
	return nil
}

func (m *mockConnectionHandler) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	if m.isGen1Fn != nil {
		return m.isGen1Fn(ctx, identifier)
	}
	return false, model.Device{}, nil
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

// gen1Handler builds a mock that reports the device as Gen1 and records which
// connection path the service took.
func gen1Handler(gen2Called, gen1Called *bool) *mockConnectionHandler {
	return &mockConnectionHandler{
		isGen1Fn: func(context.Context, string) (bool, model.Device, error) {
			return true, model.Device{Name: "gen1dev", Generation: 1}, nil
		},
		withConnectionFn: func(context.Context, string, func(*client.Client) error) error {
			*gen2Called = true
			return nil
		},
		gen1ConnectionFn: func(context.Context, string, func(*client.Gen1Client) error) error {
			*gen1Called = true
			return nil
		},
	}
}

// TestService_Gen1_RoutesToGen1Path proves the B6 fix: firmware reads on a Gen1
// device take the Gen1 HTTP path, never the Gen2 RPC path (which cannot connect
// to a Gen1 device and previously failed every firmware command).
func TestService_Gen1_RoutesToGen1Path(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Check", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		if _, err := svc.Check(ctx, "gen1dev"); err != nil {
			t.Fatalf("Check: %v", err)
		}
		if gen2 || !gen1 {
			t.Errorf("Gen1 Check must use the Gen1 path (gen1=%v gen2=%v)", gen1, gen2)
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		if _, err := svc.GetStatus(ctx, "gen1dev"); err != nil {
			t.Fatalf("GetStatus: %v", err)
		}
		if gen2 || !gen1 {
			t.Errorf("Gen1 GetStatus must use the Gen1 path (gen1=%v gen2=%v)", gen1, gen2)
		}
	})

	t.Run("UpdateStable", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		if err := svc.UpdateStable(ctx, "gen1dev"); err != nil {
			t.Fatalf("UpdateStable: %v", err)
		}
		if gen2 || !gen1 {
			t.Errorf("Gen1 stable update must use the Gen1 path (gen1=%v gen2=%v)", gen1, gen2)
		}
	})

	t.Run("UpdateFromURL", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		if err := svc.UpdateFromURL(ctx, "gen1dev", "http://example/fw.zip"); err != nil {
			t.Fatalf("UpdateFromURL: %v", err)
		}
		if gen2 || !gen1 {
			t.Errorf("Gen1 custom-URL update must use the Gen1 path (gen1=%v gen2=%v)", gen1, gen2)
		}
	})
}

// TestService_Gen1_UnsupportedOps proves Gen1-incapable operations fail with an
// honest ErrGen1Unsupported instead of a misleading connection error or a
// silent wrong-channel flash.
func TestService_Gen1_UnsupportedOps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Rollback", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		err := svc.Rollback(ctx, "gen1dev")
		if !errors.Is(err, ErrGen1Unsupported) {
			t.Fatalf("Gen1 rollback should report ErrGen1Unsupported, got %v", err)
		}
		if gen1 || gen2 {
			t.Error("unsupported op must not open any connection")
		}
	})

	t.Run("GetURL", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		_, err := svc.GetURL(ctx, "gen1dev", testStageStable)
		if !errors.Is(err, ErrGen1Unsupported) {
			t.Fatalf("Gen1 GetURL should report ErrGen1Unsupported, got %v", err)
		}
	})

	t.Run("UpdateBeta", func(t *testing.T) {
		t.Parallel()
		var gen2, gen1 bool
		svc := NewService(gen1Handler(&gen2, &gen1))
		err := svc.UpdateBeta(ctx, "gen1dev")
		if !errors.Is(err, ErrGen1Unsupported) {
			t.Fatalf("Gen1 beta update should report ErrGen1Unsupported, got %v", err)
		}
		if gen1 {
			t.Error("a rejected beta update must not flash anything on the Gen1 device")
		}
	})
}

// TestService_Gen2_StillUsesRPCPath guards against a regression: a Gen2 device
// must continue to take the RPC path.
func TestService_Gen2_StillUsesRPCPath(t *testing.T) {
	t.Parallel()
	var gen2, gen1 bool
	handler := &mockConnectionHandler{
		isGen1Fn: func(context.Context, string) (bool, model.Device, error) {
			return false, model.Device{Name: "gen2dev", Generation: 2}, nil
		},
		withConnectionFn: func(context.Context, string, func(*client.Client) error) error {
			gen2 = true
			return nil
		},
		gen1ConnectionFn: func(context.Context, string, func(*client.Gen1Client) error) error {
			gen1 = true
			return nil
		},
	}
	svc := NewService(handler)
	if _, err := svc.Check(context.Background(), "gen2dev"); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if gen1 || !gen2 {
		t.Errorf("Gen2 Check must use the RPC path (gen1=%v gen2=%v)", gen1, gen2)
	}
}

// TestService_Gen1Detection_ErrorPropagates ensures a generation-detection
// failure aborts the firmware op rather than silently falling through to a path.
func TestService_Gen1Detection_ErrorPropagates(t *testing.T) {
	t.Parallel()
	detectErr := errors.New("cannot reach device")
	handler := &mockConnectionHandler{
		isGen1Fn: func(context.Context, string) (bool, model.Device, error) {
			return false, model.Device{}, detectErr
		},
	}
	svc := NewService(handler)
	ctx := context.Background()

	if _, err := svc.Check(ctx, "dev"); !errors.Is(err, detectErr) {
		t.Errorf("Check should propagate the detection error, got %v", err)
	}
	if _, err := svc.GetStatus(ctx, "dev"); !errors.Is(err, detectErr) {
		t.Errorf("GetStatus should propagate the detection error, got %v", err)
	}
	if err := svc.UpdateStable(ctx, "dev"); !errors.Is(err, detectErr) {
		t.Errorf("Update should propagate the detection error, got %v", err)
	}
	if err := svc.Rollback(ctx, "dev"); !errors.Is(err, detectErr) {
		t.Errorf("Rollback should propagate the detection error, got %v", err)
	}
	if _, err := svc.GetURL(ctx, "dev", testStageStable); !errors.Is(err, detectErr) {
		t.Errorf("GetURL should propagate the detection error, got %v", err)
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
		Platform:    platformShelly,
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
	if info.Platform != platformShelly {
		t.Errorf("got Platform=%q, want %q", info.Platform, platformShelly)
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

	// Get existing entry — returns an isolated copy, not the stored pointer.
	entry, ok = cache.Get("device1")
	if !ok {
		t.Error("expected entry to exist")
	}
	if entry == newEntry {
		t.Error("expected Get to return a copy, not the stored pointer")
	}
	if entry.DeviceName != newEntry.DeviceName || entry.Address != newEntry.Address ||
		entry.Info.Current != newEntry.Info.Current {
		t.Errorf("returned entry %+v does not match stored %+v", entry, newEntry)
	}

	// Mutating the returned copy must not affect the cached entry.
	entry.Info.Current = "9.9.9"
	again, _ := cache.Get("device1")
	if again.Info.Current != "1.0.0" {
		t.Errorf("mutating the returned entry leaked into the cache: %q", again.Info.Current)
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
		"device1": {Address: "192.168.1.101", Platform: platformShelly},
		"device2": {Address: "192.168.1.102", Platform: "tasmota"},
		"device3": {Address: "192.168.1.103", Platform: platformShelly},
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
			platform:  platformShelly,
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
			platform:  platformShelly,
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

	// Filter for beta - ONLY entries that actually have a beta image. A
	// stable-only device (device1: HasUpdate, no beta) must NOT be selected,
	// or the beta channel gets flashed onto a device with no beta available.
	betaEntries := FilterEntriesByStage(entries, true)
	if len(betaEntries) != 2 { // device2 (HasBeta), device3 (HasBeta)
		t.Errorf("expected 2 beta entries, got %d", len(betaEntries))
	}
	for _, e := range betaEntries {
		if !e.HasBeta {
			t.Errorf("beta filter selected %q which has no beta image", e.Name)
		}
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

	// Select beta - only entries that actually have a beta image (device2,
	// device3). device1 is stable-only and must be excluded.
	indices, stage = SelectEntriesByStage(entries, true)
	if stage != testDeviceNameBeta {
		t.Errorf("got stage=%q, want %q", stage, testDeviceNameBeta)
	}
	if len(indices) != 2 {
		t.Errorf("expected 2 indices for beta, got %d", len(indices))
	}
	for _, idx := range indices {
		if !entries[idx].HasBeta {
			t.Errorf("beta selection included %q which has no beta image", entries[idx].Name)
		}
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
