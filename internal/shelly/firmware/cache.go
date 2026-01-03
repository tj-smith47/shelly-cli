// Package firmware provides firmware management for Shelly devices.
package firmware

import (
	"context"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cache"
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

// Cache provides caching of firmware info for all devices.
// It uses both in-memory cache for fast access during a session
// and file-based cache for persistence across sessions.
type Cache struct {
	mu        sync.RWMutex
	entries   map[string]*CacheEntry // keyed by device name (in-memory)
	fileCache *cache.FileCache       // optional file-based persistence
}

// NewCache creates a new firmware cache.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
	}
}

// SetFileCache sets the file cache for persistence.
func (c *Cache) SetFileCache(fc *cache.FileCache) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fileCache = fc
}

// Get returns cached firmware info for a device.
// Checks in-memory cache first, then file cache.
func (c *Cache) Get(deviceName string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check in-memory cache first.
	if entry, ok := c.entries[deviceName]; ok {
		return entry, true
	}

	// Try loading from file cache.
	entry := c.loadFromFileCache(deviceName)
	if entry != nil {
		c.entries[deviceName] = entry
		return entry, true
	}

	return nil, false
}

// loadFromFileCache loads a cache entry from the file cache.
// Returns nil if file cache is not available or entry not found.
func (c *Cache) loadFromFileCache(deviceName string) *CacheEntry {
	if c.fileCache == nil {
		return nil
	}

	fileEntry, err := c.fileCache.Get(deviceName, cache.TypeFirmware)
	if err != nil {
		iostreams.DebugErr("get firmware from file cache", err)
		return nil
	}
	if fileEntry == nil {
		return nil
	}

	var info Info
	if err := fileEntry.Unmarshal(&info); err != nil {
		iostreams.DebugErr("unmarshal firmware cache", err)
		return nil
	}

	return &CacheEntry{
		DeviceName:  deviceName,
		Info:        &info,
		LastChecked: fileEntry.CachedAt,
	}
}

// Set stores firmware info for a device.
// Stores in both in-memory and file cache.
func (c *Cache) Set(deviceName string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store in in-memory cache.
	c.entries[deviceName] = entry

	// Store in file cache if available and no error.
	if c.fileCache != nil && entry.Info != nil && entry.Error == nil {
		if err := c.fileCache.Set(deviceName, cache.TypeFirmware, entry.Info, cache.TTLFirmware); err != nil {
			iostreams.DebugErr("set firmware to file cache", err)
		}
	}
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
