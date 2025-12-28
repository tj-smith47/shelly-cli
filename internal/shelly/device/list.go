// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"
	"sort"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ListFilterOptions holds filter criteria for device lists.
type ListFilterOptions struct {
	Generation int
	DeviceType string
	Platform   string
}

// FilterList filters devices based on criteria and converts to DeviceListItem slice.
// Returns the filtered list and a set of unique platforms found.
func FilterList(devices map[string]model.Device, opts ListFilterOptions) (filtered []model.DeviceListItem, platforms map[string]struct{}) {
	filtered = make([]model.DeviceListItem, 0, len(devices))
	platforms = make(map[string]struct{})

	for name, dev := range devices {
		if !matchesFilters(dev, opts) {
			continue
		}
		devPlatform := dev.GetPlatform()
		platforms[devPlatform] = struct{}{}
		filtered = append(filtered, model.DeviceListItem{
			Name:       name,
			Address:    dev.Address,
			Platform:   devPlatform,
			Model:      dev.Model,
			Type:       dev.Type,
			Generation: dev.Generation,
			Auth:       dev.Auth != nil,
		})
	}

	return filtered, platforms
}

// matchesFilters checks if a device matches all filter criteria.
func matchesFilters(dev model.Device, opts ListFilterOptions) bool {
	if opts.Generation > 0 && dev.Generation != opts.Generation {
		return false
	}
	if opts.DeviceType != "" && dev.Type != opts.DeviceType {
		return false
	}
	if opts.Platform != "" && dev.GetPlatform() != opts.Platform {
		return false
	}
	return true
}

// SortList sorts a device list. If updatesFirst is true, devices with
// available updates are sorted to the top. Within each group, devices are sorted by name.
func SortList(devices []model.DeviceListItem, updatesFirst bool) {
	sort.Slice(devices, func(i, j int) bool {
		if updatesFirst && devices[i].HasUpdate != devices[j].HasUpdate {
			return devices[i].HasUpdate // true sorts before false
		}
		return devices[i].Name < devices[j].Name
	})
}

// PopulateListFirmware fills in firmware version info from the cache.
// Uses a short cache validity period (5 minutes) so it doesn't trigger network calls during list.
func (s *Service) PopulateListFirmware(ctx context.Context, devices []model.DeviceListItem) {
	const cacheMaxAge = 5 * time.Minute
	for i := range devices {
		entry := s.parent.GetCachedFirmware(ctx, devices[i].Name, cacheMaxAge)
		if entry != nil && entry.Info != nil {
			devices[i].CurrentVersion = entry.Info.Current
			devices[i].AvailableVersion = entry.Info.Available
			devices[i].HasUpdate = entry.Info.HasUpdate
		}
	}
}
