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
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// DeviceInfo holds extended device information.
type DeviceInfo struct {
	ID         string `json:"id" yaml:"id"`
	MAC        string `json:"mac" yaml:"mac"`
	Type       string `json:"type" yaml:"type"`   // Raw SKU/type code (e.g., "SNSW-001P16EU", "SHSW-1")
	Model      string `json:"model" yaml:"model"` // Display name (e.g., "Shelly Plus 1PM") derived via types.ModelDisplayName
	Generation int    `json:"generation" yaml:"generation"`
	Firmware   string `json:"firmware" yaml:"firmware"`
	App        string `json:"app" yaml:"app"`
	AuthEn     bool   `json:"auth_en" yaml:"auth_en"`
	Address    string `json:"address" yaml:"address"`
}

// DeviceStatus holds device status information.
type DeviceStatus struct {
	Info   *DeviceInfo    `json:"info" yaml:"info"`
	Status map[string]any `json:"status" yaml:"status"`
}

// DeviceReboot reboots the device. Supports both Gen1 (HTTP REST /reboot) and Gen2+
// (RPC). delayMS applies only to Gen2 — Gen1's reboot has no delay parameter.
func (s *Service) DeviceReboot(ctx context.Context, identifier string, delayMS int) error {
	return s.withGenAwareRestart(ctx, identifier, "reboot",
		func(conn *client.Gen1Client) error { return conn.Reboot(ctx) },
		func(conn *client.Client) error { return conn.Reboot(ctx, delayMS) },
	)
}

// withGenAwareRestart runs a gen-aware action whose purpose is to restart the device
// (reboot, factory reset). Such a device tears down the connection as it restarts, so
// a connectivity error from the action is the SUCCESS signal, not a failure.
// Generation detection runs first and must reach the device — so an unreachable device
// (which was never restarted) still errors honestly — while a non-connectivity error
// from the action itself (refused, unsupported, auth) is returned as a real failure.
func (s *Service) withGenAwareRestart(
	ctx context.Context,
	identifier, op string,
	gen1Fn func(*client.Gen1Client) error,
	gen2Fn func(*client.Client) error,
) error {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		err = s.WithGen1Connection(ctx, identifier, gen1Fn)
	} else {
		err = s.WithConnection(ctx, identifier, gen2Fn)
	}
	if err == nil {
		return nil
	}
	// If the caller's own context was cancelled or timed out, the operation was
	// aborted — surface that, never silently swallow it as a "successful" restart.
	// (IsConnectivityFailure treats context cancellation as a connectivity error, so
	// this check must come first.)
	if ctx.Err() != nil {
		return err
	}
	// With a still-live context, a connectivity error means the device tore down the
	// connection as it restarted — the success signal. Any other error is a real
	// failure (the request was refused, unsupported, or the device was never reached).
	if !ratelimit.IsConnectivityFailure(err) {
		return err
	}
	debug.TraceEvent("%s: device dropped the connection as it restarted (expected): %v", op, err)
	return nil
}

// DeviceFactoryReset performs a factory reset on the device, restarting it into AP
// mode. Supports both Gen1 (HTTP REST) and Gen2+ (RPC) devices. The device drops the
// connection as it wipes, which withGenAwareRestart treats as the reset taking effect.
func (s *Service) DeviceFactoryReset(ctx context.Context, identifier string) error {
	return s.withGenAwareRestart(ctx, identifier, "factory-reset",
		func(conn *client.Gen1Client) error { return conn.FactoryReset(ctx) },
		func(conn *client.Client) error { return conn.FactoryReset(ctx) },
	)
}

// deviceInfoFrom converts a connection's device info, reached at addr.
func deviceInfoFrom(info *client.DeviceInfo, addr string) *DeviceInfo {
	return &DeviceInfo{
		ID:         info.ID,
		MAC:        info.MAC,
		Type:       info.Model,
		Model:      types.ModelDisplayName(info.Model),
		Generation: info.Generation,
		Firmware:   info.Firmware,
		App:        info.App,
		AuthEn:     info.AuthEn,
		Address:    addr,
	}
}

// deviceInfoGen2 returns information about the resolved device dev over RPC.
func (s *Service) deviceInfoGen2(ctx context.Context, identifier string, dev model.Device) (*DeviceInfo, error) {
	var result *DeviceInfo
	err := s.connManager.WithDeviceConnection(ctx, dev, func(conn *client.Client) error {
		result = deviceInfoFrom(conn.Info(), dev.Address)
		return nil
	})
	if err != nil {
		return nil, err
	}
	refreshDeviceMetadata(identifier, result)
	return result, nil
}

// deviceInfoGen1 returns information about the resolved device dev over the
// Gen1 HTTP API.
func (s *Service) deviceInfoGen1(ctx context.Context, identifier string, dev model.Device) (*DeviceInfo, error) {
	var result *DeviceInfo
	err := s.connManager.WithGen1DeviceConnection(ctx, dev, func(conn *client.Gen1Client) error {
		result = deviceInfoFrom(conn.Info(), dev.Address)
		return nil
	})
	if err != nil {
		return nil, err
	}
	refreshDeviceMetadata(identifier, result)
	return result, nil
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
				Type:       info.Model,
				Model:      types.ModelDisplayName(info.Model),
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
				Type:       info.Model,
				Model:      types.ModelDisplayName(info.Model),
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
func (s *Service) DevicePing(ctx context.Context, identifier string) (*DeviceInfo, error) {
	return s.DeviceInfo(ctx, identifier)
}

// DeviceInfo returns information about the device, for any generation. The
// identifier is resolved once. A device known to be Gen1 is asked over the
// Gen1 HTTP API first; a Gen2+ or unknown device over RPC first. The other
// generation is tried next, except when the generation is known and the
// device did not answer at all: a device of the other generation would have
// answered, so a second attempt only doubles the wait for an offline device.
// When both fail the error of the generation tried first is returned.
func (s *Service) DeviceInfo(ctx context.Context, identifier string) (*DeviceInfo, error) {
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return s.deviceInfoFor(ctx, identifier, dev)
}

// ProbeDevice returns information about the device at dev.Address, connecting
// with dev.Auth and treating dev.Generation as a hint, without looking the
// address up in the registry. Use it to check a device before registering it,
// so the credentials the user just gave are the ones used.
func (s *Service) ProbeDevice(ctx context.Context, dev model.Device) (*DeviceInfo, error) {
	// /shelly names the generation without credentials. Asking it first keeps
	// a Gen1 device from being tried as Gen2+ with credentials, where its
	// basic-auth challenge is an unusable digest challenge that the transport
	// retries for seconds. Without credentials there is no such retry, so the
	// extra round trip is skipped.
	if dev.Generation == 0 && dev.HasAuth() {
		if detected := tryDetectGeneration(ctx, dev.Address, dev.Auth); detected != nil {
			dev.Generation = int(detected.Generation)
		}
	}
	info, err := s.deviceInfoFor(ctx, dev.Address, dev)
	return info, credentialsError(dev, err)
}

// IdentifyNewDevice asks a device that is about to be registered what it is,
// and returns dev with its generation, type and model filled in. A device of a
// plugin platform (dev.Platform other than shelly) is asked through that
// platform's plugin; any other device through the Shelly API. authEnabled
// reports whether a Shelly device said it requires authentication.
//
// The request that identifies a Shelly device is answered without
// credentials, so when the device requires authentication and dev.Auth is set
// the credentials are checked with VerifyCredentials, and a device that
// rejects them is an error (errors.Is ErrCredentialsRejected).
func (s *Service) IdentifyNewDevice(ctx context.Context, dev model.Device) (identified model.Device, authEnabled bool, err error) {
	if dev.IsPluginManaged() {
		res, err := s.DetectPluginDevice(ctx, dev.Platform, dev.Address, dev.Auth)
		if err != nil {
			return dev, false, err
		}
		dev.Generation = 0
		dev.Type = res.Model
		dev.Model = types.ModelDisplayName(res.Model)
		return dev, false, nil
	}
	info, err := s.ProbeDevice(ctx, dev)
	if err != nil {
		return dev, false, err
	}
	dev.Generation = info.Generation
	dev.Type = info.Type
	dev.Model = info.Model
	if info.AuthEn && dev.HasAuth() {
		if err := s.VerifyCredentials(ctx, dev); err != nil {
			return dev, true, err
		}
	}
	return dev, info.AuthEn, nil
}

// deviceInfoFor asks the already resolved device dev for its information,
// trying its known generation first.
func (s *Service) deviceInfoFor(ctx context.Context, identifier string, dev model.Device) (*DeviceInfo, error) {
	first, second := s.deviceInfoGen2, s.deviceInfoGen1
	if dev.Generation == 1 {
		first, second = s.deviceInfoGen1, s.deviceInfoGen2
	}
	result, err := first(ctx, identifier, dev)
	if err == nil {
		return result, nil
	}
	if dev.Generation != 0 && ratelimit.IsConnectivityFailure(err) {
		return nil, err
	}
	if result, otherErr := second(ctx, identifier, dev); otherErr == nil {
		return result, nil
	}
	return nil, err
}

// refreshDeviceMetadata opportunistically updates stored device metadata
// from a successful device info response. Uses MAC as the authoritative
// device identifier to find the config entry, falling back to name-based
// lookup for devices without a stored MAC. Updates Type, Generation, Address,
// and MAC. Model is not stored — it's derived from Type on config load.
func refreshDeviceMetadata(identifier string, info *DeviceInfo) {
	if info == nil {
		return
	}

	updates := config.DeviceUpdates{
		MAC:        info.MAC,
		Type:       info.Type,
		Generation: info.Generation,
		Address:    info.Address,
	}

	// MAC-based lookup is authoritative — find config key by MAC first
	if info.MAC != "" {
		if key := config.FindDeviceKeyByMAC(info.MAC); key != "" {
			if err := config.UpdateDeviceInfo(key, updates); err != nil {
				debug.TraceEvent("metadata refresh for %s (key=%s): %v", identifier, key, err)
			}
			return
		}
	}

	// Fallback: name-based lookup (for devices without a stored MAC yet)
	if err := config.UpdateDeviceInfo(identifier, updates); err != nil {
		// Expected for IP-addressed devices not in registry
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
			// DeviceInfo triggers automatic metadata refresh on success
			if _, err := s.DeviceInfo(ctx, name); err != nil {
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
