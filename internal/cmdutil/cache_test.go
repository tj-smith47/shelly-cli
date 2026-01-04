package cmdutil

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// testFactory creates a Factory with in-memory cache for testing.
func testFactory(t *testing.T) (*Factory, *cache.FileCache) {
	t.Helper()

	fs := afero.NewMemMapFs()
	fc, err := cache.NewWithFs("/test/cache", fs)
	if err != nil {
		t.Fatalf("failed to create test cache: %v", err)
	}

	// Create test IOStreams
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	// Create factory with test components
	f := NewFactory().
		SetIOStreams(ios).
		SetFileCache(fc)

	return f, fc
}

// Note: Tests using viper cannot be parallel because viper is a global singleton.

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_CacheHit(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	f, fc := testFactory(t)

	// Pre-populate cache
	testData := map[string]string{"key": "value"}
	if err := fc.Set("device1", "testtype", testData, time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	fetchCalled := false
	result, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		fetchCalled = true
		return map[string]string{"key": "fresh"}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fetchCalled {
		t.Error("fetch should not be called on cache hit")
	}

	if !result.FromCache {
		t.Error("result should be from cache")
	}

	if result.Data["key"] != "value" {
		t.Errorf("expected cached value, got %s", result.Data["key"])
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_CacheMiss(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	f, _ := testFactory(t)

	fetchCalled := false
	result, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		fetchCalled = true
		return map[string]string{"key": "fresh"}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fetchCalled {
		t.Error("fetch should be called on cache miss")
	}

	if result.FromCache {
		t.Error("result should not be from cache")
	}

	if result.Data["key"] != "fresh" {
		t.Errorf("expected fresh value, got %s", result.Data["key"])
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_RefreshBypass(t *testing.T) {
	viper.Reset()
	viper.Set("refresh", true)
	defer viper.Reset()

	f, fc := testFactory(t)

	// Pre-populate cache
	testData := map[string]string{"key": "cached"}
	if err := fc.Set("device1", "testtype", testData, time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	fetchCalled := false
	result, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		fetchCalled = true
		return map[string]string{"key": "fresh"}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fetchCalled {
		t.Error("fetch should be called when --refresh is set")
	}

	if result.FromCache {
		t.Error("result should not be from cache when --refresh is set")
	}

	if result.Data["key"] != "fresh" {
		t.Errorf("expected fresh value, got %s", result.Data["key"])
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_OfflineCacheHit(t *testing.T) {
	viper.Reset()
	viper.Set("offline", true)
	defer viper.Reset()

	f, fc := testFactory(t)

	// Pre-populate cache
	testData := map[string]string{"key": "cached"}
	if err := fc.Set("device1", "testtype", testData, time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	fetchCalled := false
	result, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		fetchCalled = true
		return map[string]string{"key": "fresh"}, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fetchCalled {
		t.Error("fetch should not be called in offline mode with cache hit")
	}

	if !result.FromCache {
		t.Error("result should be from cache")
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_OfflineCacheMiss(t *testing.T) {
	viper.Reset()
	viper.Set("offline", true)
	defer viper.Reset()

	f, _ := testFactory(t)

	_, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		return map[string]string{"key": "fresh"}, nil
	})

	if !errors.Is(err, ErrOfflineCacheMiss) {
		t.Errorf("expected ErrOfflineCacheMiss, got %v", err)
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_FlagConflict(t *testing.T) {
	viper.Reset()
	viper.Set("refresh", true)
	viper.Set("offline", true)
	defer viper.Reset()

	f, _ := testFactory(t)

	_, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		return map[string]string{"key": "fresh"}, nil
	})

	if !errors.Is(err, ErrFlagConflict) {
		t.Errorf("expected ErrFlagConflict, got %v", err)
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_FetchError(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	f, _ := testFactory(t)

	expectedErr := errors.New("fetch failed")
	_, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		return nil, expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected fetch error, got %v", err)
	}
}

//nolint:paralleltest // uses shared viper state
func TestCachedFetch_CachesResult(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	f, fc := testFactory(t)

	// First call - should fetch and cache
	_, err := CachedFetch(context.Background(), f, "device1", "testtype", time.Hour, func(_ context.Context) (map[string]string, error) {
		return map[string]string{"key": "fetched"}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cached
	entry, err := fc.Get("device1", "testtype")
	if err != nil {
		t.Fatalf("failed to get cached entry: %v", err)
	}
	if entry == nil {
		t.Fatal("expected cached entry")
	}

	var data map[string]string
	if err := entry.Unmarshal(&data); err != nil {
		t.Fatalf("failed to unmarshal cached data: %v", err)
	}
	if data["key"] != "fetched" {
		t.Errorf("expected 'fetched', got %s", data["key"])
	}
}

func TestInvalidateCache(t *testing.T) {
	t.Parallel()

	f, fc := testFactory(t)

	// Pre-populate cache
	if err := fc.Set("device1", "testtype", "data", time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Verify it exists
	entry, _ := fc.Get("device1", "testtype") //nolint:errcheck // checking entry only
	if entry == nil {
		t.Fatal("cache should exist before invalidation")
	}

	// Invalidate
	InvalidateCache(f, "device1", "testtype")

	// Verify it's gone
	entry, _ = fc.Get("device1", "testtype") //nolint:errcheck // checking entry only
	if entry != nil {
		t.Error("cache should be invalidated")
	}
}

func TestInvalidateDeviceCache(t *testing.T) {
	t.Parallel()

	f, fc := testFactory(t)

	// Pre-populate cache with multiple types
	if err := fc.Set("device1", "type1", "data1", time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}
	if err := fc.Set("device1", "type2", "data2", time.Hour); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Invalidate device
	InvalidateDeviceCache(f, "device1")

	// Verify both are gone
	entry1, _ := fc.Get("device1", "type1") //nolint:errcheck // checking entry only
	entry2, _ := fc.Get("device1", "type2") //nolint:errcheck // checking entry only
	if entry1 != nil || entry2 != nil {
		t.Error("all device cache should be invalidated")
	}
}

func TestCacheTypeForResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		resource string
		expected string
	}{
		{"schedule", cache.TypeSchedules},
		{"schedules", cache.TypeSchedules},
		{"webhook", cache.TypeWebhooks},
		{"webhooks", cache.TypeWebhooks},
		{"virtual", cache.TypeVirtuals},
		{"virtuals", cache.TypeVirtuals},
		{"kvs", cache.TypeKVS},
		{"script", cache.TypeScripts},
		{"scripts", cache.TypeScripts},
		{"input", cache.TypeInputs},
		{"inputs", cache.TypeInputs},
		{"firmware", cache.TypeFirmware},
		{"wifi", cache.TypeWiFi},
		{"cloud", cache.TypeCloud},
		{"mqtt", cache.TypeMQTT},
		{"ble", cache.TypeBLE},
		{"security", cache.TypeSecurity},
		{"system", cache.TypeSystem},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			t.Parallel()
			result := CacheTypeForResource(tt.resource)
			if result != tt.expected {
				t.Errorf("CacheTypeForResource(%q) = %q, want %q", tt.resource, result, tt.expected)
			}
		})
	}
}

//nolint:paralleltest // uses shared viper state
func TestCheckCacheFlags(t *testing.T) {
	tests := []struct {
		name    string
		refresh bool
		offline bool
		wantErr bool
	}{
		{"neither", false, false, false},
		{"refresh only", true, false, false},
		{"offline only", false, true, false},
		{"both", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("refresh", tt.refresh)
			viper.Set("offline", tt.offline)
			defer viper.Reset()

			err := CheckCacheFlags()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckCacheFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//nolint:paralleltest // uses shared viper state
func TestIsCacheEnabled(t *testing.T) {
	tests := []struct {
		name    string
		refresh bool
		want    bool
	}{
		{"default", false, true},
		{"refresh set", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("refresh", tt.refresh)
			defer viper.Reset()

			if got := IsCacheEnabled(); got != tt.want {
				t.Errorf("IsCacheEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // uses shared viper state
func TestIsOfflineMode(t *testing.T) {
	tests := []struct {
		name    string
		offline bool
		want    bool
	}{
		{"default", false, false},
		{"offline set", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("offline", tt.offline)
			defer viper.Reset()

			if got := IsOfflineMode(); got != tt.want {
				t.Errorf("IsOfflineMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
