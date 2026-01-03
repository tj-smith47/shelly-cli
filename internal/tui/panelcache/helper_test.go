package panelcache

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cache"
)

const testDevice = "device1"

func TestLoadWithCacheNilCache(t *testing.T) {
	t.Parallel()

	cmd := LoadWithCache(nil, testDevice, cache.TypeSystem)
	msg := cmd()

	missMsg, ok := msg.(CacheMissMsg)
	if !ok {
		t.Fatalf("Expected CacheMissMsg, got %T", msg)
	}
	if missMsg.Device != testDevice {
		t.Errorf("Device = %q, want %q", missMsg.Device, testDevice)
	}
	if missMsg.DataType != cache.TypeSystem {
		t.Errorf("DataType = %q, want %q", missMsg.DataType, cache.TypeSystem)
	}
}

func TestLoadWithCacheMiss(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	cmd := LoadWithCache(fc, testDevice, cache.TypeSystem)
	msg := cmd()

	_, ok := msg.(CacheMissMsg)
	if !ok {
		t.Fatalf("Expected CacheMissMsg, got %T", msg)
	}
}

func TestLoadWithCacheHit(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some data.
	testData := map[string]string{"name": "Test Device"}
	if err := fc.Set(testDevice, cache.TypeSystem, testData, time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cmd := LoadWithCache(fc, testDevice, cache.TypeSystem)
	msg := cmd()

	hitMsg, ok := msg.(CacheHitMsg)
	if !ok {
		t.Fatalf("Expected CacheHitMsg, got %T", msg)
	}
	if hitMsg.Device != testDevice {
		t.Errorf("Device = %q, want %q", hitMsg.Device, testDevice)
	}
	if hitMsg.NeedsRefresh {
		t.Error("NeedsRefresh should be false for fresh cache")
	}
}

func TestLoadWithCacheNeedsRefresh(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set data with expired TTL (1 nanosecond).
	testData := map[string]string{"name": "Test Device"}
	if err := fc.Set(testDevice, cache.TypeSystem, testData, time.Nanosecond); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Wait for expiration.
	time.Sleep(2 * time.Millisecond)

	cmd := LoadWithCache(fc, testDevice, cache.TypeSystem)
	msg := cmd()

	hitMsg, ok := msg.(CacheHitMsg)
	if !ok {
		t.Fatalf("Expected CacheHitMsg (expired data), got %T", msg)
	}
	if !hitMsg.NeedsRefresh {
		t.Error("NeedsRefresh should be true for expired cache")
	}
}

func TestBackgroundRefreshSuccess(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	testData := map[string]string{"name": "Refreshed"}
	fetch := func() (any, error) {
		return testData, nil
	}

	cmd := BackgroundRefresh(fc, testDevice, cache.TypeSystem, time.Hour, fetch)
	msg := cmd()

	refreshMsg, ok := msg.(RefreshCompleteMsg)
	if !ok {
		t.Fatalf("Expected RefreshCompleteMsg, got %T", msg)
	}
	if refreshMsg.Err != nil {
		t.Errorf("Err = %v, want nil", refreshMsg.Err)
	}
	if refreshMsg.Device != testDevice {
		t.Errorf("Device = %q, want %q", refreshMsg.Device, testDevice)
	}

	// Verify data was cached.
	entry, err := fc.Get(testDevice, cache.TypeSystem)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if entry == nil {
		t.Fatal("Cache entry should exist")
	}
}

func TestBackgroundRefreshError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("fetch failed")
	fetch := func() (any, error) {
		return nil, testErr
	}

	cmd := BackgroundRefresh(nil, testDevice, cache.TypeSystem, time.Hour, fetch)
	msg := cmd()

	refreshMsg, ok := msg.(RefreshCompleteMsg)
	if !ok {
		t.Fatalf("Expected RefreshCompleteMsg, got %T", msg)
	}
	if !errors.Is(refreshMsg.Err, testErr) {
		t.Errorf("Err = %v, want %v", refreshMsg.Err, testErr)
	}
}

func TestInvalidate(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some data.
	if err := fc.Set(testDevice, cache.TypeSystem, "data", time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cmd := Invalidate(fc, testDevice, cache.TypeSystem)
	_ = cmd() // Execute.

	// Verify data was removed.
	entry, err := fc.Get(testDevice, cache.TypeSystem)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if entry != nil {
		t.Error("Cache entry should be nil after invalidation")
	}
}

func TestInvalidateDevice(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set multiple types for same device.
	if err := fc.Set(testDevice, cache.TypeSystem, "sys", time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set(testDevice, cache.TypeWiFi, "wifi", time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cmd := InvalidateDevice(fc, testDevice)
	_ = cmd() // Execute.

	// Verify all data was removed.
	entry1, err := fc.Get(testDevice, cache.TypeSystem)
	if err != nil {
		t.Logf("warning: get system cache: %v", err)
	}
	entry2, err := fc.Get(testDevice, cache.TypeWiFi)
	if err != nil {
		t.Logf("warning: get wifi cache: %v", err)
	}
	if entry1 != nil || entry2 != nil {
		t.Error("All cache entries should be nil after device invalidation")
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	type TestData struct {
		Name string `json:"name"`
	}

	rawData := json.RawMessage(`{"name":"Test"}`)
	result, err := Unmarshal[TestData](rawData)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if result.Name != "Test" {
		t.Errorf("Name = %q, want %q", result.Name, "Test")
	}
}

func TestHelper(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	h := NewHelper(fc, testDevice, cache.TypeSystem, time.Hour)

	if h.Device() != testDevice {
		t.Errorf("Device() = %q, want %q", h.Device(), testDevice)
	}
	if h.DataType() != cache.TypeSystem {
		t.Errorf("DataType() = %q, want %q", h.DataType(), cache.TypeSystem)
	}

	// Test SetDevice.
	h = h.SetDevice("device2")
	if h.Device() != "device2" {
		t.Errorf("Device() after SetDevice = %q, want %q", h.Device(), "device2")
	}
}

func TestHelperLoad(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	h := NewHelper(fc, testDevice, cache.TypeSystem, time.Hour)
	cmd := h.Load()
	msg := cmd()

	_, ok := msg.(CacheMissMsg)
	if !ok {
		t.Fatalf("Expected CacheMissMsg, got %T", msg)
	}
}

func TestHelperRefresh(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	h := NewHelper(fc, testDevice, cache.TypeSystem, time.Hour)
	fetch := func() (any, error) {
		return "test", nil
	}

	cmd := h.Refresh(fetch)
	msg := cmd()

	refreshMsg, ok := msg.(RefreshCompleteMsg)
	if !ok {
		t.Fatalf("Expected RefreshCompleteMsg, got %T", msg)
	}
	if refreshMsg.Err != nil {
		t.Errorf("Err = %v, want nil", refreshMsg.Err)
	}
}

func TestHelperInvalidate(t *testing.T) {
	t.Parallel()

	afs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/cache", afs)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some data.
	if err := fc.Set(testDevice, cache.TypeSystem, "data", time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	h := NewHelper(fc, testDevice, cache.TypeSystem, time.Hour)
	cmd := h.Invalidate()
	_ = cmd()

	// Verify data was removed.
	entry, err := fc.Get(testDevice, cache.TypeSystem)
	if err != nil {
		t.Logf("warning: get cache: %v", err)
	}
	if entry != nil {
		t.Error("Cache entry should be nil after invalidation")
	}
}

func TestInvalidateNilCache(t *testing.T) {
	t.Parallel()

	cmd := Invalidate(nil, testDevice, cache.TypeSystem)
	msg := cmd()
	if msg != nil {
		t.Errorf("Expected nil msg, got %T", msg)
	}
}

func TestInvalidateDeviceNilCache(t *testing.T) {
	t.Parallel()

	cmd := InvalidateDevice(nil, testDevice)
	msg := cmd()
	if msg != nil {
		t.Errorf("Expected nil msg, got %T", msg)
	}
}
