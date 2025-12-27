// Package firmware provides firmware management for Shelly devices.
package firmware

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

// Info contains firmware update information.
type Info struct {
	Current     string
	Available   string
	Beta        string
	HasUpdate   bool
	DeviceModel string
	DeviceID    string
	Generation  int
	Platform    string // "shelly", "tasmota", etc.
}

// Status contains the current firmware status.
type Status struct {
	Status      string
	HasUpdate   bool
	NewVersion  string
	Progress    int
	CanRollback bool
}

// CheckResult holds the result of a firmware check for a single device.
type CheckResult struct {
	Name string
	Info *Info
	Err  error
}

// DeviceUpdateStatus holds the status of a device for update operations.
type DeviceUpdateStatus struct {
	Name      string
	Info      *Info
	HasUpdate bool
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

// UpdateEntry represents a device with available firmware updates.
type UpdateEntry struct {
	Name      string
	Device    model.Device
	FwInfo    *Info
	HasUpdate bool
	HasBeta   bool
	Error     error
}

// ConnectionHandler provides connection management for firmware operations.
type ConnectionHandler interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// DeviceChecker provides device firmware checking capability.
type DeviceChecker interface {
	CheckDeviceFirmware(ctx context.Context, device model.Device) (*Info, error)
}

// Service provides firmware management operations.
type Service struct {
	connHandler    ConnectionHandler
	pluginRegistry *plugins.Registry
	cache          *Cache
}

// NewService creates a new firmware service.
func NewService(connHandler ConnectionHandler) *Service {
	return &Service{
		connHandler: connHandler,
		cache:       NewCache(),
	}
}

// SetPluginRegistry sets the plugin registry for plugin-managed device support.
func (s *Service) SetPluginRegistry(registry *plugins.Registry) {
	s.pluginRegistry = registry
}

// Cache returns the firmware cache.
func (s *Service) Cache() *Cache {
	return s.cache
}

// Check checks for firmware updates on a device.
func (s *Service) Check(ctx context.Context, identifier string) (*Info, error) {
	var result *Info
	err := s.connHandler.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		info, err := mgr.CheckForUpdate(ctx)
		if err != nil {
			return err
		}

		deviceInfo := conn.Info()
		result = &Info{
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

// GetStatus gets the current firmware status.
func (s *Service) GetStatus(ctx context.Context, identifier string) (*Status, error) {
	var result *Status
	err := s.connHandler.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

		result = &Status{
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

// Update starts a firmware update on a device.
func (s *Service) Update(ctx context.Context, identifier string, opts *firmware.UpdateOptions) error {
	return s.connHandler.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Update(ctx, opts)
	})
}

// UpdateStable updates to the latest stable firmware.
func (s *Service) UpdateStable(ctx context.Context, identifier string) error {
	return s.Update(ctx, identifier, &firmware.UpdateOptions{Stage: "stable"})
}

// UpdateBeta updates to the latest beta firmware.
func (s *Service) UpdateBeta(ctx context.Context, identifier string) error {
	return s.Update(ctx, identifier, &firmware.UpdateOptions{Stage: "beta"})
}

// UpdateFromURL updates from a custom firmware URL.
func (s *Service) UpdateFromURL(ctx context.Context, identifier, url string) error {
	return s.Update(ctx, identifier, &firmware.UpdateOptions{URL: url})
}

// Rollback rolls back to the previous firmware version.
func (s *Service) Rollback(ctx context.Context, identifier string) error {
	return s.connHandler.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mgr := firmware.New(conn.RPCClient())
		return mgr.Rollback(ctx)
	})
}

// GetURL gets the firmware download URL for a device.
func (s *Service) GetURL(ctx context.Context, identifier, stage string) (string, error) {
	var result string
	err := s.connHandler.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// CheckAll checks firmware on multiple devices concurrently.
func (s *Service) CheckAll(ctx context.Context, ios *iostreams.IOStreams, devices []string) []CheckResult {
	var (
		results []CheckResult
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	// Use global rate limit for concurrency (service layer also enforces this)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for _, name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := s.Check(gctx, deviceName)
			mu.Lock()
			results = append(results, CheckResult{Name: deviceName, Info: info, Err: checkErr})
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

// CheckAllPlatforms checks firmware for all devices including plugin-managed ones.
// This is the platform-aware version that uses plugin hooks for non-Shelly devices.
func (s *Service) CheckAllPlatforms(ctx context.Context, ios *iostreams.IOStreams, deviceConfigs map[string]model.Device) []CheckResult {
	var (
		results []CheckResult
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(config.GetGlobalMaxConcurrent())

	for name, device := range deviceConfigs {
		deviceName := name
		dev := device
		g.Go(func() error {
			var info *Info
			var checkErr error

			if dev.IsPluginManaged() {
				info, checkErr = s.checkPluginFirmware(gctx, dev)
			} else {
				info, checkErr = s.Check(gctx, deviceName)
			}

			mu.Lock()
			results = append(results, CheckResult{Name: deviceName, Info: info, Err: checkErr})
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
func (s *Service) checkPluginFirmware(ctx context.Context, device model.Device) (*Info, error) {
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

	return &Info{
		Current:     result.CurrentVersion,
		Available:   result.LatestStable,
		Beta:        result.LatestBeta,
		HasUpdate:   result.HasUpdate,
		DeviceModel: device.Model,
		DeviceID:    device.Name,
		Platform:    device.Platform,
	}, nil
}

// CheckPlugin checks for firmware updates on a plugin-managed device.
// This is the public version of checkPluginFirmware for use by commands.
func (s *Service) CheckPlugin(ctx context.Context, device model.Device) (*Info, error) {
	return s.checkPluginFirmware(ctx, device)
}

// UpdatePlugin applies a firmware update to a plugin-managed device.
func (s *Service) UpdatePlugin(ctx context.Context, device model.Device, stage, url string) error {
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

// CheckDevice checks firmware for a device (either plugin-managed or native Shelly).
// This is a unified entry point that handles platform detection and dispatches to the
// appropriate check method.
func (s *Service) CheckDevice(ctx context.Context, device model.Device) (*Info, error) {
	if device.IsPluginManaged() {
		return s.CheckPlugin(ctx, device)
	}
	return s.Check(ctx, device.Name)
}

// UpdateDevice updates firmware for a device (either plugin-managed or native Shelly).
// This is a unified entry point that handles platform detection and dispatches to the
// appropriate update method. Returns nil on success.
func (s *Service) UpdateDevice(ctx context.Context, device model.Device, useBeta bool, customURL string) error {
	if device.IsPluginManaged() {
		stage := "stable"
		if useBeta {
			stage = "beta"
		}
		return s.UpdatePlugin(ctx, device, stage, customURL)
	}

	switch {
	case customURL != "":
		return s.UpdateFromURL(ctx, device.Name, customURL)
	case useBeta:
		return s.UpdateBeta(ctx, device.Name)
	default:
		return s.UpdateStable(ctx, device.Name)
	}
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
			info, checkErr := s.Check(gctx, deviceName)
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
				updateErr = s.UpdateFromURL(gctx, dev.Name, opts.CustomURL)
			case opts.Beta:
				updateErr = s.UpdateBeta(gctx, dev.Name)
			default:
				updateErr = s.UpdateStable(gctx, dev.Name)
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

// BuildUpdateList creates a sorted list of devices that have updates available.
func BuildUpdateList(results []CheckResult, devices map[string]model.Device) []UpdateEntry {
	var entries []UpdateEntry
	for _, r := range results {
		device := devices[r.Name]
		entry := UpdateEntry{
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
func FilterEntriesByStage(entries []UpdateEntry, beta bool) []UpdateEntry {
	var result []UpdateEntry
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
func AnyHasBeta(entries []UpdateEntry) bool {
	for _, e := range entries {
		if e.HasBeta {
			return true
		}
	}
	return false
}

// SelectEntriesByStage selects entry indices based on the beta flag and returns the stage name.
func SelectEntriesByStage(entries []UpdateEntry, beta bool) (indices []int, stage string) {
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
func GetEntriesByIndices(entries []UpdateEntry, indices []int) []UpdateEntry {
	var result []UpdateEntry
	for _, idx := range indices {
		if idx >= 0 && idx < len(entries) {
			result = append(result, entries[idx])
		}
	}
	return result
}

// ToDeviceUpdateStatuses converts firmware entries to device update statuses.
func ToDeviceUpdateStatuses(entries []UpdateEntry) []DeviceUpdateStatus {
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
