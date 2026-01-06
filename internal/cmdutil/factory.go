// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
// It follows the gh CLI and kubectl patterns for dependency injection and command helpers.
package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/shelly/modbus"
	"github.com/tj-smith47/shelly-cli/internal/shelly/sensoraddon"
)

// Factory provides dependencies to commands through lazy initialization.
// This pattern allows for easy testing by replacing the factory functions
// with mocks, while also avoiding initialization until dependencies are needed.
type Factory struct {
	// IOStreams provides access to stdin/stdout/stderr and terminal capabilities.
	IOStreams func() *iostreams.IOStreams

	// ConfigManager provides the config manager for all config operations.
	// Use this for mutations (RegisterDevice, CreateGroup, etc.).
	ConfigManager func() (*config.Manager, error)

	// Config provides access to the CLI configuration data.
	// This is a convenience that calls ConfigManager().Get().
	// Use ConfigManager() directly for mutations.
	Config func() (*config.Config, error)

	// ShellyService provides the business logic service for device operations.
	// Rate limiting is automatically applied from config to prevent device overload.
	ShellyService func() *shelly.Service

	// Browser provides cross-platform URL opening capabilities.
	Browser func() browser.Browser

	// Cached instances - set after first call to avoid re-initialization.
	ioStreams          *iostreams.IOStreams
	cfgMgr             *config.Manager
	shellyService      *shelly.Service
	automationService  *automation.Service
	kvsService         *kvs.Service
	modbusService      *modbus.Service
	sensorAddonService *sensoraddon.Service
	browserInst        browser.Browser
	fileCache          *cache.FileCache
}

// NewFactory creates a Factory with production dependencies.
// Dependencies are lazily initialized on first access.
func NewFactory() *Factory {
	f := &Factory{}

	f.IOStreams = func() *iostreams.IOStreams {
		if f.ioStreams == nil {
			f.ioStreams = iostreams.System()
		}
		return f.ioStreams
	}

	f.ConfigManager = func() (*config.Manager, error) {
		if f.cfgMgr == nil {
			f.cfgMgr = config.NewManager("")
			if err := f.cfgMgr.Load(); err != nil {
				return nil, err
			}
		}
		return f.cfgMgr, nil
	}

	f.Config = func() (*config.Config, error) {
		mgr, err := f.ConfigManager()
		if err != nil {
			return nil, err
		}
		return mgr.Get(), nil
	}

	f.ShellyService = func() *shelly.Service {
		if f.shellyService == nil {
			f.shellyService = f.createShellyService()
		}
		return f.shellyService
	}

	f.Browser = func() browser.Browser {
		if f.browserInst == nil {
			f.browserInst = browser.New()
		}
		return f.browserInst
	}

	return f
}

// NewWithIOStreams creates a Factory with a custom IOStreams instance.
// Useful for testing or when you need to override standard I/O.
func NewWithIOStreams(ios *iostreams.IOStreams) *Factory {
	f := NewFactory()
	f.ioStreams = ios
	f.IOStreams = func() *iostreams.IOStreams {
		return ios
	}
	return f
}

// SetIOStreams sets a custom IOStreams instance on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetIOStreams(ios *iostreams.IOStreams) *Factory {
	f.ioStreams = ios
	origIOStreams := f.IOStreams
	f.IOStreams = func() *iostreams.IOStreams {
		if f.ioStreams != nil {
			return f.ioStreams
		}
		return origIOStreams()
	}
	return f
}

// SetConfigManager sets a custom config manager on an existing factory.
// This modifies the factory in-place and returns it for chaining.
// Use this for testing with isolated config.
func (f *Factory) SetConfigManager(mgr *config.Manager) *Factory {
	f.cfgMgr = mgr
	origMgr := f.ConfigManager
	f.ConfigManager = func() (*config.Manager, error) {
		if f.cfgMgr != nil {
			return f.cfgMgr, nil
		}
		return origMgr()
	}
	// Update Config to use the new manager
	f.Config = func() (*config.Config, error) {
		mgr, err := f.ConfigManager()
		if err != nil {
			return nil, err
		}
		return mgr.Get(), nil
	}
	return f
}

// SetShellyService sets a custom shelly service on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetShellyService(svc *shelly.Service) *Factory {
	f.shellyService = svc
	origService := f.ShellyService
	f.ShellyService = func() *shelly.Service {
		if f.shellyService != nil {
			return f.shellyService
		}
		return origService()
	}
	return f
}

// MustConfig returns the configuration, panicking on error.
// Use this only when config errors should be fatal (e.g., during initialization).
func (f *Factory) MustConfig() *config.Config {
	cfg, err := f.Config()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	return cfg
}

// MustConfigManager returns the config manager, panicking on error.
// Use this only when config errors should be fatal (e.g., during initialization).
func (f *Factory) MustConfigManager() *config.Manager {
	mgr, err := f.ConfigManager()
	if err != nil {
		panic("failed to load config manager: " + err.Error())
	}
	return mgr
}

// SetBrowser sets a custom browser instance on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetBrowser(b browser.Browser) *Factory {
	f.browserInst = b
	origBrowser := f.Browser
	f.Browser = func() browser.Browser {
		if f.browserInst != nil {
			return f.browserInst
		}
		return origBrowser()
	}
	return f
}

// createShellyService creates the Shelly service with all options.
// Extracted to reduce nesting complexity in the closure.
func (f *Factory) createShellyService() *shelly.Service {
	// Cache and IOStreams enable automatic cache invalidation on mutations
	opts := []shelly.ServiceOption{
		shelly.WithFileCache(f.FileCache()),
		shelly.WithIOStreams(f.IOStreams()),
	}

	// Test mode: simple service without plugin support
	// Avoids creating real directories during tests
	if config.IsTestFs() {
		return shelly.New(shelly.NewConfigResolver(), opts...)
	}

	// Production mode: add rate limiting and plugins
	// Rate limiting prevents device overload
	// Plugin support enables control of non-Shelly devices (Tasmota, ESPHome, etc.)
	if cfg := config.Get(); cfg != nil {
		opts = append(opts, shelly.WithRateLimiterFromAppConfig(cfg.GetRateLimitConfig()))
	}
	if registry, err := plugins.NewRegistry(); err == nil {
		opts = append(opts, shelly.WithPluginRegistry(registry))
	}
	return shelly.New(shelly.NewConfigResolver(), opts...)
}

// KVSService returns the KVS service, lazily initialized.
// The service is backed by the ShellyService's connection handling.
// Cache and IOStreams are injected for automatic invalidation on mutations.
func (f *Factory) KVSService() *kvs.Service {
	if f.kvsService == nil {
		svc := f.ShellyService()
		f.kvsService = kvs.NewService(
			svc.WithConnection,
			kvs.WithFileCache(f.FileCache()),
			kvs.WithIOStreams(f.IOStreams()),
		)
	}
	return f.kvsService
}

// AutomationService returns the automation service, lazily initialized.
// Provides script, schedule, and event streaming functionality.
// Cache and IOStreams are injected for automatic invalidation on mutations.
func (f *Factory) AutomationService() *automation.Service {
	if f.automationService == nil {
		f.automationService = automation.New(f.ShellyService(), f.FileCache(), f.IOStreams())
	}
	return f.automationService
}

// ModbusService returns the Modbus service, lazily initialized.
// Provides Modbus-TCP configuration management.
func (f *Factory) ModbusService() *modbus.Service {
	if f.modbusService == nil {
		f.modbusService = modbus.New(f.ShellyService())
	}
	return f.modbusService
}

// SensorAddonService returns the Sensor Add-on service, lazily initialized.
// Provides sensor add-on peripheral management.
func (f *Factory) SensorAddonService() *sensoraddon.Service {
	if f.sensorAddonService == nil {
		f.sensorAddonService = sensoraddon.New(f.ShellyService())
	}
	return f.sensorAddonService
}

// FileCache returns the file-based cache, lazily initialized.
// Provides caching for device data shared between TUI and CLI.
// Returns nil if cache initialization fails (non-fatal).
// On first access, runs periodic cleanup if > 24h since last cleanup.
func (f *Factory) FileCache() *cache.FileCache {
	if f.fileCache == nil {
		fc, err := cache.New()
		if err != nil {
			f.IOStreams().DebugErr("initialize file cache", err)
			return nil
		}
		f.fileCache = fc

		// Run periodic cleanup (max once per 24 hours)
		removed, cleanupErr := fc.CleanupIfNeeded(24 * time.Hour)
		if cleanupErr != nil {
			f.IOStreams().DebugErr("cache cleanup", cleanupErr)
		} else if removed > 0 {
			f.IOStreams().Debug("cache cleanup removed %d expired entries", removed)
		}
	}
	return f.fileCache
}

// SetFileCache sets a custom file cache on an existing factory.
// This modifies the factory in-place and returns it for chaining.
// Use this for testing with isolated cache.
func (f *Factory) SetFileCache(fc *cache.FileCache) *Factory {
	f.fileCache = fc
	return f
}

// =============================================================================
// Context Helpers
// =============================================================================

// WithTimeout creates a child context with the specified timeout.
// This is a convenience wrapper around context.WithTimeout.
func (f *Factory) WithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

// WithDefaultTimeout creates a child context with flags.DefaultTimeout (10s).
// Useful for standard device operations.
func (f *Factory) WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, flags.DefaultTimeout)
}

// =============================================================================
// Config Accessors
// =============================================================================

// GetDevice retrieves a device from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetDevice(name string) *model.Device {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if dev, ok := cfg.Devices[name]; ok {
		return &dev
	}
	return nil
}

// GetGroup retrieves a device group from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetGroup(name string) *config.Group {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if grp, ok := cfg.Groups[name]; ok {
		return &grp
	}
	return nil
}

// GetAlias retrieves an alias from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetAlias(name string) *config.Alias {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if alias, ok := cfg.Aliases[name]; ok {
		return &alias
	}
	return nil
}

// GetScene retrieves a scene from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetScene(name string) *config.Scene {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if scene, ok := cfg.Scenes[name]; ok {
		return &scene
	}
	return nil
}

// GetDeviceTemplate retrieves a device template from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetDeviceTemplate(name string) *config.DeviceTemplate {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if tmpl, ok := cfg.Templates.Device[name]; ok {
		return &tmpl
	}
	return nil
}

// =============================================================================
// Device Resolution
// =============================================================================

// ResolveAddress resolves a device identifier to its address.
// Checks config registry first, falls back to using identifier as address.
func (f *Factory) ResolveAddress(identifier string) string {
	if dev := f.GetDevice(identifier); dev != nil && dev.Address != "" {
		return dev.Address
	}
	return identifier
}

// ResolveDevice resolves identifier to full device info.
// Returns (device, true) if found in config, (nil, false) otherwise.
func (f *Factory) ResolveDevice(identifier string) (*model.Device, bool) {
	cfg, err := f.Config()
	if err != nil {
		return nil, false
	}
	if dev, ok := cfg.Devices[identifier]; ok {
		return &dev, true
	}
	return nil, false
}

// =============================================================================
// Target Expansion
// =============================================================================

// ExpandTargets expands a list of device identifiers.
// Handles --group flag, --all flag, and individual device args.
// Returns resolved addresses ready for operations.
func (f *Factory) ExpandTargets(args []string, groupName string, all bool) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	var targets []string

	switch {
	case all:
		// Get all registered devices
		for _, dev := range cfg.Devices {
			if dev.Address != "" {
				targets = append(targets, dev.Address)
			}
		}
	case groupName != "":
		// Get devices from group
		group, ok := cfg.Groups[groupName]
		if !ok {
			return nil, fmt.Errorf("group %q not found", groupName)
		}
		for _, name := range group.Devices {
			targets = append(targets, f.ResolveAddress(name))
		}
	default:
		// Use provided args
		for _, arg := range args {
			targets = append(targets, f.ResolveAddress(arg))
		}
	}

	if len(targets) == 0 {
		return nil, errors.New("no target devices specified")
	}

	return targets, nil
}

// =============================================================================
// Confirmation Helpers
// =============================================================================

// ConfirmAction prompts for confirmation, respecting the yes flag.
// Returns true if confirmed (or yes is true), false otherwise.
func (f *Factory) ConfirmAction(message string, yes bool) (bool, error) {
	if yes {
		return true, nil
	}
	return f.IOStreams().Confirm(message, false)
}

// =============================================================================
// Output Format Helpers
// =============================================================================

// OutputFormat returns the current output format (json, yaml, table, or "").
func (f *Factory) OutputFormat() string {
	return viper.GetString("output")
}

// IsJSONOutput returns true if JSON output is requested.
func (f *Factory) IsJSONOutput() bool {
	return f.OutputFormat() == "json"
}

// IsYAMLOutput returns true if YAML output is requested.
func (f *Factory) IsYAMLOutput() bool {
	return f.OutputFormat() == "yaml"
}

// IsStructuredOutput returns true if JSON or YAML output is requested.
func (f *Factory) IsStructuredOutput() bool {
	format := f.OutputFormat()
	return format == "json" || format == "yaml"
}

// =============================================================================
// Logger Convenience
// =============================================================================

// Logger returns the structured logger from IOStreams.
func (f *Factory) Logger() *iostreams.Logger {
	return f.IOStreams().Logger()
}
