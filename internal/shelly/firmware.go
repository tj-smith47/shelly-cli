// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tj-smith47/shelly-go/firmware"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// FirmwareInfo contains firmware update information.
type FirmwareInfo struct {
	Current     string
	Available   string
	Beta        string
	HasUpdate   bool
	DeviceModel string
	DeviceID    string
	Generation  int
	Platform    string // "shelly", "tasmota", etc.
}

// FirmwareStatus contains the current firmware status.
type FirmwareStatus struct {
	Status      string
	HasUpdate   bool
	NewVersion  string
	Progress    int
	CanRollback bool
}

// CheckFirmware checks for firmware updates on a device.
func (s *Service) CheckFirmware(ctx context.Context, identifier string) (*FirmwareInfo, error) {
	var result *FirmwareInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		info, err := mgr.CheckForUpdate(ctx)
		if err != nil {
			return err
		}

		deviceInfo := conn.Info()
		result = &FirmwareInfo{
			Current:     info.Current,
			Available:   info.Available,
			Beta:        info.Beta,
			HasUpdate:   info.HasUpdate(),
			DeviceModel: deviceInfo.Model,
			DeviceID:    deviceInfo.ID,
			Generation:  deviceInfo.Generation,
			Platform:    "shelly", // Native Shelly device
		}
		return nil
	})
	return result, err
}

// GetFirmwareStatus gets the current firmware status.
func (s *Service) GetFirmwareStatus(ctx context.Context, identifier string) (*FirmwareStatus, error) {
	var result *FirmwareStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())

		status, err := mgr.GetStatus(ctx)
		if err != nil {
			return err
		}

		rollbackStatus, rollbackErr := mgr.GetRollbackStatus(ctx)
		canRollback := false
		if rollbackErr == nil && rollbackStatus != nil {
			canRollback = rollbackStatus.CanRollback
		}

		result = &FirmwareStatus{
			Status:      status.Status,
			HasUpdate:   status.HasUpdate,
			NewVersion:  status.NewVersion,
			Progress:    status.Progress,
			CanRollback: canRollback,
		}
		return nil
	})
	return result, err
}

// UpdateFirmware starts a firmware update on a device.
func (s *Service) UpdateFirmware(ctx context.Context, identifier string, opts *firmware.UpdateOptions) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Update(ctx, opts)
	})
}

// UpdateFirmwareStable updates to the latest stable firmware.
func (s *Service) UpdateFirmwareStable(ctx context.Context, identifier string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{Stage: "stable"})
}

// UpdateFirmwareBeta updates to the latest beta firmware.
func (s *Service) UpdateFirmwareBeta(ctx context.Context, identifier string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{Stage: "beta"})
}

// UpdateFirmwareFromURL updates from a custom firmware URL.
func (s *Service) UpdateFirmwareFromURL(ctx context.Context, identifier, url string) error {
	return s.UpdateFirmware(ctx, identifier, &firmware.UpdateOptions{URL: url})
}

// RollbackFirmware rolls back to the previous firmware version.
func (s *Service) RollbackFirmware(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Rollback(ctx)
	})
}

// GetFirmwareURL gets the firmware download URL for a device.
func (s *Service) GetFirmwareURL(ctx context.Context, identifier, stage string) (string, error) {
	var result string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		url, err := mgr.GetFirmwareURL(ctx, stage)
		if err != nil {
			return err
		}
		result = url
		return nil
	})
	return result, err
}

// FirmwareCheckResult holds the result of a firmware check for a single device.
type FirmwareCheckResult struct {
	Name string
	Info *FirmwareInfo
	Err  error
}

// CheckFirmwareAll checks firmware on multiple devices concurrently.
func (s *Service) CheckFirmwareAll(ctx context.Context, ios *iostreams.IOStreams, devices []string) []FirmwareCheckResult {
	var (
		results []FirmwareCheckResult
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for _, name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := s.CheckFirmware(gctx, deviceName)
			mu.Lock()
			results = append(results, FirmwareCheckResult{Name: deviceName, Info: info, Err: checkErr})
			mu.Unlock()
			return nil // Don't fail the whole group on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	return results
}

// CheckFirmwareAllPlatforms checks firmware for all devices including plugin-managed ones.
// This is the platform-aware version that uses plugin hooks for non-Shelly devices.
func (s *Service) CheckFirmwareAllPlatforms(ctx context.Context, ios *iostreams.IOStreams, deviceConfigs map[string]model.Device) []FirmwareCheckResult {
	var (
		results []FirmwareCheckResult
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for name, device := range deviceConfigs {
		deviceName := name
		dev := device
		g.Go(func() error {
			var info *FirmwareInfo
			var checkErr error

			if dev.IsPluginManaged() {
				info, checkErr = s.checkPluginFirmware(gctx, dev)
			} else {
				info, checkErr = s.CheckFirmware(gctx, deviceName)
			}

			mu.Lock()
			results = append(results, FirmwareCheckResult{Name: deviceName, Info: info, Err: checkErr})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	return results
}

// checkPluginFirmware checks firmware updates for a plugin-managed device.
func (s *Service) checkPluginFirmware(ctx context.Context, device model.Device) (*FirmwareInfo, error) {
	if s.pluginRegistry == nil {
		return nil, fmt.Errorf("plugin registry not configured")
	}

	plugin, err := s.pluginRegistry.FindByPlatform(device.Platform)
	if err != nil {
		return nil, fmt.Errorf("error finding plugin for platform %q: %w", device.Platform, err)
	}
	if plugin == nil {
		return nil, fmt.Errorf("no plugin found for platform %q", device.Platform)
	}

	executor := plugins.NewHookExecutor(plugin)
	result, err := executor.ExecuteCheckUpdates(ctx, device.Address, device.Auth)
	if err != nil {
		return nil, err
	}

	return &FirmwareInfo{
		Current:     result.CurrentVersion,
		Available:   result.LatestStable,
		Beta:        result.LatestBeta,
		HasUpdate:   result.HasUpdate,
		DeviceModel: device.Model,
		DeviceID:    device.Name,
		Platform:    device.Platform,
	}, nil
}

// UpdatePluginFirmware applies a firmware update to a plugin-managed device.
func (s *Service) UpdatePluginFirmware(ctx context.Context, device model.Device, stage, url string) error {
	if s.pluginRegistry == nil {
		return fmt.Errorf("plugin registry not configured")
	}

	plugin, err := s.pluginRegistry.FindByPlatform(device.Platform)
	if err != nil {
		return fmt.Errorf("error finding plugin for platform %q: %w", device.Platform, err)
	}
	if plugin == nil {
		return fmt.Errorf("no plugin found for platform %q", device.Platform)
	}

	executor := plugins.NewHookExecutor(plugin)
	result, err := executor.ExecuteApplyUpdate(ctx, device.Address, device.Auth, stage, url)
	if err != nil {
		return err
	}

	if !result.Success {
		errMsg := result.Error
		if errMsg == "" {
			errMsg = "update failed (no error details)"
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// UpdateDeviceFirmware updates firmware for a device (either plugin-managed or native Shelly).
// This is a unified entry point that handles platform detection and dispatches to the
// appropriate update method. Returns nil on success.
func (s *Service) UpdateDeviceFirmware(ctx context.Context, device model.Device, useBeta bool, customURL string) error {
	if device.IsPluginManaged() {
		stage := "stable"
		if useBeta {
			stage = "beta"
		}
		return s.UpdatePluginFirmware(ctx, device, stage, customURL)
	}

	switch {
	case customURL != "":
		return s.UpdateFirmwareFromURL(ctx, device.Name, customURL)
	case useBeta:
		return s.UpdateFirmwareBeta(ctx, device.Name)
	default:
		return s.UpdateFirmwareStable(ctx, device.Name)
	}
}

// CheckDeviceFirmware checks firmware for a device (either plugin-managed or native Shelly).
// This is a unified entry point that handles platform detection and dispatches to the
// appropriate check method.
func (s *Service) CheckDeviceFirmware(ctx context.Context, device model.Device) (*FirmwareInfo, error) {
	if device.IsPluginManaged() {
		return s.CheckPluginFirmware(ctx, device)
	}
	return s.CheckFirmware(ctx, device.Name)
}

// CheckPluginFirmware checks for firmware updates on a plugin-managed device.
// This is the public version of checkPluginFirmware for use by commands.
func (s *Service) CheckPluginFirmware(ctx context.Context, device model.Device) (*FirmwareInfo, error) {
	return s.checkPluginFirmware(ctx, device)
}

// DeviceUpdateStatus holds the status of a device for update operations.
type DeviceUpdateStatus struct {
	Name      string
	Info      *FirmwareInfo
	HasUpdate bool
}

// CheckDevicesForUpdates checks multiple devices for firmware updates and returns those needing updates.
// The staged parameter controls what percentage of devices with updates to return (for staged rollouts).
func (s *Service) CheckDevicesForUpdates(ctx context.Context, ios *iostreams.IOStreams, devices []string, staged int) []DeviceUpdateStatus {
	ios.StartProgress("Checking devices for updates...")

	var (
		statuses []DeviceUpdateStatus
		mu       sync.Mutex
	)

	g, gctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for _, name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := s.CheckFirmware(gctx, deviceName)
			hasUpdate := checkErr == nil && info != nil && info.HasUpdate
			mu.Lock()
			statuses = append(statuses, DeviceUpdateStatus{
				Name:      deviceName,
				Info:      info,
				HasUpdate: hasUpdate,
			})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	// Filter devices with updates and apply staged percentage
	var toUpdate []DeviceUpdateStatus
	for _, s := range statuses {
		if s.HasUpdate {
			toUpdate = append(toUpdate, s)
		}
	}

	// Apply staged percentage
	targetCount := len(toUpdate) * staged / 100
	if targetCount == 0 && staged > 0 && len(toUpdate) > 0 {
		targetCount = 1
	}
	if targetCount < len(toUpdate) {
		toUpdate = toUpdate[:targetCount]
	}

	return toUpdate
}

// UpdateOpts contains options for firmware update operations.
type UpdateOpts struct {
	Beta        bool
	CustomURL   string
	Parallelism int
}

// UpdateResult holds the result of a single device update.
type UpdateResult struct {
	Name    string
	Success bool
	Err     error
}

// UpdateDevices performs firmware updates on multiple devices concurrently.
func (s *Service) UpdateDevices(ctx context.Context, ios *iostreams.IOStreams, devices []DeviceUpdateStatus, opts UpdateOpts) []UpdateResult {
	// Cap parallelism to global rate limit
	parallelism := opts.Parallelism
	globalMax := config.GetGlobalMaxConcurrent()
	if parallelism > globalMax {
		ios.Warning("Requested parallelism %d exceeds global rate limit %d; capping to %d.\n"+
			"  Adjust ratelimit.global.max_concurrent in config to increase.",
			parallelism, globalMax, globalMax)
		parallelism = globalMax
	}

	ios.StartProgress("Updating devices...")

	var (
		results []UpdateResult
		mu      sync.Mutex
	)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(parallelism)

	for _, status := range devices {
		dev := status
		g.Go(func() error {
			var updateErr error
			switch {
			case opts.CustomURL != "":
				updateErr = s.UpdateFirmwareFromURL(gctx, dev.Name, opts.CustomURL)
			case opts.Beta:
				updateErr = s.UpdateFirmwareBeta(gctx, dev.Name)
			default:
				updateErr = s.UpdateFirmwareStable(gctx, dev.Name)
			}
			mu.Lock()
			results = append(results, UpdateResult{
				Name:    dev.Name,
				Success: updateErr == nil,
				Err:     updateErr,
			})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	return results
}

// FirmwareUpdateEntry represents a device with available firmware updates.
type FirmwareUpdateEntry struct {
	Name      string
	Device    model.Device
	FwInfo    *FirmwareInfo
	HasUpdate bool
	HasBeta   bool
	Error     error
}

// BuildFirmwareUpdateList creates a sorted list of devices that have updates available.
func BuildFirmwareUpdateList(results []FirmwareCheckResult, devices map[string]model.Device) []FirmwareUpdateEntry {
	var entries []FirmwareUpdateEntry
	for _, r := range results {
		device := devices[r.Name]
		entry := FirmwareUpdateEntry{
			Name:   r.Name,
			Device: device,
			FwInfo: r.Info,
		}
		if r.Err != nil {
			entry.Error = r.Err
		} else if r.Info != nil {
			entry.HasUpdate = r.Info.HasUpdate
			entry.HasBeta = r.Info.Beta != "" && r.Info.Beta != r.Info.Current
		}
		if entry.HasUpdate || entry.HasBeta {
			entries = append(entries, entry)
		}
	}

	// Sort by device name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// FilterDevicesByNameAndPlatform filters devices based on device name list and platform.
func FilterDevicesByNameAndPlatform(devices map[string]model.Device, devicesList, platform string) map[string]model.Device {
	result := make(map[string]model.Device)

	// If devicesList is specified, filter by names
	var selectedNames map[string]bool
	if devicesList != "" {
		selectedNames = make(map[string]bool)
		for _, name := range strings.Split(devicesList, ",") {
			selectedNames[strings.TrimSpace(name)] = true
		}
	}

	for name, device := range devices {
		// Filter by name if devicesList specified
		if selectedNames != nil && !selectedNames[name] {
			continue
		}

		// Filter by platform if specified
		if platform != "" {
			devicePlatform := device.Platform
			if devicePlatform == "" {
				devicePlatform = "shelly"
			}
			if devicePlatform != platform {
				continue
			}
		}

		result[name] = device
	}

	return result
}

// FilterEntriesByStage filters firmware update entries based on the requested stage.
func FilterEntriesByStage(entries []FirmwareUpdateEntry, beta bool) []FirmwareUpdateEntry {
	var result []FirmwareUpdateEntry
	for _, e := range entries {
		if beta && e.HasBeta {
			result = append(result, e)
		} else if e.HasUpdate {
			result = append(result, e)
		}
	}
	return result
}

// AnyHasBeta returns true if any entry has a beta update available.
func AnyHasBeta(entries []FirmwareUpdateEntry) bool {
	for _, e := range entries {
		if e.HasBeta {
			return true
		}
	}
	return false
}

// SelectEntriesByStage selects entry indices based on the beta flag and returns the stage name.
func SelectEntriesByStage(entries []FirmwareUpdateEntry, beta bool) (indices []int, stage string) {
	stage = "stable"
	if beta {
		stage = "beta"
	}
	filtered := FilterEntriesByStage(entries, beta)
	nameSet := make(map[string]bool, len(filtered))
	for _, f := range filtered {
		nameSet[f.Name] = true
	}
	for i, e := range entries {
		if nameSet[e.Name] {
			indices = append(indices, i)
		}
	}
	return indices, stage
}

// GetEntriesByIndices returns entries at the specified indices.
func GetEntriesByIndices(entries []FirmwareUpdateEntry, indices []int) []FirmwareUpdateEntry {
	var result []FirmwareUpdateEntry
	for _, idx := range indices {
		if idx >= 0 && idx < len(entries) {
			result = append(result, entries[idx])
		}
	}
	return result
}

// ToDeviceUpdateStatuses converts firmware entries to device update statuses.
func ToDeviceUpdateStatuses(entries []FirmwareUpdateEntry) []DeviceUpdateStatus {
	result := make([]DeviceUpdateStatus, len(entries))
	for i, d := range entries {
		result[i] = DeviceUpdateStatus{
			Name:      d.Name,
			Info:      d.FwInfo,
			HasUpdate: d.HasUpdate,
		}
	}
	return result
}
