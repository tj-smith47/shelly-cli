package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestFileCache_GetSet(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Test cache miss
	entry, err := cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("Get() expected nil for cache miss")
	}

	// Test Set and Get
	testData := map[string]string{"version": "1.0.0", "available": "1.0.1"}
	if err := cache.Set("device1", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	entry, err = cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry == nil {
		t.Fatal("Get() returned nil, expected entry")
	}

	if entry.Device != "device1" {
		t.Errorf("Device = %q, want %q", entry.Device, "device1")
	}
	if entry.DataType != TypeFirmware {
		t.Errorf("DataType = %q, want %q", entry.DataType, TypeFirmware)
	}
	if entry.Version != CurrentVersion {
		t.Errorf("Version = %d, want %d", entry.Version, CurrentVersion)
	}

	// Unmarshal data
	var result map[string]string
	if err := entry.Unmarshal(&result); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if result["version"] != "1.0.0" {
		t.Errorf("data[version] = %q, want %q", result["version"], "1.0.0")
	}
}

func TestFileCache_Expiration(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set with very short TTL
	testData := map[string]string{"test": "data"}
	if err := cache.Set("device1", TypeSystem, testData, time.Millisecond); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Get should return nil for expired entry
	entry, err := cache.Get("device1", TypeSystem)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("Get() expected nil for expired entry")
	}

	// GetWithExpired should still return the entry
	entry, err = cache.GetWithExpired("device1", TypeSystem)
	if err != nil {
		t.Fatalf("GetWithExpired() error: %v", err)
	}
	if entry == nil {
		t.Fatal("GetWithExpired() returned nil, expected expired entry")
	}
	if !entry.IsExpired() {
		t.Error("IsExpired() = false, want true")
	}
}

func TestFileCache_Invalidate(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set some data
	testData := map[string]string{"test": "data"}
	if err := cache.Set("device1", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Verify it exists
	entry, err := cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry == nil {
		t.Fatal("Get() returned nil, expected entry")
	}

	// Invalidate
	if err := cache.Invalidate("device1", TypeFirmware); err != nil {
		t.Fatalf("Invalidate() error: %v", err)
	}

	// Verify it's gone
	entry, err = cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("Get() expected nil after invalidation")
	}

	// Invalidate non-existent should not error
	if err := cache.Invalidate("device1", TypeFirmware); err != nil {
		t.Errorf("Invalidate() non-existent error: %v", err)
	}
}

func TestFileCache_InvalidateDevice(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set multiple entries for same device
	testData := map[string]string{"test": "data"}
	if err := cache.Set("device1", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("device1", TypeSystem, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("device2", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Invalidate device1
	if err := cache.InvalidateDevice("device1"); err != nil {
		t.Fatalf("InvalidateDevice() error: %v", err)
	}

	// Verify device1 entries are gone
	entry, err := cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("device1 firmware entry should be gone")
	}

	entry, err = cache.Get("device1", TypeSystem)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("device1 system entry should be gone")
	}

	// Verify device2 entry still exists
	entry, err = cache.Get("device2", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry == nil {
		t.Error("device2 entry should still exist")
	}
}

func TestFileCache_InvalidateAll(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set multiple entries
	testData := map[string]string{"test": "data"}
	if err := cache.Set("device1", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("device2", TypeSystem, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Invalidate all
	if err := cache.InvalidateAll(); err != nil {
		t.Fatalf("InvalidateAll() error: %v", err)
	}

	// Verify all entries are gone
	entry, err := cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("device1 entry should be gone")
	}

	entry, err = cache.Get("device2", TypeSystem)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("device2 entry should be gone")
	}
}

func TestFileCache_Cleanup(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set one expired and one valid entry
	testData := map[string]string{"test": "data"}
	if err := cache.Set("expired", TypeFirmware, testData, time.Millisecond); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("valid", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Wait for first to expire
	time.Sleep(5 * time.Millisecond)

	// Cleanup
	removed, err := cache.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup() error: %v", err)
	}
	if removed != 1 {
		t.Errorf("Cleanup() removed = %d, want 1", removed)
	}

	// Expired entry should be gone (even with GetWithExpired)
	entry, err := cache.GetWithExpired("expired", TypeFirmware)
	if err != nil {
		t.Fatalf("GetWithExpired() error: %v", err)
	}
	if entry != nil {
		t.Error("expired entry should be removed by Cleanup")
	}

	// Valid entry should still exist
	entry, err = cache.Get("valid", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry == nil {
		t.Error("valid entry should still exist")
	}
}

func TestFileCache_Stats(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Set some entries
	testData := map[string]string{"test": "data"}
	if err := cache.Set("device1", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("device1", TypeSystem, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	if err := cache.Set("device2", TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("TotalEntries = %d, want 3", stats.TotalEntries)
	}
	if stats.DeviceCount != 2 {
		t.Errorf("DeviceCount = %d, want 2", stats.DeviceCount)
	}
	if stats.ExpiredEntries != 0 {
		t.Errorf("ExpiredEntries = %d, want 0", stats.ExpiredEntries)
	}
	if stats.TypeCounts[TypeFirmware] != 2 {
		t.Errorf("TypeCounts[firmware] = %d, want 2", stats.TypeCounts[TypeFirmware])
	}
	if stats.TypeCounts[TypeSystem] != 1 {
		t.Errorf("TypeCounts[system] = %d, want 1", stats.TypeCounts[TypeSystem])
	}
}

func TestFileCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	var wg sync.WaitGroup
	testData := map[string]string{"test": "data"}

	// Concurrent writes
	for i := range 10 {
		wg.Go(func() {
			device := "device" + string(rune('0'+i))
			if err := cache.Set(device, TypeFirmware, testData, time.Hour); err != nil {
				t.Errorf("Set() error: %v", err)
			}
		})
	}

	// Concurrent reads
	for range 10 {
		wg.Go(func() {
			_, err := cache.Get("device0", TypeFirmware)
			if err != nil {
				t.Errorf("Get() error: %v", err)
			}
		})
	}

	wg.Wait()

	// Verify final state
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}
	if stats.TotalEntries != 10 {
		t.Errorf("TotalEntries = %d, want 10", stats.TotalEntries)
	}
}

func TestFileCache_SetWithID(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	testData := map[string]string{"test": "data"}
	deviceID := "shellyplus1pm-abc123"
	if err := cache.SetWithID("kitchen", deviceID, TypeFirmware, testData, time.Hour); err != nil {
		t.Fatalf("SetWithID() error: %v", err)
	}

	entry, err := cache.Get("kitchen", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry == nil {
		t.Fatal("Get() returned nil")
	}
	if entry.DeviceID != deviceID {
		t.Errorf("DeviceID = %q, want %q", entry.DeviceID, deviceID)
	}
}

func TestFileCache_CorruptFile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	// Write corrupt JSON directly
	path := cache.entryPath("device1", TypeFirmware)
	if err := fs.MkdirAll("/cache/firmware", 0o755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	if err := afero.WriteFile(fs, path, []byte("not valid json{"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Get should handle corrupt file gracefully
	entry, err := cache.Get("device1", TypeFirmware)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if entry != nil {
		t.Error("Get() expected nil for corrupt file")
	}

	// Corrupt file should be removed
	exists, err := afero.Exists(fs, path)
	if err != nil {
		t.Fatalf("Exists() error: %v", err)
	}
	if exists {
		t.Error("corrupt file should be removed")
	}
}

func TestEntry_NeedsRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		agePercent float64 // percentage of TTL elapsed
		want       bool
	}{
		{"fresh (10%)", 0.10, false},
		{"fresh (49%)", 0.49, false},
		{"stale (51%)", 0.51, true},
		{"stale (90%)", 0.90, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ttl := time.Hour
			age := time.Duration(float64(ttl) * tt.agePercent)

			entry := &Entry{
				CachedAt:  time.Now().Add(-age),
				ExpiresAt: time.Now().Add(ttl - age),
			}

			if got := entry.NeedsRefresh(); got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"with-dash", "with-dash"},
		{"with_underscore", "with_underscore"},
		{"192.168.1.1", "192.168.1.1"},
		{"path/to/file", "path_to_file"},
		{"file:name", "file_name"},
		{"file*name?", "file_name_"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := sanitizeFilename(tt.input); got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFileCache_Path(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	cache, err := NewWithFs("/test/cache", fs)
	if err != nil {
		t.Fatalf("NewWithFs() error: %v", err)
	}

	if cache.Path() != "/test/cache" {
		t.Errorf("Path() = %q, want %q", cache.Path(), "/test/cache")
	}
}
