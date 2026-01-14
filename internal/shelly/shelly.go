// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	libfirmware "github.com/tj-smith47/shelly-go/firmware"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
	"github.com/tj-smith47/shelly-cli/internal/shelly/component"
	"github.com/tj-smith47/shelly-cli/internal/shelly/connection"
	"github.com/tj-smith47/shelly-cli/internal/shelly/device"
	"github.com/tj-smith47/shelly-cli/internal/shelly/firmware"
	"github.com/tj-smith47/shelly-cli/internal/shelly/modbus"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
	"github.com/tj-smith47/shelly-cli/internal/shelly/provision"
	"github.com/tj-smith47/shelly-cli/internal/shelly/wireless"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// DefaultTimeout is the default timeout for device operations.
const DefaultTimeout = 10 * time.Second

// Service provides high-level operations on Shelly devices.
type Service struct {
	resolver          DeviceResolver
	connManager       *connection.Manager
	rateLimiter       *ratelimit.DeviceRateLimiter
	pluginRegistry    *plugins.Registry
	cache             *cache.FileCache
	ios               *iostreams.IOStreams
	firmwareService   *firmware.Service
	wirelessService   *wireless.Service
	networkService    *network.WiFiService
	mqttService       *network.MQTTService
	ethernetService   *network.EthernetService
	deviceService     *device.Service
	componentService  *component.Service
	authService       *auth.Service
	modbusService     *modbus.Service
	provisionService  *provision.Service
	monitoringService *monitoring.Service
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
// If not provided, no rate limiting is applied.
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

// WithFileCache configures the service with a file cache for caching responses.
// This enables automatic cache invalidation after mutations.
func WithFileCache(fc *cache.FileCache) ServiceOption {
	return func(s *Service) {
		s.cache = fc
	}
}

// WithIOStreams configures the service with IOStreams for debug logging.
// This enables logging cache invalidation errors in debug mode.
func WithIOStreams(ios *iostreams.IOStreams) ServiceOption {
	return func(s *Service) {
		s.ios = ios
	}
}

// invalidateCache invalidates cached data for a device/type after mutations.
// Errors are logged but not returned (cache invalidation is best-effort).
func (s *Service) invalidateCache(target, dataType string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.Invalidate(target, dataType); err != nil && s.ios != nil {
		s.ios.DebugErr("cache invalidate "+target+"/"+dataType, err)
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
	// Initialize connection manager with rate limiting
	// Service implements connection.Resolver (via ResolveWithGeneration)
	// Service implements connection.Discoverer (via DiscoverByMAC)
	var connOpts []connection.Option
	if svc.rateLimiter != nil {
		connOpts = append(connOpts, connection.WithRateLimiter(svc.rateLimiter))
	}
	svc.connManager = connection.NewManager(&resolverAdapter{svc}, svc, connOpts...)
	// Initialize firmware service after options are applied (it needs the service for connection handling)
	svc.firmwareService = firmware.NewService(svc)
	if svc.pluginRegistry != nil {
		svc.firmwareService.SetPluginRegistry(svc.pluginRegistry)
	}
	// Initialize wireless service
	svc.wirelessService = wireless.New(svc)
	// Initialize network services
	svc.networkService = network.NewWiFiService(svc)
	svc.mqttService = network.NewMQTTService(svc)
	svc.ethernetService = network.NewEthernetService(svc)
	// Initialize device service
	svc.deviceService = device.New(svc)
	// Initialize component service using adapter
	svc.componentService = component.New(&componentAdapter{svc})
	// Initialize auth service with adapter for device info
	svc.authService = auth.New(svc, &authAdapter{svc})
	// Initialize modbus service
	svc.modbusService = modbus.New(svc)
	// Initialize provision service
	svc.provisionService = provision.New(svc)
	return svc
}

// componentAdapter adapts shelly.Service to implement component.ConnectionProvider.
// This is necessary because shelly.WithDevice uses func(*connection.DeviceClient) but
// component.ConnectionProvider expects func(component.DeviceClient).
type componentAdapter struct {
	*Service
}

// WithDevice adapts the function signature for component.ConnectionProvider.
func (a *componentAdapter) WithDevice(ctx context.Context, identifier string, fn func(component.DeviceClient) error) error {
	return a.Service.WithDevice(ctx, identifier, func(dev *connection.DeviceClient) error {
		return fn(dev) // *connection.DeviceClient implements component.DeviceClient
	})
}

// authAdapter adapts shelly.Service to implement auth.DeviceInfoProvider.
type authAdapter struct {
	*Service
}

// GetAuthEnabled implements auth.DeviceInfoProvider.
func (a *authAdapter) GetAuthEnabled(ctx context.Context, identifier string) (bool, error) {
	info, err := a.DeviceInfo(ctx, identifier)
	if err != nil {
		return false, err
	}
	return info.AuthEn, nil
}

// resolverAdapter adapts shelly.Service to implement connection.Resolver.
type resolverAdapter struct {
	*Service
}

// ResolveWithGeneration implements connection.Resolver.
func (a *resolverAdapter) ResolveWithGeneration(ctx context.Context, identifier string) (model.Device, error) {
	return a.Service.ResolveWithGeneration(ctx, identifier)
}

// DeviceClient is a type alias to connection.DeviceClient for convenience.
type DeviceClient = connection.DeviceClient

// Connect establishes a connection to a device by identifier (name or address).
func (s *Service) Connect(ctx context.Context, identifier string) (*client.Client, error) {
	dev, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, err
	}

	return client.Connect(ctx, dev)
}

// WithConnection executes a function with a device connection, handling cleanup.
// Rate limiting is automatically applied if configured.
// Delegates to the connection manager.
func (s *Service) WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error {
	return s.connManager.WithConnection(ctx, identifier, fn)
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

// RawHTTPCall makes a raw HTTP GET request to a device at the given path.
// This works for all device generations and returns the raw response body.
// Use this for generation-agnostic HTTP calls like /status (Gen1) or /rpc/Method (Gen2+).
func (s *Service) RawHTTPCall(ctx context.Context, identifier, path string) ([]byte, error) {
	dev, err := s.resolver.Resolve(identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device: %w", err)
	}

	addr := dev.Address
	if addr == "" {
		return nil, fmt.Errorf("device %q has no address", identifier)
	}

	// Build URL - ensure path starts with /
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	url := fmt.Sprintf("http://%s%s", addr, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth if device has credentials
	if dev.HasAuth() {
		req.SetBasicAuth(dev.Auth.Username, dev.Auth.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && s.ios != nil {
			s.ios.DebugErr("close response body", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("HTTP %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
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
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return client.ConnectGen1(ctx, dev)
}

// WithGen1Connection executes a function with a Gen1 device connection, handling cleanup.
// Rate limiting is automatically applied if configured.
// Delegates to the connection manager.
func (s *Service) WithGen1Connection(ctx context.Context, identifier string, fn func(*client.Gen1Client) error) error {
	return s.connManager.WithGen1Connection(ctx, identifier, fn)
}

// IsGen1Device checks if a device is Gen1.
func (s *Service) IsGen1Device(ctx context.Context, identifier string) (bool, model.Device, error) {
	dev, err := s.ResolveWithGeneration(ctx, identifier)
	if err != nil {
		return false, model.Device{}, err
	}
	return dev.Generation == 1, dev, nil
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

// WithDevice executes a function with a unified device connection.
// The DeviceClient auto-detects the device generation and provides
// access to both Gen1 and Gen2 APIs through a single interface.
// Delegates to the connection manager.
func (s *Service) WithDevice(ctx context.Context, identifier string, fn func(*DeviceClient) error) error {
	return s.connManager.WithDevice(ctx, identifier, fn)
}

// WithDevices executes a function for multiple devices concurrently.
// Each device gets its own DeviceClient with auto-detected generation.
// Errors are collected and returned as a combined error.
// Delegates to the connection manager.
func (s *Service) WithDevices(ctx context.Context, devices []string, concurrency int, fn func(device string, dev *DeviceClient) error) error {
	return s.connManager.WithDevices(ctx, devices, concurrency, fn)
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
	// No rate limiter configured - execute directly
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
	s.recordCircuitResult(ctx, address, err)

	return err
}

// recordCircuitResult records success/failure for the circuit breaker.
// Only counts actual connectivity failures, not expected API responses.
// Skips failure recording for polling requests (BUG-015) - polling failures
// shouldn't block user-initiated actions like toggles.
func (s *Service) recordCircuitResult(ctx context.Context, address string, err error) {
	if err == nil {
		s.rateLimiter.RecordSuccess(address)
		return
	}

	if !ratelimit.IsConnectivityFailure(err) {
		debug.TraceEvent("circuit: Ignoring non-connectivity error for %s: %v", address, err)
		s.rateLimiter.RecordSuccess(address)
		return
	}

	// Connectivity failure - skip for polling, record for user actions
	if ratelimit.IsPolling(ctx) {
		debug.TraceEvent("circuit: Skipping RecordFailure for polling request %s: %v", address, err)
		return
	}

	debug.TraceEvent("circuit: RecordFailure for %s: %v", address, err)
	s.rateLimiter.RecordFailure(address)
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

// ----- Firmware type aliases for convenience -----

// FirmwareInfo is an alias to firmware.Info for convenience.
type FirmwareInfo = firmware.Info

// FirmwareStatus is an alias to firmware.Status for convenience.
type FirmwareStatus = firmware.Status

// FirmwareCheckResult is an alias to firmware.CheckResult for convenience.
type FirmwareCheckResult = firmware.CheckResult

// DeviceUpdateStatus is an alias to firmware.DeviceUpdateStatus for convenience.
type DeviceUpdateStatus = firmware.DeviceUpdateStatus

// UpdateOpts is an alias to firmware.UpdateOpts for convenience.
type UpdateOpts = firmware.UpdateOpts

// UpdateResult is an alias to firmware.UpdateResult for convenience.
type UpdateResult = firmware.UpdateResult

// FirmwareUpdateEntry is an alias to firmware.UpdateEntry for convenience.
type FirmwareUpdateEntry = firmware.UpdateEntry

// FirmwareCache is an alias to firmware.Cache for convenience.
type FirmwareCache = firmware.Cache

// FirmwareCacheEntry is an alias to firmware.CacheEntry for convenience.
type FirmwareCacheEntry = firmware.CacheEntry

// NewFirmwareCache creates a new firmware cache (delegated for convenience).
func NewFirmwareCache() *FirmwareCache {
	return firmware.NewCache()
}

// ----- Firmware helper functions for convenience -----

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

// ----- Firmware Service method delegations for convenience -----

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
func (s *Service) CheckPluginFirmware(ctx context.Context, dev model.Device) (*FirmwareInfo, error) {
	return s.firmwareService.CheckPlugin(ctx, dev)
}

// UpdatePluginFirmware delegates to the firmware service.
func (s *Service) UpdatePluginFirmware(ctx context.Context, dev model.Device, stage, url string) error {
	return s.firmwareService.UpdatePlugin(ctx, dev, stage, url)
}

// UpdateDeviceFirmware delegates to the firmware service.
func (s *Service) UpdateDeviceFirmware(ctx context.Context, dev model.Device, useBeta bool, customURL string) error {
	return s.firmwareService.UpdateDevice(ctx, dev, useBeta, customURL)
}

// CheckDeviceFirmware delegates to the firmware service.
func (s *Service) CheckDeviceFirmware(ctx context.Context, dev model.Device) (*FirmwareInfo, error) {
	return s.firmwareService.CheckDevice(ctx, dev)
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

// ----- Device Service accessor and delegations -----

// DeviceService returns the device service for direct access.
func (s *Service) DeviceService() *device.Service {
	return s.deviceService
}

// GetFullStatus returns the complete device status from Shelly.GetStatus (Gen2+).
// Use this for Gen2+ devices only. For auto-detection, use GetFullStatusAuto.
func (s *Service) GetFullStatus(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	return s.deviceService.GetFullStatus(ctx, identifier)
}

// GetFullStatusGen1 returns the complete device status from /status endpoint (Gen1).
// Use this for Gen1 devices only. For auto-detection, use GetFullStatusAuto.
func (s *Service) GetFullStatusGen1(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	return s.deviceService.GetFullStatusGen1(ctx, identifier)
}

// GetFullStatusAuto returns the complete device status, auto-detecting the device generation.
// This calls Shelly.GetStatus for Gen2+ devices or /status for Gen1 devices.
func (s *Service) GetFullStatusAuto(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		return s.deviceService.GetFullStatusGen1(ctx, identifier)
	}
	return s.deviceService.GetFullStatus(ctx, identifier)
}

// GetFullConfig returns the complete device config from Shelly.GetConfig (Gen2+).
func (s *Service) GetFullConfig(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	return s.deviceService.GetFullConfig(ctx, identifier)
}

// GetFullConfigGen1 returns the complete device config from /settings endpoint (Gen1).
func (s *Service) GetFullConfigGen1(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	return s.deviceService.GetFullConfigGen1(ctx, identifier)
}

// GetFullConfigAuto returns the complete device config, auto-detecting the device generation.
func (s *Service) GetFullConfigAuto(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	isGen1, _, err := s.IsGen1Device(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if isGen1 {
		return s.deviceService.GetFullConfigGen1(ctx, identifier)
	}
	return s.deviceService.GetFullConfig(ctx, identifier)
}

// ----- Component Service accessor and delegations -----

// ComponentService returns the component service for direct access.
func (s *Service) ComponentService() *component.Service {
	return s.componentService
}

// ----- MQTT Service accessor and delegations -----

// MQTTService returns the MQTT service for direct access.
func (s *Service) MQTTService() *network.MQTTService {
	return s.mqttService
}

// GetMQTTStatus delegates to the MQTT service for convenience.
func (s *Service) GetMQTTStatus(ctx context.Context, identifier string) (*network.MQTTStatus, error) {
	return s.mqttService.GetStatus(ctx, identifier)
}

// GetMQTTConfig delegates to the MQTT service for convenience.
func (s *Service) GetMQTTConfig(ctx context.Context, identifier string) (map[string]any, error) {
	cfg, err := s.mqttService.GetConfig(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"enable":       cfg.Enable,
		"server":       cfg.Server,
		"user":         cfg.User,
		"client_id":    cfg.ClientID,
		"topic_prefix": cfg.TopicPrefix,
		"ssl_ca":       cfg.SSLCA,
		"rpc_ntf":      cfg.RPCNTF,
		"status_ntf":   cfg.StatusNTF,
	}, nil
}

// SetMQTTConfig delegates to the MQTT service for convenience.
func (s *Service) SetMQTTConfig(ctx context.Context, identifier string, enable *bool, server, user, password, topicPrefix string) error {
	err := s.mqttService.SetConfig(ctx, identifier, network.SetConfigParams{
		Enable:      enable,
		Server:      server,
		User:        user,
		Password:    password,
		TopicPrefix: topicPrefix,
	})
	if err == nil {
		s.invalidateCache(identifier, cache.TypeMQTT)
	}
	return err
}

// MQTTSetConfigParams is an alias for network.SetConfigParams.
type MQTTSetConfigParams = network.SetConfigParams

// SetMQTTConfigFull delegates to the MQTT service with full configuration options.
func (s *Service) SetMQTTConfigFull(ctx context.Context, identifier string, params MQTTSetConfigParams) error {
	err := s.mqttService.SetConfig(ctx, identifier, params)
	if err == nil {
		s.invalidateCache(identifier, cache.TypeMQTT)
	}
	return err
}

// ----- Ethernet Service accessor and delegations -----

// EthernetService returns the Ethernet service for direct access.
func (s *Service) EthernetService() *network.EthernetService {
	return s.ethernetService
}

// GetEthernetStatus delegates to the Ethernet service for convenience.
func (s *Service) GetEthernetStatus(ctx context.Context, identifier string) (*network.EthernetStatus, error) {
	return s.ethernetService.GetStatus(ctx, identifier)
}

// GetEthernetConfig delegates to the Ethernet service for convenience.
func (s *Service) GetEthernetConfig(ctx context.Context, identifier string) (map[string]any, error) {
	cfg, err := s.ethernetService.GetConfig(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"enable":     cfg.Enable,
		"ipv4mode":   cfg.IPv4Mode,
		"ip":         cfg.IP,
		"netmask":    cfg.Netmask,
		"gw":         cfg.GW,
		"nameserver": cfg.Nameserver,
	}, nil
}

// SetEthernetConfig delegates to the Ethernet service for convenience.
func (s *Service) SetEthernetConfig(ctx context.Context, identifier string, enable *bool, ipv4Mode, ip, netmask, gw, nameserver string) error {
	return s.ethernetService.SetConfig(ctx, identifier, network.EthernetSetConfigParams{
		Enable:     enable,
		IPv4Mode:   ipv4Mode,
		IP:         ip,
		Netmask:    netmask,
		GW:         gw,
		Nameserver: nameserver,
	})
}

// ----- Auth Service accessor and delegations -----

// AuthService returns the Auth service for direct access.
func (s *Service) AuthService() *auth.Service {
	return s.authService
}

// GetAuthStatus delegates to the Auth service for convenience.
func (s *Service) GetAuthStatus(ctx context.Context, identifier string) (*auth.Status, error) {
	return s.authService.GetStatus(ctx, identifier)
}

// SetAuth delegates to the Auth service for convenience.
func (s *Service) SetAuth(ctx context.Context, identifier, user, realm, password string) error {
	return s.authService.Set(ctx, identifier, user, realm, password)
}

// DisableAuth delegates to the Auth service for convenience.
func (s *Service) DisableAuth(ctx context.Context, identifier string) error {
	return s.authService.Disable(ctx, identifier)
}

// ----- Modbus Service accessor and delegations -----

// ModbusService returns the Modbus service for direct access.
func (s *Service) ModbusService() *modbus.Service {
	return s.modbusService
}

// GetModbusStatus delegates to the Modbus service for convenience.
func (s *Service) GetModbusStatus(ctx context.Context, identifier string) (*modbus.Status, error) {
	return s.modbusService.GetStatus(ctx, identifier)
}

// GetModbusConfig delegates to the Modbus service for convenience.
func (s *Service) GetModbusConfig(ctx context.Context, identifier string) (map[string]any, error) {
	cfg, err := s.modbusService.GetConfig(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"enable": cfg.Enable,
	}, nil
}

// SetModbusConfig delegates to the Modbus service for convenience.
func (s *Service) SetModbusConfig(ctx context.Context, identifier string, enable bool) error {
	return s.modbusService.SetConfig(ctx, identifier, enable)
}

// ----- Provision Service accessor and delegations -----

// ProvisionService returns the Provision service for direct access.
func (s *Service) ProvisionService() *provision.Service {
	return s.provisionService
}

// GetDeviceInfoByAddress delegates to the Provision service for convenience.
func (s *Service) GetDeviceInfoByAddress(ctx context.Context, address string) (*provision.DeviceInfo, error) {
	return s.provisionService.GetDeviceInfoByAddress(ctx, address)
}

// ConfigureWiFi delegates to the Provision service for convenience.
func (s *Service) ConfigureWiFi(ctx context.Context, address, ssid, password string) error {
	return s.provisionService.ConfigureWiFi(ctx, address, ssid, password)
}

// GetBTHomeStatus delegates to the Provision service for convenience.
func (s *Service) GetBTHomeStatus(ctx context.Context, identifier string) (*provision.BTHomeDiscovery, error) {
	return s.provisionService.GetBTHomeStatus(ctx, identifier)
}

// StartBTHomeDiscovery delegates to the Provision service for convenience.
func (s *Service) StartBTHomeDiscovery(ctx context.Context, identifier string, duration int) error {
	return s.provisionService.StartBTHomeDiscovery(ctx, identifier, duration)
}

// ----- Type aliases for convenience -----

// MQTTStatus is an alias for network.MQTTStatus.
type MQTTStatus = network.MQTTStatus

// EthernetStatus is an alias for network.EthernetStatus.
type EthernetStatus = network.EthernetStatus

// AuthStatus is an alias for auth.Status.
type AuthStatus = auth.Status

// ModbusStatus is an alias for modbus.Status.
type ModbusStatus = modbus.Status

// BTHomeDiscovery is an alias for provision.BTHomeDiscovery.
type BTHomeDiscovery = provision.BTHomeDiscovery

// ProvisioningDeviceInfo is an alias for provision.DeviceInfo.
type ProvisioningDeviceInfo = provision.DeviceInfo

// ExtractWiFiSSID extracts the station SSID from a raw WiFi.GetConfig result.
// Re-exported from provision package for convenience.
func ExtractWiFiSSID(rawResult any) string {
	return provision.ExtractWiFiSSID(rawResult)
}
