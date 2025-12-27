// Package firmware provides firmware management for Shelly devices.
package firmware

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

// CacheEntry holds cached firmware info for a single device.
type CacheEntry struct {
	DeviceName  string
	Address     string
	Info        *Info
	LastChecked time.Time
	Error       error
}

// IsStale returns true if the cache entry is older than the given duration.
func (e *CacheEntry) IsStale(maxAge time.Duration) bool {
	return time.Since(e.LastChecked) > maxAge
}

// Cache provides in-memory caching of firmware info for all devices.
// This is NOT persisted to config - it's populated at startup and refreshed on demand.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry // keyed by device name
}

// NewCache creates a new firmware cache.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get returns cached firmware info for a device.
func (c *Cache) Get(deviceName string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[deviceName]
	return entry, ok
}

// Set stores firmware info for a device.
func (c *Cache) Set(deviceName string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[deviceName] = entry
}

// All returns all cached entries.
func (c *Cache) All() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*CacheEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		result = append(result, entry)
	}
	return result
}

// AllSorted returns all cached entries, sorted with updates first.
func (c *Cache) AllSorted() []*CacheEntry {
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
func (c *Cache) DevicesWithUpdates() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*CacheEntry
	for _, entry := range c.entries {
		if entry.Info != nil && entry.Info.HasUpdate {
			result = append(result, entry)
		}
	}
	return result
}

// UpdateCount returns the number of devices with available updates.
func (c *Cache) UpdateCount() int {
	return len(c.DevicesWithUpdates())
}

// Clear removes all cached entries.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Prefetch populates the firmware cache for all registered devices.
// This runs concurrently and respects the global rate limit.
func (s *Service) Prefetch(ctx context.Context, ios *iostreams.IOStreams) {
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
			s.prefetchDevice(gctx, deviceName, dev)
			return nil // Don't fail the whole group on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("firmware prefetch", err)
	}
}

// prefetchDevice fetches and caches firmware info for a single device.
func (s *Service) prefetchDevice(ctx context.Context, name string, device model.Device) {
	entry := &CacheEntry{
		DeviceName:  name,
		Address:     device.Address,
		LastChecked: time.Now(),
	}

	info, err := s.CheckDevice(ctx, device)
	if err != nil {
		entry.Error = err
		iostreams.DebugErrCat(iostreams.CategoryDevice, "prefetch firmware "+name, err)
	} else {
		entry.Info = info
	}

	s.cache.Set(name, entry)
}

// GetCached returns cached firmware info, fetching if not cached or stale.
// Returns the cached entry or nil if the device is not found.
func (s *Service) GetCached(ctx context.Context, deviceName string, maxAge time.Duration) *CacheEntry {
	// Check cache first
	if entry, ok := s.cache.Get(deviceName); ok {
		if !entry.IsStale(maxAge) {
			return entry
		}
	}

	// Fetch fresh data
	device, ok := config.GetDevice(deviceName)
	if !ok {
		return nil
	}

	s.prefetchDevice(ctx, deviceName, device)
	entry, _ := s.cache.Get(deviceName)
	return entry
}
