// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ErrOfflineCacheMiss is returned when --offline is set but no cached data exists.
var ErrOfflineCacheMiss = errors.New("no cached data available (run without --offline to fetch from device)")

// ErrFlagConflict is returned when --offline and --refresh are both set.
var ErrFlagConflict = errors.New("cannot use --offline and --refresh together")

// CacheResult wraps a result with cache metadata.
type CacheResult[T any] struct {
	Data      T
	FromCache bool
	CachedAt  time.Time
}

// CachedFetch checks cache first, fetches if miss/expired, and caches the result.
// It respects --refresh (bypass cache) and --offline (cache only) flags.
//
// Returns:
//   - CacheResult with data and cache metadata
//   - Error if fetch fails or --offline with cache miss
//
// In verbose mode, logs "(cached)" when serving from cache.
func CachedFetch[T any](
	ctx context.Context,
	f *Factory,
	device string,
	dataType string,
	ttl time.Duration,
	fetch func(ctx context.Context) (T, error),
) (CacheResult[T], error) {
	var zero CacheResult[T]

	refresh := viper.GetBool("refresh")
	offline := viper.GetBool("offline")

	// Check for flag conflict
	if refresh && offline {
		return zero, ErrFlagConflict
	}

	fc := f.FileCache()
	ios := f.IOStreams()

	// Try cache first (unless --refresh)
	if fc != nil && !refresh {
		if entry, err := fc.Get(device, dataType); err == nil && entry != nil {
			var data T
			if err := entry.Unmarshal(&data); err == nil {
				logCacheHit(ios, device, dataType, entry.CachedAt)
				return CacheResult[T]{
					Data:      data,
					FromCache: true,
					CachedAt:  entry.CachedAt,
				}, nil
			}
			// Unmarshal failed - treat as cache miss
			ios.DebugErr("unmarshal cache "+dataType, err)
		}
	}

	// Offline mode with cache miss
	if offline {
		return zero, ErrOfflineCacheMiss
	}

	// Fetch fresh data
	data, err := fetch(ctx)
	if err != nil {
		return zero, err
	}

	// Cache the result
	now := time.Now()
	if fc != nil {
		if err := fc.Set(device, dataType, data, ttl); err != nil {
			ios.DebugErr("cache "+dataType, err)
		}
	}

	return CacheResult[T]{
		Data:      data,
		FromCache: false,
		CachedAt:  now,
	}, nil
}

// CachedFetchList is like CachedFetch but for list operations.
// Convenience wrapper that handles slice types.
func CachedFetchList[T any](
	ctx context.Context,
	f *Factory,
	device string,
	dataType string,
	ttl time.Duration,
	fetch func(ctx context.Context) ([]T, error),
) (CacheResult[[]T], error) {
	return CachedFetch(ctx, f, device, dataType, ttl, fetch)
}

// InvalidateCache removes cached data for a device/type after a mutation.
// Errors are logged to debug (non-fatal) since cache invalidation is best-effort.
func InvalidateCache(f *Factory, device, dataType string) {
	fc := f.FileCache()
	if fc == nil {
		return
	}

	if err := fc.Invalidate(device, dataType); err != nil {
		f.IOStreams().DebugErr("invalidate cache "+dataType, err)
	}
}

// InvalidateDeviceCache removes all cached data for a device.
// Useful after device-wide operations like firmware updates.
func InvalidateDeviceCache(f *Factory, device string) {
	fc := f.FileCache()
	if fc == nil {
		return
	}

	if err := fc.InvalidateDevice(device); err != nil {
		f.IOStreams().DebugErr("invalidate device cache", err)
	}
}

// logCacheHit logs a verbose message when serving data from cache.
func logCacheHit(ios *iostreams.IOStreams, device, dataType string, cachedAt time.Time) {
	if !iostreams.IsVerbose() {
		return
	}

	age := time.Since(cachedAt)
	var ageStr string
	switch {
	case age < time.Minute:
		ageStr = "just now"
	case age < time.Hour:
		ageStr = fmt.Sprintf("%dm ago", int(age.Minutes()))
	default:
		ageStr = fmt.Sprintf("%dh ago", int(age.Hours()))
	}

	ios.Debug("cache hit for %s/%s (cached %s)", device, dataType, ageStr)
}

// CheckCacheFlags validates that --refresh and --offline aren't both set.
// Call this early in commands to provide a clear error message.
func CheckCacheFlags() error {
	if viper.GetBool("refresh") && viper.GetBool("offline") {
		return ErrFlagConflict
	}
	return nil
}

// IsCacheEnabled returns true if caching should be used (--refresh not set).
func IsCacheEnabled() bool {
	return !viper.GetBool("refresh")
}

// IsOfflineMode returns true if --offline flag is set.
func IsOfflineMode() bool {
	return viper.GetBool("offline")
}

// CacheTypeForResource returns the cache type constant for a resource name.
// This is useful for mutation commands that need to invalidate cache.
func CacheTypeForResource(resource string) string {
	switch resource {
	case "schedule", "schedules":
		return cache.TypeSchedules
	case "webhook", "webhooks":
		return cache.TypeWebhooks
	case "virtual", "virtuals":
		return cache.TypeVirtuals
	case "kvs":
		return cache.TypeKVS
	case "script", "scripts":
		return cache.TypeScripts
	case "input", "inputs":
		return cache.TypeInputs
	case "firmware":
		return cache.TypeFirmware
	case "wifi":
		return cache.TypeWiFi
	case "cloud":
		return cache.TypeCloud
	case "mqtt":
		return cache.TypeMQTT
	case "ble":
		return cache.TypeBLE
	case "security":
		return cache.TypeSecurity
	case "system":
		return cache.TypeSystem
	default:
		return ""
	}
}

// CachedComponentList returns the component list for a device, using cache if available.
// If --refresh flag is set or cache is empty, fetches fresh data and caches it.
// This is useful for commands that need the component list and want to avoid
// repeated RPC calls for data that rarely changes.
func CachedComponentList(ctx context.Context, f *Factory, device string) ([]model.Component, error) {
	fc := f.FileCache()
	svc := f.ShellyService()
	ios := f.IOStreams()

	// Try cache first unless --refresh flag is set
	if fc != nil && !viper.GetBool("refresh") {
		if entry, err := fc.Get(device, cache.TypeComponents); err == nil && entry != nil {
			var components []model.Component
			if err := entry.Unmarshal(&components); err == nil {
				logCacheHit(ios, device, cache.TypeComponents, entry.CachedAt)
				return components, nil
			}
		}
	}

	// Offline mode with cache miss
	if viper.GetBool("offline") {
		return nil, ErrOfflineCacheMiss
	}

	// Fetch fresh component list
	var components []model.Component
	err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
		var err error
		components, err = conn.ListComponents(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Cache the result (24 hour TTL for components)
	if fc != nil && components != nil {
		if err := fc.Set(device, cache.TypeComponents, components, cache.TTLComponents); err != nil {
			ios.DebugErr("cache components", err)
		}
	}

	return components, nil
}
