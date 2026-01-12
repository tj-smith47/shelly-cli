// Package cache provides file-based caching for device data.
// This cache is shared between TUI and CLI for instant panel switching
// and faster command execution with offline capability.
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CurrentVersion is the cache format version.
// Increment this when the cache format changes to invalidate old caches.
const CurrentVersion = 1

// Common data type constants for cache keys.
const (
	TypeDeviceInfo = "deviceinfo"
	TypeComponents = "components"
	TypeSystem     = "system"
	TypeWiFi       = "wifi"
	TypeSecurity   = "security"
	TypeCloud      = "cloud"
	TypeBLE        = "ble"
	TypeMQTT       = "protocols/mqtt"
	TypeModbus     = "protocols/modbus"
	TypeEthernet   = "protocols/ethernet"
	TypeMatter     = "smarthome/matter"
	TypeZigbee     = "smarthome/zigbee"
	TypeLoRa       = "smarthome/lora"
	TypeZWave      = "smarthome/zwave"
	TypeFirmware   = "firmware"
	TypeSchedules  = "automation/schedules"
	TypeWebhooks   = "automation/webhooks"
	TypeVirtuals   = "automation/virtuals"
	TypeInputs     = "automation/inputs"
	TypeKVS        = "automation/kvs"
	TypeScripts    = "automation/scripts"
)

// Common TTL values based on data volatility.
const (
	TTLDeviceInfo = 24 * time.Hour // Hardware info rarely changes.
	TTLComponents = 24 * time.Hour // Component list rarely changes.
	TTLSystem     = 1 * time.Hour  // System settings change occasionally.
	TTLWiFi       = 30 * time.Minute
	TTLSecurity   = 1 * time.Hour
	TTLCloud      = 30 * time.Minute
	TTLBLE        = 1 * time.Hour
	TTLProtocols  = 1 * time.Hour    // MQTT, Modbus, Ethernet.
	TTLSmartHome  = 30 * time.Minute // Matter, Zigbee, LoRa, Z-Wave.
	TTLFirmware   = 1 * time.Hour
	TTLAutomation = 5 * time.Minute // Schedules, Webhooks, Virtuals, KVS, Scripts.
	TTLInputs     = 10 * time.Minute
)

// Entry represents a cached data entry with metadata.
type Entry struct {
	Version   int             `json:"version"`
	Device    string          `json:"device"`
	DeviceID  string          `json:"device_id,omitempty"`
	DataType  string          `json:"data_type"`
	CachedAt  time.Time       `json:"cached_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	Data      json.RawMessage `json:"data"`
}

// IsExpired returns true if the entry has expired.
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Age returns the duration since the entry was cached.
func (e *Entry) Age() time.Duration {
	return time.Since(e.CachedAt)
}

// TTL returns the total time-to-live duration for this entry.
func (e *Entry) TTL() time.Duration {
	return e.ExpiresAt.Sub(e.CachedAt)
}

// NeedsRefresh returns true if the entry is past 50% of its TTL.
// This is used to trigger background refresh while serving cached data.
func (e *Entry) NeedsRefresh() bool {
	return e.Age() > e.TTL()/2
}

// Unmarshal unmarshals the entry's data into the provided value.
func (e *Entry) Unmarshal(v any) error {
	return json.Unmarshal(e.Data, v)
}

// Stats contains cache statistics for monitoring.
type Stats struct {
	TotalEntries   int            `json:"total_entries"`
	TotalSize      int64          `json:"total_size_bytes"`
	ExpiredEntries int            `json:"expired_entries"`
	DeviceCount    int            `json:"device_count"`
	OldestEntry    time.Time      `json:"oldest_entry"`
	NewestEntry    time.Time      `json:"newest_entry"`
	TypeCounts     map[string]int `json:"type_counts"`
}

// FileCache provides file-based caching for device data.
// It uses atomic writes and JSON validation for data integrity.
type FileCache struct {
	basePath string
	afs      afero.Fs
	mu       sync.RWMutex
}

// New creates a FileCache using the default cache directory.
// The directory is created if it doesn't exist.
func New() (*FileCache, error) {
	cacheDir, err := config.CacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}
	return NewWithPath(cacheDir)
}

// NewWithPath creates a FileCache with a custom path.
// The directory is created if it doesn't exist.
// Uses the package-level filesystem from config.Fs().
func NewWithPath(basePath string) (*FileCache, error) {
	afs := config.Fs()
	if err := afs.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &FileCache{basePath: basePath, afs: afs}, nil
}

// NewWithFs creates a FileCache with a custom filesystem.
// This is useful for testing with an in-memory filesystem.
func NewWithFs(basePath string, afs afero.Fs) (*FileCache, error) {
	if err := afs.MkdirAll(basePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &FileCache{basePath: basePath, afs: afs}, nil
}

// Path returns the base cache directory path.
func (c *FileCache) Path() string {
	return c.basePath
}

// Get retrieves cached data if fresh, returns nil if expired or missing.
// An error is returned only for file system errors (not cache miss or expiration).
func (c *FileCache) Get(device, dataType string) (*Entry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.entryPath(device, dataType)
	data, err := afero.ReadFile(c.afs, path)
	if err != nil {
		if isNotExist(err) {
			return nil, nil //nolint:nilnil // nil,nil is the expected cache miss response
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Corrupt cache file - remove it and return nil
		iostreams.DebugErr("remove corrupt cache file", c.afs.Remove(path))
		return nil, nil //nolint:nilnil // corrupt file treated as cache miss
	}

	// Validate version
	if entry.Version != CurrentVersion {
		iostreams.DebugErr("remove old version cache file", c.afs.Remove(path))
		return nil, nil //nolint:nilnil // old version treated as cache miss
	}

	// Check expiration
	if entry.IsExpired() {
		return nil, nil //nolint:nilnil // expired entry treated as cache miss
	}

	return &entry, nil
}

// GetWithExpired retrieves cached data even if expired.
// This is useful for displaying stale data while refreshing.
func (c *FileCache) GetWithExpired(device, dataType string) (*Entry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.entryPath(device, dataType)
	data, err := afero.ReadFile(c.afs, path)
	if err != nil {
		if isNotExist(err) {
			return nil, nil //nolint:nilnil // cache miss
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		iostreams.DebugErr("remove corrupt cache file", c.afs.Remove(path))
		return nil, nil //nolint:nilnil // corrupt file treated as cache miss
	}

	if entry.Version != CurrentVersion {
		iostreams.DebugErr("remove old version cache file", c.afs.Remove(path))
		return nil, nil //nolint:nilnil // old version treated as cache miss
	}

	return &entry, nil
}

// Set stores data with the specified TTL.
func (c *FileCache) Set(device, dataType string, data any, ttl time.Duration) error {
	return c.SetWithID(device, "", dataType, data, ttl)
}

// SetWithID stores data with device ID for cache key stability.
// The device ID helps identify cached data when device names change.
func (c *FileCache) SetWithID(device, deviceID, dataType string, data any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Marshal the data
	rawData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	now := time.Now()
	entry := Entry{
		Version:   CurrentVersion,
		Device:    device,
		DeviceID:  deviceID,
		DataType:  dataType,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
		Data:      rawData,
	}

	entryData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	path := c.entryPath(device, dataType)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := c.afs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := path + ".tmp"
	if err := afero.WriteFile(c.afs, tmpPath, entryData, 0o644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := c.afs.Rename(tmpPath, path); err != nil {
		iostreams.DebugErr("remove temp cache file", c.afs.Remove(tmpPath))
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// Invalidate removes cached data for a device/type.
func (c *FileCache) Invalidate(device, dataType string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.entryPath(device, dataType)
	if err := c.afs.Remove(path); err != nil && !isNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	return nil
}

// InvalidateDevice removes all cached data for a device.
func (c *FileCache) InvalidateDevice(device string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Walk all cache directories and remove files matching the device
	return c.walkEntries(func(path string, entry *Entry) error {
		if entry.Device == device {
			if err := c.afs.Remove(path); err != nil && !isNotExist(err) {
				return err
			}
		}
		return nil
	})
}

// InvalidateAll removes all cached data.
func (c *FileCache) InvalidateAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all files but keep the base directory
	entries, err := afero.ReadDir(c.afs, c.basePath)
	if err != nil {
		if isNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(c.basePath, entry.Name())
		if err := c.afs.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// Cleanup removes expired entries.
// Returns the number of entries removed.
func (c *FileCache) Cleanup() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	err := c.walkEntries(func(path string, entry *Entry) error {
		if entry.IsExpired() {
			if err := c.afs.Remove(path); err != nil && !isNotExist(err) {
				return err
			}
			removed++
		}
		return nil
	})

	return removed, err
}

// Stats returns cache statistics.
func (c *FileCache) Stats() (*Stats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &Stats{
		TypeCounts: make(map[string]int),
	}
	devices := make(map[string]struct{})

	err := c.walkEntries(func(path string, entry *Entry) error {
		info, statErr := c.afs.Stat(path)
		if statErr != nil {
			return nil //nolint:nilerr // skip files we can't stat
		}

		stats.TotalEntries++
		stats.TotalSize += info.Size()
		stats.TypeCounts[entry.DataType]++
		devices[entry.Device] = struct{}{}

		if entry.IsExpired() {
			stats.ExpiredEntries++
		}

		if stats.OldestEntry.IsZero() || entry.CachedAt.Before(stats.OldestEntry) {
			stats.OldestEntry = entry.CachedAt
		}
		if entry.CachedAt.After(stats.NewestEntry) {
			stats.NewestEntry = entry.CachedAt
		}

		return nil
	})

	stats.DeviceCount = len(devices)

	return stats, err
}

// entryPath returns the file path for a cache entry.
// Format: basePath/dataType/device.json.
func (c *FileCache) entryPath(device, dataType string) string {
	// Sanitize device name to be filesystem-safe
	safeDevice := sanitizeFilename(device)
	return filepath.Join(c.basePath, dataType, safeDevice+".json")
}

// walkEntries iterates over all cache entries.
func (c *FileCache) walkEntries(fn func(path string, entry *Entry) error) error {
	return afero.Walk(c.afs, c.basePath, func(path string, d os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil //nolint:nilerr // skip walk errors
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		base := filepath.Base(path)
		// Skip meta.json (metadata file, not a cache entry)
		if base == "meta.json" {
			return nil
		}
		// Skip temp files (*.tmp.json)
		if len(base) >= 9 && base[len(base)-9:] == ".tmp.json" {
			return nil
		}

		data, readErr := afero.ReadFile(c.afs, path)
		if readErr != nil {
			return nil //nolint:nilerr // skip unreadable files
		}

		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil //nolint:nilerr // skip corrupt files
		}

		return fn(path, &entry)
	})
}

// sanitizeFilename makes a string safe for use as a filename.
func sanitizeFilename(s string) string {
	// Replace problematic characters with underscores
	result := make([]byte, len(s))
	for i := range len(s) {
		ch := s[i]
		switch ch {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			result[i] = '_'
		default:
			result[i] = ch
		}
	}
	return string(result)
}

// isNotExist checks if the error indicates file not found.
func isNotExist(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, iofs.ErrNotExist)
}

// Meta holds cache metadata stored in meta.json.
type Meta struct {
	Version     int       `json:"version"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// metaPath returns the path to the meta.json file.
func (c *FileCache) metaPath() string {
	return filepath.Join(c.basePath, "meta.json")
}

// ReadMeta reads the cache metadata from meta.json.
func (c *FileCache) ReadMeta() (*Meta, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := afero.ReadFile(c.afs, c.metaPath())
	if err != nil {
		if isNotExist(err) {
			return &Meta{Version: CurrentVersion}, nil
		}
		return nil, fmt.Errorf("failed to read meta file: %w", err)
	}

	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return &Meta{Version: CurrentVersion}, nil
	}

	return &meta, nil
}

// WriteMeta writes the cache metadata to meta.json.
func (c *FileCache) WriteMeta(meta *Meta) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	if err := afero.WriteFile(c.afs, c.metaPath(), data, 0o644); err != nil {
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	return nil
}

// CleanupIfNeeded runs cleanup only if the specified interval has passed since the last cleanup.
// Returns the number of entries removed, or 0 if cleanup was skipped.
func (c *FileCache) CleanupIfNeeded(interval time.Duration) (int, error) {
	meta, err := c.ReadMeta()
	if err != nil {
		iostreams.DebugErr("read cache meta", err)
		// Continue with cleanup anyway
		meta = &Meta{Version: CurrentVersion}
	}

	// Check if cleanup is needed
	if time.Since(meta.LastCleanup) < interval {
		return 0, nil // Cleanup not needed
	}

	// Run cleanup
	removed, err := c.Cleanup()
	if err != nil {
		return removed, err
	}

	// Update last cleanup time
	meta.LastCleanup = time.Now()
	if writeErr := c.WriteMeta(meta); writeErr != nil {
		iostreams.DebugErr("write cache meta", writeErr)
	}

	return removed, nil
}
