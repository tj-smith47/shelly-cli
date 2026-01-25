// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// File extension constants for config file parsing.
const (
	ExtJSON = ".json"
	ExtYAML = ".yaml"
	ExtYML  = ".yml"
)

// ThemeConfig supports both string and block theme configuration formats.
// It allows users to specify just a theme name, or customize with color overrides.
type ThemeConfig struct {
	Name     string                   `mapstructure:"name" json:"name,omitempty" yaml:"name,omitempty"`
	Colors   map[string]string        `mapstructure:"colors" json:"colors,omitempty" yaml:"colors,omitempty"`
	Semantic *theme.SemanticOverrides `mapstructure:"semantic" json:"semantic,omitempty" yaml:"semantic,omitempty"`
	File     string                   `mapstructure:"file" json:"file,omitempty" yaml:"file,omitempty"`
}

// Config holds all CLI configuration.
type Config struct {
	// Global settings
	Output    string `mapstructure:"output" yaml:"output,omitempty"`
	Color     bool   `mapstructure:"color" yaml:"color,omitempty"`
	Theme     any    `mapstructure:"theme" yaml:"theme,omitempty"` // Can be string or ThemeConfig
	APIMode   string `mapstructure:"api_mode" yaml:"api_mode,omitempty"`
	Verbose   bool   `mapstructure:"verbose" yaml:"verbose,omitempty"`
	Quiet     bool   `mapstructure:"quiet" yaml:"quiet,omitempty"`
	Editor    string `mapstructure:"editor" yaml:"editor,omitempty"`       // Preferred editor command (falls back to $EDITOR, $VISUAL, then nano)
	Telemetry bool   `mapstructure:"telemetry" yaml:"telemetry,omitempty"` // Opt-in anonymous usage telemetry

	// Discovery settings
	Discovery DiscoveryConfig `mapstructure:"discovery" yaml:"discovery,omitempty"`

	// Cloud settings
	Cloud CloudConfig `mapstructure:"cloud" yaml:"cloud,omitempty"`

	// Integrator settings
	Integrator IntegratorConfig `mapstructure:"integrator" yaml:"integrator,omitempty"`

	// Device registry
	Devices map[string]model.Device `mapstructure:"devices" yaml:"devices,omitempty"`

	// Aliases
	Aliases map[string]Alias `mapstructure:"aliases" yaml:"aliases,omitempty"`

	// Groups
	Groups map[string]Group `mapstructure:"groups" yaml:"groups,omitempty"`

	// Scenes
	Scenes map[string]Scene `mapstructure:"scenes" yaml:"scenes,omitempty"`

	// Templates (device and script)
	Templates TemplatesConfig `mapstructure:"templates" yaml:"templates,omitempty"`

	// Alerts
	Alerts map[string]Alert `mapstructure:"alerts" yaml:"alerts,omitempty"`

	// Plugin settings
	Plugins PluginsConfig `mapstructure:"plugins" yaml:"plugins,omitempty"`

	// Rate limiting settings
	RateLimit RateLimitConfig `mapstructure:"ratelimit" yaml:"ratelimit,omitempty"`

	// TUI settings
	TUI TUIConfig `mapstructure:"tui" yaml:"tui,omitempty"`

	// Energy monitoring settings
	Energy EnergyConfig `mapstructure:"energy" yaml:"energy,omitempty"`
}

// TUIConfig holds TUI dashboard settings.
type TUIConfig struct {
	Refresh     TUIRefreshConfig  `mapstructure:"refresh" yaml:"refresh,omitempty"`
	Keybindings KeybindingsConfig `mapstructure:"keybindings" yaml:"keybindings,omitempty"`
	Theme       *ThemeConfig      `mapstructure:"theme" yaml:"theme,omitempty"`
	Events      TUIEventsConfig   `mapstructure:"events" yaml:"events,omitempty"`
}

// TUIEventsConfig holds event filtering settings for the TUI events panel.
type TUIEventsConfig struct {
	// FilteredEvents is a list of event names to filter out from the events panel.
	// Common noisy events: "ble.scan_result", "sys" status changes, "wifi" updates.
	// Use glob patterns like "ble.*" to filter all BLE events.
	FilteredEvents []string `mapstructure:"filtered_events" yaml:"filtered_events,omitempty"`

	// FilteredComponents is a list of component types to filter out.
	// Examples: "sys", "wifi", "cloud", "ts" (timestamp-only changes).
	FilteredComponents []string `mapstructure:"filtered_components" yaml:"filtered_components,omitempty"`

	// MaxItems is the maximum number of events to keep in each list (user/system).
	// Defaults to 100.
	MaxItems int `mapstructure:"max_items" yaml:"max_items,omitempty"`
}

// DefaultTUIEventsConfig returns sensible defaults for TUI event filtering.
func DefaultTUIEventsConfig() TUIEventsConfig {
	return TUIEventsConfig{
		// Default filtered events - high-frequency noise that clutters the UI
		FilteredEvents: []string{
			"ble.scan_result", // BLE scanning produces many events
		},
		// Default filtered components - status changes that aren't user-actionable
		FilteredComponents: []string{
			"sys",   // System heartbeats
			"wifi",  // WiFi signal fluctuations
			"cloud", // Cloud connection status
			"ts",    // Timestamp-only updates
		},
		MaxItems: 100,
	}
}

// EnergyConfig holds energy monitoring and cost calculation settings.
type EnergyConfig struct {
	CostRate float64 `mapstructure:"cost_rate" yaml:"cost_rate,omitempty"` // Cost per kWh in local currency
	Currency string  `mapstructure:"currency" yaml:"currency,omitempty"`   // Currency symbol (e.g., "$", "€", "£")
}

// DefaultEnergyConfig returns sensible defaults for energy configuration.
func DefaultEnergyConfig() EnergyConfig {
	return EnergyConfig{
		CostRate: 0.12, // Default $0.12/kWh (US average)
		Currency: "$",
	}
}

// KeybindingsConfig holds customizable keybindings for the TUI.
type KeybindingsConfig struct {
	// Navigation
	Up       []string `mapstructure:"up" yaml:"up,omitempty"`
	Down     []string `mapstructure:"down" yaml:"down,omitempty"`
	Left     []string `mapstructure:"left" yaml:"left,omitempty"`
	Right    []string `mapstructure:"right" yaml:"right,omitempty"`
	PageUp   []string `mapstructure:"page_up" yaml:"page_up,omitempty"`
	PageDown []string `mapstructure:"page_down" yaml:"page_down,omitempty"`
	Home     []string `mapstructure:"home" yaml:"home,omitempty"`
	End      []string `mapstructure:"end" yaml:"end,omitempty"`

	// Actions
	Enter      []string `mapstructure:"enter" yaml:"enter,omitempty"`
	Escape     []string `mapstructure:"escape" yaml:"escape,omitempty"`
	Refresh    []string `mapstructure:"refresh" yaml:"refresh,omitempty"`
	RefreshAll []string `mapstructure:"refresh_all" yaml:"refresh_all,omitempty"`
	Filter     []string `mapstructure:"filter" yaml:"filter,omitempty"`
	Command    []string `mapstructure:"command" yaml:"command,omitempty"`
	Help       []string `mapstructure:"help" yaml:"help,omitempty"`
	Quit       []string `mapstructure:"quit" yaml:"quit,omitempty"`

	// Device actions
	Toggle  []string `mapstructure:"toggle" yaml:"toggle,omitempty"`
	TurnOn  []string `mapstructure:"turn_on" yaml:"turn_on,omitempty"`
	TurnOff []string `mapstructure:"turn_off" yaml:"turn_off,omitempty"`
	Reboot  []string `mapstructure:"reboot" yaml:"reboot,omitempty"`

	// View switching
	Tab      []string `mapstructure:"tab" yaml:"tab,omitempty"`
	ShiftTab []string `mapstructure:"shift_tab" yaml:"shift_tab,omitempty"`
	View1    []string `mapstructure:"view1" yaml:"view1,omitempty"`
	View2    []string `mapstructure:"view2" yaml:"view2,omitempty"`
	View3    []string `mapstructure:"view3" yaml:"view3,omitempty"`
	View4    []string `mapstructure:"view4" yaml:"view4,omitempty"`
	View5    []string `mapstructure:"view5" yaml:"view5,omitempty"`
	View6    []string `mapstructure:"view6" yaml:"view6,omitempty"`
}

// DiscoveryConfig holds device discovery settings.
type DiscoveryConfig struct {
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout,omitempty"`
	MDNS    bool          `mapstructure:"mdns" yaml:"mdns,omitempty"`
	BLE     bool          `mapstructure:"ble" yaml:"ble,omitempty"`
	CoIoT   bool          `mapstructure:"coiot" yaml:"coiot,omitempty"`
	Network string        `mapstructure:"network" yaml:"network,omitempty"` // Default subnet for scanning
}

// CloudConfig holds Shelly Cloud API settings.
type CloudConfig struct {
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled,omitempty"`
	Email        string `mapstructure:"email" yaml:"email,omitempty"`
	AccessToken  string `mapstructure:"access_token" yaml:"access_token,omitempty"`
	RefreshToken string `mapstructure:"refresh_token" yaml:"refresh_token,omitempty"`
	AuthKey      string `mapstructure:"auth_key" yaml:"auth_key,omitempty"`
	ServerURL    string `mapstructure:"server_url" yaml:"server_url,omitempty"`
}

// IntegratorConfig holds Shelly Integrator API settings.
type IntegratorConfig struct {
	Tag   string `mapstructure:"tag" yaml:"tag,omitempty"`
	Token string `mapstructure:"token" yaml:"token,omitempty"`
}

// GetIntegratorCredentials returns integrator credentials from environment or config.
// Environment variables take precedence: SHELLY_INTEGRATOR_TAG, SHELLY_INTEGRATOR_TOKEN.
// Returns an error if credentials are not configured.
func (c *Config) GetIntegratorCredentials() (tag, token string, err error) {
	tag = os.Getenv("SHELLY_INTEGRATOR_TAG")
	token = os.Getenv("SHELLY_INTEGRATOR_TOKEN")

	if c != nil {
		if tag == "" {
			tag = c.Integrator.Tag
		}
		if token == "" {
			token = c.Integrator.Token
		}
	}

	if tag == "" || token == "" {
		return "", "", fmt.Errorf("integrator credentials required")
	}
	return tag, token, nil
}

// Alias represents a command alias.
type Alias struct {
	Command string `mapstructure:"command" yaml:"command,omitempty"`
	Shell   bool   `mapstructure:"shell" yaml:"shell,omitempty"` // If true, execute via shell
}

// Group represents a device group.
type Group struct {
	Devices []string `mapstructure:"devices" yaml:"devices,omitempty"`
}

// Scene represents a saved device state configuration.
type Scene struct {
	Name        string        `mapstructure:"name" json:"name" yaml:"name"`
	Description string        `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Actions     []SceneAction `mapstructure:"actions" json:"actions" yaml:"actions"`
}

// SceneAction represents a single action within a scene.
type SceneAction struct {
	Device string         `mapstructure:"device" json:"device" yaml:"device"`
	Method string         `mapstructure:"method" json:"method" yaml:"method"`
	Params map[string]any `mapstructure:"params,omitempty" json:"params,omitempty" yaml:"params,omitempty"`
}

// TemplatesConfig holds all template types.
type TemplatesConfig struct {
	Device map[string]DeviceTemplate `mapstructure:"device" yaml:"device,omitempty"`
	Script map[string]ScriptTemplate `mapstructure:"script" yaml:"script,omitempty"`
}

// DeviceTemplate represents a device configuration template.
type DeviceTemplate struct {
	Name         string         `mapstructure:"name" json:"name" yaml:"name"`
	Description  string         `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Model        string         `mapstructure:"model" json:"model" yaml:"model"`
	App          string         `mapstructure:"app,omitempty" json:"app,omitempty" yaml:"app,omitempty"`
	Generation   int            `mapstructure:"generation" json:"generation" yaml:"generation"`
	Config       map[string]any `mapstructure:"config" json:"config" yaml:"config"`
	CreatedAt    string         `mapstructure:"created_at" json:"created_at" yaml:"created_at"`
	SourceDevice string         `mapstructure:"source_device,omitempty" json:"source_device,omitempty" yaml:"source_device,omitempty"`
}

// ScriptTemplate represents a reusable JavaScript script template.
type ScriptTemplate struct {
	Name        string           `mapstructure:"name" json:"name" yaml:"name"`
	Description string           `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Code        string           `mapstructure:"code" json:"code" yaml:"code"`
	Category    string           `mapstructure:"category,omitempty" json:"category,omitempty" yaml:"category,omitempty"`    // e.g., "automation", "monitoring", "utility"
	Variables   []ScriptVariable `mapstructure:"variables,omitempty" json:"variables,omitempty" yaml:"variables,omitempty"` // Configurable variables
	Models      []string         `mapstructure:"models,omitempty" json:"models,omitempty" yaml:"models,omitempty"`          // Compatible device models (empty = all)
	MinGen      int              `mapstructure:"min_gen,omitempty" json:"min_gen,omitempty" yaml:"min_gen,omitempty"`       // Minimum generation (2 for Gen2+)
	BuiltIn     bool             `mapstructure:"-" json:"-" yaml:"-"`                                                       // True for bundled templates
	Author      string           `mapstructure:"author,omitempty" json:"author,omitempty" yaml:"author,omitempty"`
	Version     string           `mapstructure:"version,omitempty" json:"version,omitempty" yaml:"version,omitempty"`
}

// ScriptVariable represents a configurable variable in a script template.
type ScriptVariable struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name"` // Variable name in code (e.g., "THRESHOLD")
	Description string `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Type        string `mapstructure:"type,omitempty" json:"type,omitempty" yaml:"type,omitempty"` // "string", "number", "boolean"
	Default     any    `mapstructure:"default,omitempty" json:"default,omitempty" yaml:"default,omitempty"`
	Required    bool   `mapstructure:"required,omitempty" json:"required,omitempty" yaml:"required,omitempty"`
}

// PluginsConfig holds plugin system settings.
type PluginsConfig struct {
	Enabled bool     `mapstructure:"enabled" yaml:"enabled,omitempty"`
	Path    []string `mapstructure:"path" yaml:"path,omitempty"` // Additional plugin search paths
}

// RateLimitConfig holds rate limiting settings to prevent overloading Shelly devices.
// Shelly devices have hardware limitations:
//   - Gen1 (ESP8266): MAX 2 concurrent HTTP connections
//   - Gen2 (ESP32): MAX 5 concurrent HTTP transactions
type RateLimitConfig struct {
	Gen1   GenerationRateLimitConfig `mapstructure:"gen1" yaml:"gen1,omitempty"`
	Gen2   GenerationRateLimitConfig `mapstructure:"gen2" yaml:"gen2,omitempty"`
	Global GlobalRateLimitConfig     `mapstructure:"global" yaml:"global,omitempty"`
}

// GenerationRateLimitConfig holds rate limiting settings for a specific device generation.
type GenerationRateLimitConfig struct {
	MinInterval      time.Duration `mapstructure:"min_interval" yaml:"min_interval,omitempty"`           // Min time between requests to same device
	MaxConcurrent    int           `mapstructure:"max_concurrent" yaml:"max_concurrent,omitempty"`       // Max in-flight requests per device
	CircuitThreshold int           `mapstructure:"circuit_threshold" yaml:"circuit_threshold,omitempty"` // Failures before circuit opens
}

// GlobalRateLimitConfig holds global rate limiting settings.
type GlobalRateLimitConfig struct {
	MaxConcurrent           int           `mapstructure:"max_concurrent" yaml:"max_concurrent,omitempty"`                       // Total concurrent requests across all devices
	CircuitOpenDuration     time.Duration `mapstructure:"circuit_open_duration" yaml:"circuit_open_duration,omitempty"`         // How long circuit stays open
	CircuitSuccessThreshold int           `mapstructure:"circuit_success_threshold" yaml:"circuit_success_threshold,omitempty"` // Successes to close circuit
}

// DefaultRateLimitConfig returns sensible defaults based on Shelly hardware constraints.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Gen1: GenerationRateLimitConfig{
			MinInterval:      2 * time.Second, // Gen1 needs breathing room
			MaxConcurrent:    1,               // Leave 1 connection for safety
			CircuitThreshold: 3,               // Open circuit after 3 failures
		},
		Gen2: GenerationRateLimitConfig{
			MinInterval:      500 * time.Millisecond, // Gen2 handles faster polling
			MaxConcurrent:    3,                      // Leave 2 connections for safety
			CircuitThreshold: 5,                      // Gen2 is more resilient
		},
		Global: GlobalRateLimitConfig{
			MaxConcurrent:           20,               // Total across all devices
			CircuitOpenDuration:     60 * time.Second, // Standard backoff
			CircuitSuccessThreshold: 2,                // Successes to close circuit
		},
	}
}

// TUIRefreshConfig holds adaptive refresh interval settings for the TUI.
type TUIRefreshConfig struct {
	Gen1Online   time.Duration `mapstructure:"gen1_online" yaml:"gen1_online,omitempty"`     // Refresh for online Gen1 devices
	Gen1Offline  time.Duration `mapstructure:"gen1_offline" yaml:"gen1_offline,omitempty"`   // Refresh for offline Gen1 devices
	Gen2Online   time.Duration `mapstructure:"gen2_online" yaml:"gen2_online,omitempty"`     // Refresh for online Gen2 devices
	Gen2Offline  time.Duration `mapstructure:"gen2_offline" yaml:"gen2_offline,omitempty"`   // Refresh for offline Gen2 devices
	FocusedBoost time.Duration `mapstructure:"focused_boost" yaml:"focused_boost,omitempty"` // Refresh for focused device
}

// DefaultTUIRefreshConfig returns sensible defaults for TUI refresh intervals.
func DefaultTUIRefreshConfig() TUIRefreshConfig {
	return TUIRefreshConfig{
		Gen1Online:   15 * time.Second, // Gen1 responds well but is fragile
		Gen1Offline:  60 * time.Second, // Don't hammer offline Gen1
		Gen2Online:   5 * time.Second,  // Gen2 handles faster polling
		Gen2Offline:  30 * time.Second, // Back off for offline Gen2
		FocusedBoost: 3 * time.Second,  // Focused device gets priority
	}
}

// GetRateLimitConfig returns the rate limit config with defaults applied.
// Zero values in the config are replaced with sensible defaults.
func (c *Config) GetRateLimitConfig() RateLimitConfig {
	defaults := DefaultRateLimitConfig()
	cfg := c.RateLimit

	// Apply defaults for zero values
	if cfg.Gen1.MinInterval == 0 {
		cfg.Gen1.MinInterval = defaults.Gen1.MinInterval
	}
	if cfg.Gen1.MaxConcurrent == 0 {
		cfg.Gen1.MaxConcurrent = defaults.Gen1.MaxConcurrent
	}
	if cfg.Gen1.CircuitThreshold == 0 {
		cfg.Gen1.CircuitThreshold = defaults.Gen1.CircuitThreshold
	}

	if cfg.Gen2.MinInterval == 0 {
		cfg.Gen2.MinInterval = defaults.Gen2.MinInterval
	}
	if cfg.Gen2.MaxConcurrent == 0 {
		cfg.Gen2.MaxConcurrent = defaults.Gen2.MaxConcurrent
	}
	if cfg.Gen2.CircuitThreshold == 0 {
		cfg.Gen2.CircuitThreshold = defaults.Gen2.CircuitThreshold
	}

	if cfg.Global.MaxConcurrent == 0 {
		cfg.Global.MaxConcurrent = defaults.Global.MaxConcurrent
	}
	if cfg.Global.CircuitOpenDuration == 0 {
		cfg.Global.CircuitOpenDuration = defaults.Global.CircuitOpenDuration
	}
	if cfg.Global.CircuitSuccessThreshold == 0 {
		cfg.Global.CircuitSuccessThreshold = defaults.Global.CircuitSuccessThreshold
	}

	return cfg
}

// GetTUIRefreshConfig returns the TUI refresh config with defaults applied.
// Zero values in the config are replaced with sensible defaults.
func (c *Config) GetTUIRefreshConfig() TUIRefreshConfig {
	defaults := DefaultTUIRefreshConfig()
	cfg := c.TUI.Refresh

	if cfg.Gen1Online == 0 {
		cfg.Gen1Online = defaults.Gen1Online
	}
	if cfg.Gen1Offline == 0 {
		cfg.Gen1Offline = defaults.Gen1Offline
	}
	if cfg.Gen2Online == 0 {
		cfg.Gen2Online = defaults.Gen2Online
	}
	if cfg.Gen2Offline == 0 {
		cfg.Gen2Offline = defaults.Gen2Offline
	}
	if cfg.FocusedBoost == 0 {
		cfg.FocusedBoost = defaults.FocusedBoost
	}

	return cfg
}

// GetTUIEventsConfig returns the TUI events config with defaults applied.
// Empty lists use defaults; user-provided lists replace defaults entirely.
func (c *Config) GetTUIEventsConfig() TUIEventsConfig {
	defaults := DefaultTUIEventsConfig()
	cfg := c.TUI.Events

	// If no filtered events specified, use defaults
	if len(cfg.FilteredEvents) == 0 {
		cfg.FilteredEvents = defaults.FilteredEvents
	}

	// If no filtered components specified, use defaults
	if len(cfg.FilteredComponents) == 0 {
		cfg.FilteredComponents = defaults.FilteredComponents
	}

	// Use default max items if not specified
	if cfg.MaxItems == 0 {
		cfg.MaxItems = defaults.MaxItems
	}

	return cfg
}

// GetEnergyConfig returns the energy config with defaults applied.
// Zero values in the config are replaced with sensible defaults.
func (c *Config) GetEnergyConfig() EnergyConfig {
	defaults := DefaultEnergyConfig()
	cfg := c.Energy

	if cfg.CostRate == 0 {
		cfg.CostRate = defaults.CostRate
	}
	if cfg.Currency == "" {
		cfg.Currency = defaults.Currency
	}

	return cfg
}

// GetGlobalMaxConcurrent returns the global max concurrent setting.
// This is a convenience function for rate limiting concurrency across the codebase.
// It uses the configured value if available, otherwise returns the default.
func GetGlobalMaxConcurrent() int {
	cfg := Get()
	if cfg != nil {
		rlCfg := cfg.GetRateLimitConfig()
		if rlCfg.Global.MaxConcurrent > 0 {
			return rlCfg.Global.MaxConcurrent
		}
	}
	return DefaultRateLimitConfig().Global.MaxConcurrent
}

// Validate validates the rate limit configuration.
func (c *RateLimitConfig) Validate() error {
	if c.Gen1.MaxConcurrent < 0 || c.Gen1.MaxConcurrent > 2 {
		return fmt.Errorf("gen1.max_concurrent must be 0-2 (Gen1 devices only support 2 connections), got %d", c.Gen1.MaxConcurrent)
	}
	if c.Gen2.MaxConcurrent < 0 || c.Gen2.MaxConcurrent > 5 {
		return fmt.Errorf("gen2.max_concurrent must be 0-5 (Gen2 devices only support 5 connections), got %d", c.Gen2.MaxConcurrent)
	}
	if c.Gen1.MinInterval < 0 {
		return fmt.Errorf("gen1.min_interval must be non-negative, got %v", c.Gen1.MinInterval)
	}
	if c.Gen2.MinInterval < 0 {
		return fmt.Errorf("gen2.min_interval must be non-negative, got %v", c.Gen2.MinInterval)
	}
	if c.Global.MaxConcurrent < 1 {
		return fmt.Errorf("global.max_concurrent must be at least 1, got %d", c.Global.MaxConcurrent)
	}
	return nil
}

// IsZero returns true if all fields are zero values (no config specified).
func (c *RateLimitConfig) IsZero() bool {
	return c.Gen1.MinInterval == 0 && c.Gen1.MaxConcurrent == 0 && c.Gen1.CircuitThreshold == 0 &&
		c.Gen2.MinInterval == 0 && c.Gen2.MaxConcurrent == 0 && c.Gen2.CircuitThreshold == 0 &&
		c.Global.MaxConcurrent == 0 && c.Global.CircuitOpenDuration == 0 && c.Global.CircuitSuccessThreshold == 0
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
	defaultManagerMu   sync.Mutex
)

// ResetDefaultManagerForTesting resets the default manager singleton.
// This MUST only be called from tests to ensure proper isolation.
// After calling this, the next call to getDefaultManager will re-initialize
// with the current HOME directory.
func ResetDefaultManagerForTesting() {
	defaultManagerMu.Lock()
	defer defaultManagerMu.Unlock()
	defaultManager = nil
	defaultManagerOnce = sync.Once{}
}

// SetDefaultManager sets a custom manager as the default.
// This is used for demo mode to inject mock configuration.
func SetDefaultManager(mgr *Manager) {
	defaultManagerMu.Lock()
	defer defaultManagerMu.Unlock()
	defaultManager = mgr
	defaultManagerOnce.Do(func() {})
}

// getDefaultManager returns the package-level default manager.
// Used by package-level convenience functions (Load, Get, etc.).
func getDefaultManager() *Manager {
	defaultManagerMu.Lock()
	defer defaultManagerMu.Unlock()
	defaultManagerOnce.Do(func() {
		defaultManager = NewManager("")
		if err := defaultManager.Load(); err != nil {
			// Load failure is not fatal - Get() will return nil and we handle that
			// This can happen if config file is missing or unreadable
			defaultManager.config = &Config{
				Devices: make(map[string]model.Device),
				Aliases: make(map[string]Alias),
				Groups:  make(map[string]Group),
				Scenes:  make(map[string]Scene),
				Templates: TemplatesConfig{
					Device: make(map[string]DeviceTemplate),
					Script: make(map[string]ScriptTemplate),
				},
				Alerts: make(map[string]Alert),
			}
			defaultManager.loaded = true
		}
	})
	return defaultManager
}

// Get returns the current configuration.
// For mutations, use a Manager instance directly.
func Get() *Config {
	mgr := getDefaultManager()
	cfg := mgr.Get()
	if cfg == nil {
		return &Config{
			Output:  "table",
			Color:   true,
			Theme:   "dracula",
			APIMode: "local",
			Devices: make(map[string]model.Device),
			Aliases: make(map[string]Alias),
			Groups:  make(map[string]Group),
			Scenes:  make(map[string]Scene),
			Templates: TemplatesConfig{
				Device: make(map[string]DeviceTemplate),
				Script: make(map[string]ScriptTemplate),
			},
			Alerts: make(map[string]Alert),
		}
	}
	return cfg
}

// GetThemeConfig normalizes the Theme field to ThemeConfig.
// It handles both string format (e.g., "dracula") and block format.
func (c *Config) GetThemeConfig() ThemeConfig {
	if c == nil {
		return ThemeConfig{Name: "dracula"}
	}

	switch v := c.Theme.(type) {
	case string:
		return ThemeConfig{Name: v}
	case map[string]any:
		var tc ThemeConfig
		if err := mapstructure.Decode(v, &tc); err != nil {
			return ThemeConfig{Name: "dracula"}
		}
		return tc
	case ThemeConfig:
		return v
	case *ThemeConfig:
		if v != nil {
			return *v
		}
		return ThemeConfig{Name: "dracula"}
	default:
		return ThemeConfig{Name: "dracula"}
	}
}

// GetTUIThemeConfig returns the TUI-specific theme config, or nil if not set.
// When TUI theme is set, it completely replaces the main theme (independent).
func (c *Config) GetTUIThemeConfig() *ThemeConfig {
	if c == nil || c.TUI.Theme == nil {
		return nil
	}
	return c.TUI.Theme
}

// GetEditor returns the configured editor command.
// Returns empty string if not configured (caller should fall back to env vars).
func (c *Config) GetEditor() string {
	if c == nil {
		return ""
	}
	return c.Editor
}

// DefaultAliases are built-in aliases provided out of the box.
// These provide shortcuts for common operations.
var DefaultAliases = map[string]Alias{
	// Device management shortcuts
	"add":    {Command: "device add $@"},
	"remove": {Command: "device remove $@"},
	"rename": {Command: "device rename $@"},
	"ping":   {Command: "device ping $@"},
	"reboot": {Command: "device reboot $@"},
	"reset":  {Command: "device factory-reset $@"},
	"info":   {Command: "device info $@"},
	// Config shortcuts
	"diff":        {Command: "device config diff $@"},
	"export-json": {Command: "device config export $1 $2 --format json"},
	"export-yaml": {Command: "device config export $1 $2 --format yaml"},
	// Schedule shortcuts
	"cron": {Command: "schedule create --timespec \"$1\" $2"},
}

// Load reads configuration from file and environment using the default manager.
func Load() (*Config, error) {
	mgr := getDefaultManager()
	// Manager.Load() is safe to call multiple times
	if err := mgr.Load(); err != nil {
		return nil, err
	}
	return mgr.Get(), nil
}

// Reload forces a reload of the configuration.
func Reload() (*Config, error) {
	mgr := getDefaultManager()
	if err := mgr.Reload(); err != nil {
		return nil, err
	}
	return mgr.Get(), nil
}

// Save writes the current configuration to file (method version).
func (c *Config) Save() error {
	return Save()
}

// Save writes the current configuration to file.
// In test mode (when using an in-memory filesystem), this is a no-op to prevent
// accidental writes to the real config file.
func Save() error {
	if IsTestFs() {
		return nil // Skip save in test mode
	}
	return getDefaultManager().Save()
}

// Dir returns the configuration directory path.
func Dir() (string, error) {
	// Respect XDG_CONFIG_HOME if set (standard for Linux config directories)
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "shelly"), nil
	}
	// Fall back to ~/.config/shelly
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "shelly"), nil
}

// CacheDir returns the cache directory path.
// This uses the OS-standard cache directory (~/.cache/shelly on Linux).
func CacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "shelly"), nil
}

// PluginsDir returns the plugins directory path.
func PluginsDir() (string, error) {
	configDir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}

// GetAllDeviceCredentials returns credentials for all devices that have auth configured.
func (c *Config) GetAllDeviceCredentials() map[string]struct{ Username, Password string } {
	return getDefaultManager().GetAllDeviceCredentials()
}

// SetDeviceAuth sets authentication credentials for a device.
func (c *Config) SetDeviceAuth(deviceName, username, password string) error {
	return getDefaultManager().SetDeviceAuth(deviceName, username, password)
}
