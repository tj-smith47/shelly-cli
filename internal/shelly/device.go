// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/tj-smith47/shelly-go/types"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// DeviceInfo holds extended device information.
type DeviceInfo struct {
	ID         string
	MAC        string
	Model      string
	Generation int
	Firmware   string
	App        string
	AuthEn     bool
	Address    string
}

// DeviceStatus holds device status information.
type DeviceStatus struct {
	Info   *DeviceInfo
	Status map[string]any
}

// DeviceReboot reboots the device.
func (s *Service) DeviceReboot(ctx context.Context, identifier string, delayMS int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Reboot(ctx, delayMS)
	})
}

// DeviceFactoryReset performs a factory reset on the device.
func (s *Service) DeviceFactoryReset(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.FactoryReset(ctx)
	})
}

// DeviceInfo returns information about the device.
func (s *Service) DeviceInfo(ctx context.Context, identifier string) (*DeviceInfo, error) {
	// Resolve first to capture the address
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	var result *DeviceInfo
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		result = &DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: info.Generation,
			Firmware:   info.Firmware,
			App:        info.App,
			AuthEn:     info.AuthEn,
			Address:    dev.Address,
		}
		return nil
	})
	if err == nil {
		refreshDeviceMetadata(identifier, result)
	}
	return result, err
}

// DeviceStatus returns the full status of the device.
func (s *Service) DeviceStatus(ctx context.Context, identifier string) (*DeviceStatus, error) {
	// Resolve first to capture the address
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	var result *DeviceStatus
	err = s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		info := conn.Info()
		status, err := conn.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &DeviceStatus{
			Info: &DeviceInfo{
				ID:         info.ID,
				MAC:        info.MAC,
				Model:      info.Model,
				Generation: info.Generation,
				Firmware:   info.Firmware,
				App:        info.App,
				AuthEn:     info.AuthEn,
				Address:    dev.Address,
			},
			Status: status,
		}
		return nil
	})
	if err == nil && result.Info != nil {
		refreshDeviceMetadata(identifier, result.Info)
	}
	return result, err
}

// DeviceStatusGen1 returns the full status of a Gen1 device.
// The Gen1 status struct is converted to map[string]any via JSON roundtrip.
func (s *Service) DeviceStatusGen1(ctx context.Context, identifier string) (*DeviceStatus, error) {
	// Resolve first to capture the address
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	var result *DeviceStatus
	err = s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		info := conn.Info()
		status, err := conn.GetStatus(ctx)
		if err != nil {
			return err
		}

		// JSON roundtrip to convert *gen1.Status to map[string]any
		statusBytes, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("marshal gen1 status: %w", err)
		}
		var statusMap map[string]any
		if err := json.Unmarshal(statusBytes, &statusMap); err != nil {
			return fmt.Errorf("unmarshal gen1 status: %w", err)
		}

		result = &DeviceStatus{
			Info: &DeviceInfo{
				ID:         info.ID,
				MAC:        info.MAC,
				Model:      info.Model,
				Generation: info.Generation,
				Firmware:   info.Firmware,
				App:        info.App,
				AuthEn:     info.AuthEn,
				Address:    dev.Address,
			},
			Status: statusMap,
		}
		return nil
	})
	if err == nil && result.Info != nil {
		refreshDeviceMetadata(identifier, result.Info)
	}
	return result, err
}

// DeviceStatusAuto returns device status, auto-detecting generation (Gen1 vs Gen2).
// If generation is known from config, it tries that generation first for efficiency.
// Otherwise it tries Gen2 first (more common), then falls back to Gen1 if Gen2 fails.
func (s *Service) DeviceStatusAuto(ctx context.Context, identifier string) (*DeviceStatus, error) {
	// First resolve to check if we have a stored generation
	// Error is intentionally ignored - if resolution fails, we try Gen2 first
	device, err := s.ResolveWithGeneration(ctx, identifier)

	// If we know it's Gen1, try Gen1 first to avoid wasting time on Gen2
	if err == nil && device.Generation == 1 {
		gen1Result, gen1Err := s.DeviceStatusGen1(ctx, identifier)
		if gen1Err == nil {
			return gen1Result, nil
		}
		// Gen1 failed unexpectedly, try Gen2 as fallback
		result, err := s.DeviceStatus(ctx, identifier)
		if err == nil {
			return result, nil
		}
		// Both failed, return Gen1 error since we knew it was Gen1
		return nil, gen1Err
	}

	// Gen2+ or unknown: Try Gen2 first (more common)
	result, err := s.DeviceStatus(ctx, identifier)
	if err == nil {
		return result, nil
	}

	// Gen2 failed, try Gen1
	gen1Result, gen1Err := s.DeviceStatusGen1(ctx, identifier)
	if gen1Err == nil {
		return gen1Result, nil
	}

	// Both failed, return the original Gen2 error (more informative)
	return nil, err
}

// DevicePing checks if the device is reachable by attempting to connect.
// Uses DeviceInfoAuto to support both Gen1 and Gen2 devices.
func (s *Service) DevicePing(ctx context.Context, identifier string) (*DeviceInfo, error) {
	return s.DeviceInfoAuto(ctx, identifier)
}

// DeviceInfoAuto returns device info, auto-detecting generation (Gen1 vs Gen2).
// If generation is known from config, it tries that generation first for efficiency.
// Otherwise it tries Gen2 first (more common), then falls back to Gen1 if Gen2 fails.
// Use this for TUI/cache where we need to handle all device types.
func (s *Service) DeviceInfoAuto(ctx context.Context, identifier string) (*DeviceInfo, error) {
	// First resolve to check if we have a stored generation
	// Error is intentionally ignored - if resolution fails, we try Gen2 first
	device, err := s.ResolveWithGeneration(ctx, identifier)

	// If we know it's Gen1, try Gen1 first to avoid wasting time on Gen2
	if err == nil && device.Generation == 1 {
		gen1Result, gen1Err := s.DeviceInfoGen1(ctx, identifier)
		if gen1Err == nil {
			return gen1Result, nil
		}
		// Gen1 failed unexpectedly, try Gen2 as fallback
		result, err := s.DeviceInfo(ctx, identifier)
		if err == nil {
			return result, nil
		}
		// Both failed, return Gen1 error since we knew it was Gen1
		return nil, gen1Err
	}

	// Gen2+ or unknown: Try Gen2 first (more common)
	result, err := s.DeviceInfo(ctx, identifier)
	if err == nil {
		return result, nil
	}

	// Gen2 failed, try Gen1
	gen1Result, gen1Err := s.DeviceInfoGen1(ctx, identifier)
	if gen1Err == nil {
		return gen1Result, nil
	}

	// Both failed, return the original Gen2 error (more informative)
	return nil, err
}

// DeviceInfoGen1 returns information about a Gen1 device.
func (s *Service) DeviceInfoGen1(ctx context.Context, identifier string) (*DeviceInfo, error) {
	// Resolve first to capture the address
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	var result *DeviceInfo
	err = s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		info := conn.Info()
		result = &DeviceInfo{
			ID:         info.ID,
			MAC:        info.MAC,
			Model:      info.Model,
			Generation: info.Generation,
			Firmware:   info.Firmware,
			App:        info.App,
			AuthEn:     info.AuthEn,
			Address:    dev.Address,
		}
		return nil
	})
	if err == nil {
		refreshDeviceMetadata(identifier, result)
	}
	return result, err
}

// refreshDeviceMetadata opportunistically updates stored device metadata
// from a successful device info response. This keeps MAC, Type, Model,
// and Generation up to date without requiring explicit refresh commands.
// Errors are expected when the device is addressed by IP and not registered.
func refreshDeviceMetadata(identifier string, info *DeviceInfo) {
	if info == nil {
		return
	}
	updates := config.DeviceUpdates{
		MAC:        info.MAC,
		Type:       info.Model, // info.Model is the SKU/type identifier
		Generation: info.Generation,
	}
	// Derive display model name from SKU
	if info.Model != "" {
		if displayName := types.ModelDisplayName(info.Model); displayName != info.Model {
			updates.Model = displayName
		}
	}
	if err := config.UpdateDeviceInfo(identifier, updates); err != nil {
		// Expected for IP-addressed devices not in registry - trace only
		debug.TraceEvent("metadata refresh for %s: %v", identifier, err)
	}
}

// RefreshAllDeviceMetadata fetches device info from all devices concurrently
// and updates the config with the latest metadata. Shows progress via spinner.
func (s *Service) RefreshAllDeviceMetadata(ctx context.Context, ios *iostreams.IOStreams, devices map[string]model.Device) {
	ios.StartProgress("Refreshing device metadata...")
	defer ios.StopProgress()

	var wg sync.WaitGroup
	for name := range devices {
		wg.Go(func() {
			// DeviceInfoAuto triggers automatic metadata refresh on success
			if _, err := s.DeviceInfoAuto(ctx, name); err != nil {
				debug.TraceEvent("refresh metadata for %s: %v", name, err)
			}
		})
	}
	wg.Wait()
}

// DeviceListFilterOptions holds filter criteria for device lists.
type DeviceListFilterOptions struct {
	Generation int
	DeviceType string
	Platform   string
}

// FilterDeviceList filters devices based on criteria and converts to DeviceListItem slice.
// Returns the filtered list and a set of unique platforms found.
func FilterDeviceList(devices map[string]model.Device, opts DeviceListFilterOptions) (filtered []model.DeviceListItem, platforms map[string]struct{}) {
	filtered = make([]model.DeviceListItem, 0, len(devices))
	platforms = make(map[string]struct{})

	for name, dev := range devices {
		if !matchesDeviceFilters(dev, opts) {
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

// matchesDeviceFilters checks if a device matches all filter criteria.
func matchesDeviceFilters(dev model.Device, opts DeviceListFilterOptions) bool {
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

// SortDeviceList sorts a device list. If updatesFirst is true, devices with
// available updates are sorted to the top. Within each group, devices are sorted by name.
func SortDeviceList(devices []model.DeviceListItem, updatesFirst bool) {
	sort.Slice(devices, func(i, j int) bool {
		if updatesFirst && devices[i].HasUpdate != devices[j].HasUpdate {
			return devices[i].HasUpdate // true sorts before false
		}
		return devices[i].Name < devices[j].Name
	})
}

// PopulateDeviceListFirmware fills in firmware version info from the cache.
// Uses a short cache validity period (5 minutes) so it doesn't trigger network calls during list.
func (s *Service) PopulateDeviceListFirmware(ctx context.Context, devices []model.DeviceListItem) {
	const cacheMaxAge = 5 * time.Minute
	for i := range devices {
		entry := s.GetCachedFirmware(ctx, devices[i].Name, cacheMaxAge)
		if entry != nil && entry.Info != nil {
			devices[i].CurrentVersion = entry.Info.Current
			devices[i].AvailableVersion = entry.Info.Available
			devices[i].HasUpdate = entry.Info.HasUpdate
		}
	}
}
