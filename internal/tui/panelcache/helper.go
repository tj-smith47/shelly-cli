// Package panelcache provides cache integration helpers for TUI panels.
// It enables panels to display cached data instantly while refreshing
// in the background when data is stale.
package panelcache

import (
	"encoding/json"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// FetchFunc is a function that fetches data from a device.
// It should return the data to cache and any error that occurred.
type FetchFunc func() (any, error)

// CacheHitMsg indicates data was found in cache.
// Panels should display this data immediately.
type CacheHitMsg struct {
	Device       string
	DataType     string
	Data         json.RawMessage
	CachedAt     time.Time
	NeedsRefresh bool // True if cache is > 50% of TTL
}

// CacheMissMsg indicates no valid cache exists.
// Panels should show loading state.
type CacheMissMsg struct {
	Device   string
	DataType string
}

// RefreshCompleteMsg indicates a background refresh completed.
// Panels should update their display with new data.
type RefreshCompleteMsg struct {
	Device   string
	DataType string
	Data     any
	CachedAt time.Time
	Err      error
}

// RefreshStartedMsg indicates a background refresh has started.
type RefreshStartedMsg struct {
	Device   string
	DataType string
}

// LoadWithCache returns a command that checks the cache and triggers
// background refresh if needed. The flow is:
//
//  1. Check cache for valid entry
//  2. If hit: return CacheHitMsg with data and NeedsRefresh flag
//  3. If miss: return CacheMissMsg, then fetch and return RefreshCompleteMsg
//  4. If hit with NeedsRefresh: also trigger background refresh
//
// Usage in panel Update:
//
//	case CacheHitMsg:
//	    m.data = unmarshalData(msg.Data)
//	    m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt)
//	    if msg.NeedsRefresh {
//	        return m, BackgroundRefresh(...)
//	    }
//	case CacheMissMsg:
//	    m.loading = true
//	    return m, FetchFresh(...)
//	case RefreshCompleteMsg:
//	    m.data = msg.Data
//	    m.cacheStatus = m.cacheStatus.SetUpdatedAt(msg.CachedAt).SetRefreshing(false)
func LoadWithCache(fc *cache.FileCache, device, dataType string) tea.Cmd {
	return func() tea.Msg {
		if fc == nil {
			return CacheMissMsg{Device: device, DataType: dataType}
		}

		entry, err := fc.GetWithExpired(device, dataType)
		if err != nil || entry == nil {
			return CacheMissMsg{Device: device, DataType: dataType}
		}

		return CacheHitMsg{
			Device:       device,
			DataType:     dataType,
			Data:         entry.Data,
			CachedAt:     entry.CachedAt,
			NeedsRefresh: entry.NeedsRefresh() || entry.IsExpired(),
		}
	}
}

// BackgroundRefresh returns a command that fetches fresh data and caches it.
// Use this after receiving a CacheHitMsg with NeedsRefresh=true.
func BackgroundRefresh(fc *cache.FileCache, device, dataType string, ttl time.Duration, fetch FetchFunc) tea.Cmd {
	return func() tea.Msg {
		data, err := fetch()
		if err != nil {
			return RefreshCompleteMsg{
				Device:   device,
				DataType: dataType,
				Err:      err,
			}
		}

		// Cache the result
		if fc != nil {
			if err := fc.Set(device, dataType, data, ttl); err != nil {
				iostreams.DebugErr("cache set", err)
			}
		}

		return RefreshCompleteMsg{
			Device:   device,
			DataType: dataType,
			Data:     data,
			CachedAt: time.Now(),
		}
	}
}

// FetchAndCache returns a command that fetches data and caches it.
// Use this after receiving a CacheMissMsg.
func FetchAndCache(fc *cache.FileCache, device, dataType string, ttl time.Duration, fetch FetchFunc) tea.Cmd {
	return BackgroundRefresh(fc, device, dataType, ttl, fetch)
}

// Invalidate returns a command that invalidates cache for a device/type.
// Use this after modifying data (create/update/delete).
func Invalidate(fc *cache.FileCache, device, dataType string) tea.Cmd {
	return func() tea.Msg {
		if fc != nil {
			if err := fc.Invalidate(device, dataType); err != nil {
				iostreams.DebugErr("cache invalidate", err)
			}
		}
		return nil
	}
}

// InvalidateDevice returns a command that invalidates all cache for a device.
func InvalidateDevice(fc *cache.FileCache, device string) tea.Cmd {
	return func() tea.Msg {
		if fc != nil {
			if err := fc.InvalidateDevice(device); err != nil {
				iostreams.DebugErr("cache invalidate device", err)
			}
		}
		return nil
	}
}

// Unmarshal is a helper to unmarshal CacheHitMsg data into a typed value.
func Unmarshal[T any](data json.RawMessage) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	return result, err
}

// Helper provides caching functionality for a specific device and data type.
// It simplifies cache operations by storing the device and type context.
type Helper struct {
	fc       *cache.FileCache
	device   string
	dataType string
	ttl      time.Duration
}

// NewHelper creates a new cache helper for a specific device and data type.
func NewHelper(fc *cache.FileCache, device, dataType string, ttl time.Duration) *Helper {
	return &Helper{
		fc:       fc,
		device:   device,
		dataType: dataType,
		ttl:      ttl,
	}
}

// Load returns a command that loads data from cache.
func (h *Helper) Load() tea.Cmd {
	return LoadWithCache(h.fc, h.device, h.dataType)
}

// Refresh returns a command that refreshes data.
func (h *Helper) Refresh(fetch FetchFunc) tea.Cmd {
	return BackgroundRefresh(h.fc, h.device, h.dataType, h.ttl, fetch)
}

// Invalidate returns a command that invalidates the cache.
func (h *Helper) Invalidate() tea.Cmd {
	return Invalidate(h.fc, h.device, h.dataType)
}

// SetDevice updates the device and returns the updated helper.
func (h *Helper) SetDevice(device string) *Helper {
	h.device = device
	return h
}

// Device returns the current device.
func (h *Helper) Device() string {
	return h.device
}

// DataType returns the cache data type.
func (h *Helper) DataType() string {
	return h.dataType
}
