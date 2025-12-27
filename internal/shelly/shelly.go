// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"errors"
	"strings"
	"time"

	libfirmware "github.com/tj-smith47/shelly-go/firmware"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/firmware"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
)

// DefaultTimeout is the default timeout for device operations.
const DefaultTimeout = 10 * time.Second

// Service provides high-level operations on Shelly devices.
type Service struct {
	resolver        DeviceResolver
	rateLimiter     *ratelimit.DeviceRateLimiter
	pluginRegistry  *plugins.Registry
	firmwareService *firmware.Service
	wirelessService *wireless.Service
	networkService  *network.WiFiService
}

// DeviceResolver resolves device identifiers to device configurations.
type DeviceResolver interface {
	Resolve(identifier string) (model.Device, error)
}

// GenerationAwareResolver extends DeviceResolver with generation detection.
type GenerationAwareResolver interface {
	DeviceResolver
	ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error)
}

// ServiceOption configures a Service.
type ServiceOption func(*Service)

// WithRateLimiter configures the service to use rate limiting.
// If not provided, no rate limiting is applied (backward compatible).
func WithRateLimiter(rl *ratelimit.DeviceRateLimiter) ServiceOption {
	return func(s *Service) {
		s.rateLimiter = rl
	}
}

// WithDefaultRateLimiter configures the service with default rate limiting.
// This is recommended for TUI usage to prevent overloading Shelly devices.
func WithDefaultRateLimiter() ServiceOption {
	return func(s *Service) {
		s.rateLimiter = ratelimit.New()
	}
}

// WithRateLimiterFromConfig configures the service with rate limiting from ratelimit.Config.
// This allows using custom rate limit settings from configuration files.
func WithRateLimiterFromConfig(cfg ratelimit.Config) ServiceOption {
	return func(s *Service) {
		s.rateLimiter = ratelimit.NewWithConfig(cfg)
	}
}

// WithRateLimiterFromAppConfig configures the service with rate limiting from app config.
// This converts config.RateLimitConfig to ratelimit.Config and creates a rate limiter.
func WithRateLimiterFromAppConfig(cfg config.RateLimitConfig) ServiceOption {
	return func(s *Service) {
		rlConfig := ratelimit.Config{
			Gen1: ratelimit.GenerationConfig{
				MinInterval:      cfg.Gen1.MinInterval,
				MaxConcurrent:    cfg.Gen1.MaxConcurrent,
				CircuitThreshold: cfg.Gen1.CircuitThreshold,
			},
			Gen2: ratelimit.GenerationConfig{
				MinInterval:      cfg.Gen2.MinInterval,
				MaxConcurrent:    cfg.Gen2.MaxConcurrent,
				CircuitThreshold: cfg.Gen2.CircuitThreshold,
			},
			Global: ratelimit.GlobalConfig{
				MaxConcurrent:           cfg.Global.MaxConcurrent,
				CircuitOpenDuration:     cfg.Global.CircuitOpenDuration,
				CircuitSuccessThreshold: cfg.Global.CircuitSuccessThreshold,
			},
		}
		s.rateLimiter = ratelimit.NewWithConfig(rlConfig)
	}
}

// WithPluginRegistry configures the service with a plugin registry.
// This enables dispatching commands to plugin-managed devices.
// If not provided, plugin-managed devices will return ErrPluginNotFound.
func WithPluginRegistry(registry *plugins.Registry) ServiceOption {
	return func(s *Service) {
		s.pluginRegistry = registry
	}
}

// New creates a new Shelly service with optional configuration.
func New(resolver DeviceResolver, opts ...ServiceOption) *Service {
	svc := &Service{
		resolver: resolver,
	}
	for _, opt := range opts {
		opt(svc)
	}
	// Initialize firmware service after options are applied (it needs the service for connection handling)
	svc.firmwareService = firmware.NewService(svc)
	if svc.pluginRegistry != nil {
		svc.firmwareService.SetPluginRegistry(svc.pluginRegistry)
	}
	// Initialize wireless service
	svc.wirelessService = wireless.New(svc)
	// Initialize network service
	svc.networkService = network.NewWiFiService(svc)
	return svc
}

// Connect establishes a connection to a device by identifier (name or address).
func (s *Service) Connect(ctx context.Context, identifier string) (*client.Client, error) {
	device, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	return client.Connect(ctx, device)
}

// WithConnection executes a function with a device connection, handling cleanup.
// Rate limiting is automatically applied if configured.
func (s *Service) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	// Resolve device to get address and generation for rate limiting
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return err
	}

	// If no rate limiter, execute directly (backward compatible)
	if s.rateLimiter == nil {
		return s.executeWithConnection(ctx, device, fn)
	}

	// Acquire rate limiter slot
	release, err := s.rateLimiter.Acquire(ctx, device.Address, device.Generation)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = s.executeWithConnection(ctx, device, fn)

	// Record success/failure for circuit breaker
	if err != nil {
		s.rateLimiter.RecordFailure(device.Address)
	} else {
		s.rateLimiter.RecordSuccess(device.Address)
	}

	return err
}

// executeWithConnection performs the actual connection and function execution.
// Includes automatic IP remapping: if connection fails and device has a MAC,
// attempts mDNS discovery to find the device's new IP address.
func (s *Service) executeWithConnection(ctx context.Context, device model.Device, fn func(*client.Client) error) error {
	conn, err := client.Connect(ctx, device)
	if err != nil {
		// Try IP remapping if connection failed and we have a MAC address
		conn, err = s.tryIPRemap(ctx, device, err)
		if err != nil {
			return err
		}
	}
	defer iostreams.CloseWithDebug("closing device connection", conn)

	return fn(conn)
}

// tryIPRemap attempts to remap a device's IP address via mDNS discovery.
// Returns a new connection if remapping succeeds, or the original error if not.
func (s *Service) tryIPRemap(ctx context.Context, device model.Device, originalErr error) (*client.Client, error) {
	// Only attempt remap for connection errors with a known MAC
	if !isConnectionError(originalErr) || device.MAC == "" {
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "connection failed for %s, attempting MAC discovery...", device.Name)

	// Quick mDNS discovery (~2 seconds)
	newIP, discoverErr := s.DiscoverByMAC(ctx, device.MAC)
	if discoverErr != nil {
		iostreams.DebugErr("MAC discovery failed", discoverErr)
		return nil, originalErr
	}
	if newIP == "" || newIP == device.Address {
		// Not found or same IP - return original error
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "found new IP %s for MAC %s, verifying...", newIP, device.MAC)

	// Try connecting to new IP
	deviceCopy := device
	deviceCopy.Address = newIP
	conn, retryErr := client.Connect(ctx, deviceCopy)
	if retryErr != nil {
		iostreams.DebugErr("connection to new IP failed", retryErr)
		return nil, originalErr
	}

	// Verify MAC matches (security check)
	info := conn.Info()
	if info == nil || model.NormalizeMAC(info.MAC) != device.NormalizedMAC() {
		iostreams.DebugCat(iostreams.CategoryDevice, "MAC mismatch: expected %s, got %s", device.NormalizedMAC(), model.NormalizeMAC(info.MAC))
		iostreams.CloseWithDebug("closing mismatched connection", conn)
		return nil, originalErr
	}

	// Success! Update config silently
	if err := config.UpdateDeviceAddress(device.Name, newIP); err != nil {
		iostreams.DebugErr("failed to persist new IP", err)
		// Continue anyway - connection works, just won't persist
	} else {
		iostreams.DebugCat(iostreams.CategoryDevice, "remapped %s: %s -> %s", device.Name, device.Address, newIP)
	}

	return conn, nil
}

// isConnectionError returns true if the error indicates a connection failure
// that could be resolved by trying a different IP address.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, model.ErrConnectionFailed) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "dial tcp")
}

// RawRPC sends a raw RPC command to a device and returns the response.
func (s *Service) RawRPC(ctx context.Context, identifier, method string, params map[string]any) (any, error) {
	var result any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		res, err := conn.Call(ctx, method, params)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

// RawGen1Call sends a raw REST API call to a Gen1 device and returns the response as bytes.
func (s *Service) RawGen1Call(ctx context.Context, identifier, path string) ([]byte, error) {
	var result []byte
	err := s.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		res, err := conn.Call(ctx, path)
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	return result, err
}

// ResolveWithGeneration resolves a device identifier with generation auto-detection.
// If the resolver implements GenerationAwareResolver, it uses that; otherwise falls back to basic resolution.
func (s *Service) ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error) {
	if gar, ok := s.resolver.(GenerationAwareResolver); ok {
		return gar.ResolveWithGeneration(ctx, identifier)
	}
	return s.resolver.Resolve(identifier)
}

// ConnectGen1 establishes a connection to a Gen1 device by identifier.
func (s *Service) ConnectGen1(ctx context.Context, identifier string) (*client.Gen1Client, error) {
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return client.ConnectGen1(ctx, device)
}

// WithGen1Connection executes a function with a Gen1 device connection, handling cleanup.
// Rate limiting is automatically applied if configured.
func (s *Service) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	// Resolve device to get address for rate limiting
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return err
	}

	// If no rate limiter, execute directly (backward compatible)
	if s.rateLimiter == nil {
		return s.executeWithGen1Connection(ctx, device, fn)
	}

	// Gen1 devices always use generation=1 for rate limiting
	release, err := s.rateLimiter.Acquire(ctx, device.Address, 1)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = s.executeWithGen1Connection(ctx, device, fn)

	// Record success/failure for circuit breaker
	if err != nil {
		s.rateLimiter.RecordFailure(device.Address)
	} else {
		s.rateLimiter.RecordSuccess(device.Address)
	}

	return err
}

// executeWithGen1Connection performs the actual Gen1 connection and function execution.
// Includes automatic IP remapping: if connection fails and device has a MAC,
// attempts mDNS discovery to find the device's new IP address.
func (s *Service) executeWithGen1Connection(ctx context.Context, device model.Device, fn func(*client.Gen1Client) error) error {
	conn, err := client.ConnectGen1(ctx, device)
	if err != nil {
		// Try IP remapping if connection failed and we have a MAC address
		conn, err = s.tryGen1IPRemap(ctx, device, err)
		if err != nil {
			return err
		}
	}
	defer iostreams.CloseWithDebug("closing gen1 device connection", conn)

	return fn(conn)
}

// tryGen1IPRemap attempts to remap a Gen1 device's IP address via mDNS discovery.
// Returns a new connection if remapping succeeds, or the original error if not.
func (s *Service) tryGen1IPRemap(ctx context.Context, device model.Device, originalErr error) (*client.Gen1Client, error) {
	// Only attempt remap for connection errors with a known MAC
	if !isConnectionError(originalErr) || device.MAC == "" {
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "Gen1 connection failed for %s, attempting MAC discovery...", device.Name)

	// Quick mDNS discovery (~2 seconds)
	newIP, discoverErr := s.DiscoverByMAC(ctx, device.MAC)
	if discoverErr != nil {
		iostreams.DebugErr("MAC discovery failed", discoverErr)
		return nil, originalErr
	}
	if newIP == "" || newIP == device.Address {
		// Not found or same IP - return original error
		return nil, originalErr
	}

	iostreams.DebugCat(iostreams.CategoryDevice, "found new IP %s for MAC %s, verifying...", newIP, device.MAC)

	// Try connecting to new IP
	deviceCopy := device
	deviceCopy.Address = newIP
	conn, retryErr := client.ConnectGen1(ctx, deviceCopy)
	if retryErr != nil {
		iostreams.DebugErr("Gen1 connection to new IP failed", retryErr)
		return nil, originalErr
	}

	// For Gen1, we can't easily verify MAC from connection info,
	// but the discovery already matched by MAC, so we trust it

	// Success! Update config silently
	if err := config.UpdateDeviceAddress(device.Name, newIP); err != nil {
		iostreams.DebugErr("failed to persist new IP", err)
		// Continue anyway - connection works, just won't persist
	} else {
		iostreams.DebugCat(iostreams.CategoryDevice, "remapped Gen1 %s: %s -> %s", device.Name, device.Address, newIP)
	}

	return conn, nil
}

// IsGen1Device checks if a device is Gen1.
func (s *Service) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	device, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return false, model.Device{}, err
	}
	return device.Generation == 1, device, nil
}

// withGenAwareAction executes gen1Fn for Gen1 devices, gen2Fn for Gen2+ devices.
// This centralizes the generation detection and routing logic.
func (s *Service) withGenAwareAction(
	ctx context.Context,
	identifier string,
	gen1Fn func(*client.Gen1Client) error,
	gen2Fn func(*client.Client) error,
) error {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return err
	}

	if isGen1 {
		return s.WithGen1Connection(ctx, identifier, gen1Fn)
	}
	return s.WithConnection(ctx, identifier, gen2Fn)
}

// WithRateLimitedCall wraps a device operation with rate limiting.
// If no rate limiter is configured, the operation executes directly.
//
// The generation parameter controls rate limiting behavior:
//   - 1: Gen1 limits (1 concurrent, 2s interval - ESP8266 constraints)
//   - 2: Gen2 limits (3 concurrent, 500ms interval - ESP32 constraints)
//   - 0: Unknown, treated as Gen1 for safety
//
// Returns ErrCircuitOpen if the device's circuit breaker is open.
func (s *Service) WithRateLimitedCall(ctx context.Context, address string, generation int, fn func() error) error {
	// No rate limiter configured - execute directly (backward compatible)
	if s.rateLimiter == nil {
		return fn()
	}

	// Acquire rate limiter slot
	release, err := s.rateLimiter.Acquire(ctx, address, generation)
	if err != nil {
		return err
	}
	defer release()

	// Execute the operation
	err = fn()

	// Record success/failure for circuit breaker
	if err != nil {
		s.rateLimiter.RecordFailure(address)
	} else {
		s.rateLimiter.RecordSuccess(address)
	}

	return err
}

// RateLimiter returns the service's rate limiter, if configured.
// Returns nil if no rate limiting is enabled.
func (s *Service) RateLimiter() *ratelimit.DeviceRateLimiter {
	return s.rateLimiter
}

// SetDeviceGeneration updates the rate limiter's generation info for a device.
// This is useful after auto-detection to optimize rate limiting.
// No-op if rate limiting is not enabled.
func (s *Service) SetDeviceGeneration(address string, generation int) {
	if s.rateLimiter != nil {
		s.rateLimiter.SetGeneration(address, generation)
	}
}

// PluginRegistry returns the service's plugin registry, if configured.
// Returns nil if no plugin registry is enabled.
func (s *Service) PluginRegistry() *plugins.Registry {
	return s.pluginRegistry
}

// SetPluginRegistry sets the plugin registry after service creation.
// This is useful when the registry needs to be set up after the service is created.
func (s *Service) SetPluginRegistry(registry *plugins.Registry) {
	s.pluginRegistry = registry
	if s.firmwareService != nil {
		s.firmwareService.SetPluginRegistry(registry)
	}
}

// FirmwareService returns the firmware service for direct access.
func (s *Service) FirmwareService() *firmware.Service {
	return s.firmwareService
}

// Wireless returns the wireless service for Zigbee, BTHome, LoRa, and Matter operations.
func (s *Service) Wireless() *wireless.Service {
	return s.wirelessService
}

// ----- Firmware type aliases for backward compatibility -----

// FirmwareInfo is an alias to firmware.Info for backward compatibility.
type FirmwareInfo = firmware.Info

// FirmwareStatus is an alias to firmware.Status for backward compatibility.
type FirmwareStatus = firmware.Status

// FirmwareCheckResult is an alias to firmware.CheckResult for backward compatibility.
type FirmwareCheckResult = firmware.CheckResult

// DeviceUpdateStatus is an alias to firmware.DeviceUpdateStatus for backward compatibility.
type DeviceUpdateStatus = firmware.DeviceUpdateStatus

// UpdateOpts is an alias to firmware.UpdateOpts for backward compatibility.
type UpdateOpts = firmware.UpdateOpts

// UpdateResult is an alias to firmware.UpdateResult for backward compatibility.
type UpdateResult = firmware.UpdateResult

// FirmwareUpdateEntry is an alias to firmware.UpdateEntry for backward compatibility.
type FirmwareUpdateEntry = firmware.UpdateEntry

// FirmwareCache is an alias to firmware.Cache for backward compatibility.
type FirmwareCache = firmware.Cache

// FirmwareCacheEntry is an alias to firmware.CacheEntry for backward compatibility.
type FirmwareCacheEntry = firmware.CacheEntry

// NewFirmwareCache creates a new firmware cache (delegated for backward compatibility).
func NewFirmwareCache() *FirmwareCache {
	return firmware.NewCache()
}

// ----- Firmware helper functions for backward compatibility -----

// BuildFirmwareUpdateList delegates to firmware.BuildUpdateList.
func BuildFirmwareUpdateList(results []FirmwareCheckResult, devices map[string]model.Device) []FirmwareUpdateEntry {
	return firmware.BuildUpdateList(results, devices)
}

// FilterDevicesByNameAndPlatform delegates to firmware.FilterDevicesByNameAndPlatform.
func FilterDevicesByNameAndPlatform(devices map[string]model.Device, devicesList, platform string) map[string]model.Device {
	return firmware.FilterDevicesByNameAndPlatform(devices, devicesList, platform)
}

// FilterEntriesByStage delegates to firmware.FilterEntriesByStage.
func FilterEntriesByStage(entries []FirmwareUpdateEntry, beta bool) []FirmwareUpdateEntry {
	return firmware.FilterEntriesByStage(entries, beta)
}

// AnyHasBeta delegates to firmware.AnyHasBeta.
func AnyHasBeta(entries []FirmwareUpdateEntry) bool {
	return firmware.AnyHasBeta(entries)
}

// SelectEntriesByStage delegates to firmware.SelectEntriesByStage.
func SelectEntriesByStage(entries []FirmwareUpdateEntry, beta bool) (indices []int, stage string) {
	return firmware.SelectEntriesByStage(entries, beta)
}

// GetEntriesByIndices delegates to firmware.GetEntriesByIndices.
func GetEntriesByIndices(entries []FirmwareUpdateEntry, indices []int) []FirmwareUpdateEntry {
	return firmware.GetEntriesByIndices(entries, indices)
}

// ToDeviceUpdateStatuses delegates to firmware.ToDeviceUpdateStatuses.
func ToDeviceUpdateStatuses(entries []FirmwareUpdateEntry) []DeviceUpdateStatus {
	return firmware.ToDeviceUpdateStatuses(entries)
}

// ----- Firmware Service method delegations for backward compatibility -----

// CheckFirmware delegates to the firmware service.
func (s *Service) CheckFirmware(ctx context.Context, identifier string) (*FirmwareInfo, error) {
	return s.firmwareService.Check(ctx, identifier)
}

// GetFirmwareStatus delegates to the firmware service.
func (s *Service) GetFirmwareStatus(ctx context.Context, identifier string) (*FirmwareStatus, error) {
	return s.firmwareService.GetStatus(ctx, identifier)
}

// UpdateFirmware delegates to the firmware service.
func (s *Service) UpdateFirmware(ctx context.Context, identifier string, opts *libfirmware.UpdateOptions) error {
	return s.firmwareService.Update(ctx, identifier, opts)
}

// UpdateFirmwareStable delegates to the firmware service.
func (s *Service) UpdateFirmwareStable(ctx context.Context, identifier string) error {
	return s.firmwareService.UpdateStable(ctx, identifier)
}

// UpdateFirmwareBeta delegates to the firmware service.
func (s *Service) UpdateFirmwareBeta(ctx context.Context, identifier string) error {
	return s.firmwareService.UpdateBeta(ctx, identifier)
}

// UpdateFirmwareFromURL delegates to the firmware service.
func (s *Service) UpdateFirmwareFromURL(ctx context.Context, identifier, url string) error {
	return s.firmwareService.UpdateFromURL(ctx, identifier, url)
}

// RollbackFirmware delegates to the firmware service.
func (s *Service) RollbackFirmware(ctx context.Context, identifier string) error {
	return s.firmwareService.Rollback(ctx, identifier)
}

// GetFirmwareURL delegates to the firmware service.
func (s *Service) GetFirmwareURL(ctx context.Context, identifier, stage string) (string, error) {
	return s.firmwareService.GetURL(ctx, identifier, stage)
}

// CheckFirmwareAll delegates to the firmware service.
func (s *Service) CheckFirmwareAll(ctx context.Context, ios *iostreams.IOStreams, devices []string) []FirmwareCheckResult {
	return s.firmwareService.CheckAll(ctx, ios, devices)
}

// CheckFirmwareAllPlatforms delegates to the firmware service.
func (s *Service) CheckFirmwareAllPlatforms(ctx context.Context, ios *iostreams.IOStreams, deviceConfigs map[string]model.Device) []FirmwareCheckResult {
	return s.firmwareService.CheckAllPlatforms(ctx, ios, deviceConfigs)
}

// CheckPluginFirmware delegates to the firmware service.
func (s *Service) CheckPluginFirmware(ctx context.Context, device model.Device) (*FirmwareInfo, error) {
	return s.firmwareService.CheckPlugin(ctx, device)
}

// UpdatePluginFirmware delegates to the firmware service.
func (s *Service) UpdatePluginFirmware(ctx context.Context, device model.Device, stage, url string) error {
	return s.firmwareService.UpdatePlugin(ctx, device, stage, url)
}

// UpdateDeviceFirmware delegates to the firmware service.
func (s *Service) UpdateDeviceFirmware(ctx context.Context, device model.Device, useBeta bool, customURL string) error {
	return s.firmwareService.UpdateDevice(ctx, device, useBeta, customURL)
}

// CheckDeviceFirmware delegates to the firmware service.
func (s *Service) CheckDeviceFirmware(ctx context.Context, device model.Device) (*FirmwareInfo, error) {
	return s.firmwareService.CheckDevice(ctx, device)
}

// CheckDevicesForUpdates delegates to the firmware service.
func (s *Service) CheckDevicesForUpdates(ctx context.Context, ios *iostreams.IOStreams, devices []string, staged int) []DeviceUpdateStatus {
	return s.firmwareService.CheckDevicesForUpdates(ctx, ios, devices, staged)
}

// UpdateDevices delegates to the firmware service.
func (s *Service) UpdateDevices(ctx context.Context, ios *iostreams.IOStreams, devices []DeviceUpdateStatus, opts UpdateOpts) []UpdateResult {
	return s.firmwareService.UpdateDevices(ctx, ios, devices, opts)
}

// PrefetchFirmwareCache delegates to the firmware service.
func (s *Service) PrefetchFirmwareCache(ctx context.Context, ios *iostreams.IOStreams) {
	s.firmwareService.Prefetch(ctx, ios)
}

// GetCachedFirmware delegates to the firmware service.
func (s *Service) GetCachedFirmware(ctx context.Context, deviceName string, maxAge time.Duration) *FirmwareCacheEntry {
	return s.firmwareService.GetCached(ctx, deviceName, maxAge)
}

// FirmwareCache returns the firmware cache.
func (s *Service) FirmwareCache() *FirmwareCache {
	return s.firmwareService.Cache()
}

// GetWiFiStatusFull gets the full WiFi status from a device.
// Delegates to the network service.
func (s *Service) GetWiFiStatusFull(ctx context.Context, identifier string) (*network.WiFiStatusFull, error) {
	return s.networkService.GetStatusFull(ctx, identifier)
}

// GetWiFiConfigFull gets the full WiFi configuration from a device.
// Delegates to the network service.
func (s *Service) GetWiFiConfigFull(ctx context.Context, identifier string) (*network.WiFiConfigFull, error) {
	return s.networkService.GetConfigFull(ctx, identifier)
}

// ScanWiFiNetworksFull scans for available WiFi networks with full details.
// Delegates to the network service.
func (s *Service) ScanWiFiNetworksFull(ctx context.Context, identifier string) ([]network.WiFiNetworkFull, error) {
	return s.networkService.ScanNetworksFull(ctx, identifier)
}

// SetWiFiStation configures the primary WiFi station.
// Delegates to the network service.
func (s *Service) SetWiFiStation(ctx context.Context, identifier, ssid, password string, enable bool) error {
	return s.networkService.SetStation(ctx, identifier, ssid, password, enable)
}

// SetWiFiAP configures the access point.
// Delegates to the network service.
func (s *Service) SetWiFiAP(ctx context.Context, identifier, ssid, password string, enable bool) error {
	return s.networkService.SetAP(ctx, identifier, ssid, password, enable)
}
