// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// FirmwareCache provides in-memory caching of firmware info for all devices.
// This is NOT persisted to config - it's populated at startup and refreshed on demand.
type FirmwareCache struct {
	mu      sync.RWMutex
	entries map[string]*FirmwareCacheEntry // keyed by device name
}

// FirmwareCacheEntry holds cached firmware info for a single device.
type FirmwareCacheEntry struct {
	DeviceName  string
	Address     string
	Info        *FirmwareInfo
	LastChecked time.Time
	Error       error
}

// NewFirmwareCache creates a new firmware cache.
func NewFirmwareCache() *FirmwareCache {
	return &FirmwareCache{
		entries: make(map[string]*FirmwareCacheEntry),
	}
}

// Get returns cached firmware info for a device.
func (c *FirmwareCache) Get(deviceName string) (*FirmwareCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[deviceName]
	return entry, ok
}

// Set stores firmware info for a device.
func (c *FirmwareCache) Set(deviceName string, entry *FirmwareCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[deviceName] = entry
}

// All returns all cached entries.
func (c *FirmwareCache) All() []*FirmwareCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*FirmwareCacheEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		result = append(result, entry)
	}
	return result
}

// AllSorted returns all cached entries, sorted with updates first.
func (c *FirmwareCache) AllSorted() []*FirmwareCacheEntry {
	entries := c.All()
	sort.Slice(entries, func(i, j int) bool {
		// Updates first
		iHasUpdate := entries[i].Info != nil && entries[i].Info.HasUpdate
		jHasUpdate := entries[j].Info != nil && entries[j].Info.HasUpdate
		if iHasUpdate != jHasUpdate {
			return iHasUpdate // true sorts before false
		}
		// Then by name
		return entries[i].DeviceName < entries[j].DeviceName
	})
	return entries
}

// DevicesWithUpdates returns devices that have available updates.
func (c *FirmwareCache) DevicesWithUpdates() []*FirmwareCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*FirmwareCacheEntry
	for _, entry := range c.entries {
		if entry.Info != nil && entry.Info.HasUpdate {
			result = append(result, entry)
		}
	}
	return result
}

// UpdateCount returns the number of devices with available updates.
func (c *FirmwareCache) UpdateCount() int {
	return len(c.DevicesWithUpdates())
}

// Clear removes all cached entries.
func (c *FirmwareCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*FirmwareCacheEntry)
}

// IsStale returns true if the cache entry is older than the given duration.
func (e *FirmwareCacheEntry) IsStale(maxAge time.Duration) bool {
	return time.Since(e.LastChecked) > maxAge
}

// PrefetchFirmwareCache populates the firmware cache for all registered devices.
// This runs concurrently and respects the global rate limit.
func (s *Service) PrefetchFirmwareCache(ctx context.Context, ios *iostreams.IOStreams) {
	devices := config.ListDevices()
	if len(devices) == 0 {
		return
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for name, device := range devices {
		deviceName := name
		dev := device
		g.Go(func() error {
			s.prefetchDeviceFirmware(gctx, deviceName, dev)
			return nil // Don't fail the whole group on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("firmware prefetch", err)
	}
}

// prefetchDeviceFirmware fetches and caches firmware info for a single device.
func (s *Service) prefetchDeviceFirmware(ctx context.Context, name string, device model.Device) {
	entry := &FirmwareCacheEntry{
		DeviceName:  name,
		Address:     device.Address,
		LastChecked: time.Now(),
	}

	info, err := s.CheckDeviceFirmware(ctx, device)
	if err != nil {
		entry.Error = err
		iostreams.DebugErrCat(iostreams.CategoryDevice, "prefetch firmware "+name, err)
	} else {
		entry.Info = info
	}

	s.firmwareCache.Set(name, entry)
}

// GetCachedFirmware returns cached firmware info, fetching if not cached or stale.
// Returns the cached entry or nil if the device is not found.
func (s *Service) GetCachedFirmware(ctx context.Context, deviceName string, maxAge time.Duration) *FirmwareCacheEntry {
	// Check cache first
	if entry, ok := s.firmwareCache.Get(deviceName); ok {
		if !entry.IsStale(maxAge) {
			return entry
		}
	}

	// Fetch fresh data
	device, ok := config.GetDevice(deviceName)
	if !ok {
		return nil
	}

	s.prefetchDeviceFirmware(ctx, deviceName, device)
	entry, _ := s.firmwareCache.Get(deviceName)
	return entry
}

// FirmwareCache returns the firmware cache.
func (s *Service) FirmwareCache() *FirmwareCache {
	return s.firmwareCache
}
