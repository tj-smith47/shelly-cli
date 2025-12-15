// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ThemeConfig supports both string and block theme configuration formats.
// It allows users to specify just a theme name, or customize with color overrides.
type ThemeConfig struct {
	Name   string            `mapstructure:"name" json:"name,omitempty" yaml:"name,omitempty"`
	Colors map[string]string `mapstructure:"colors" json:"colors,omitempty" yaml:"colors,omitempty"`
	File   string            `mapstructure:"file" json:"file,omitempty" yaml:"file,omitempty"`
}

// Config holds all CLI configuration.
type Config struct {
	// Global settings
	Output  string `mapstructure:"output"`
	Color   bool   `mapstructure:"color"`
	Theme   any    `mapstructure:"theme"` // Can be string or ThemeConfig
	APIMode string `mapstructure:"api_mode"`
	Verbose bool   `mapstructure:"verbose"`
	Quiet   bool   `mapstructure:"quiet"`

	// Discovery settings
	Discovery DiscoveryConfig `mapstructure:"discovery"`

	// Cloud settings
	Cloud CloudConfig `mapstructure:"cloud"`

	// Device registry
	Devices map[string]model.Device `mapstructure:"devices"`

	// Aliases
	Aliases map[string]Alias `mapstructure:"aliases"`

	// Groups
	Groups map[string]Group `mapstructure:"groups"`

	// Scenes
	Scenes map[string]Scene `mapstructure:"scenes"`

	// Templates
	Templates map[string]Template `mapstructure:"templates"`

	// Plugin settings
	Plugins PluginsConfig `mapstructure:"plugins"`

	// TUI settings
	TUI TUIConfig `mapstructure:"tui"`
}

// TUIConfig holds TUI dashboard settings.
type TUIConfig struct {
	RefreshInterval int               `mapstructure:"refresh_interval"` // Refresh interval in seconds
	Keybindings     KeybindingsConfig `mapstructure:"keybindings"`
	Theme           *ThemeConfig      `mapstructure:"theme"` // Independent TUI theme (replaces main theme when set)
}

// KeybindingsConfig holds customizable keybindings for the TUI.
type KeybindingsConfig struct {
	// Navigation
	Up       []string `mapstructure:"up"`
	Down     []string `mapstructure:"down"`
	Left     []string `mapstructure:"left"`
	Right    []string `mapstructure:"right"`
	PageUp   []string `mapstructure:"page_up"`
	PageDown []string `mapstructure:"page_down"`
	Home     []string `mapstructure:"home"`
	End      []string `mapstructure:"end"`

	// Actions
	Enter   []string `mapstructure:"enter"`
	Escape  []string `mapstructure:"escape"`
	Refresh []string `mapstructure:"refresh"`
	Filter  []string `mapstructure:"filter"`
	Command []string `mapstructure:"command"`
	Help    []string `mapstructure:"help"`
	Quit    []string `mapstructure:"quit"`

	// Device actions
	Toggle  []string `mapstructure:"toggle"`
	TurnOn  []string `mapstructure:"turn_on"`
	TurnOff []string `mapstructure:"turn_off"`
	Reboot  []string `mapstructure:"reboot"`

	// View switching
	Tab      []string `mapstructure:"tab"`
	ShiftTab []string `mapstructure:"shift_tab"`
	View1    []string `mapstructure:"view1"`
	View2    []string `mapstructure:"view2"`
	View3    []string `mapstructure:"view3"`
	View4    []string `mapstructure:"view4"`
}

// DiscoveryConfig holds device discovery settings.
type DiscoveryConfig struct {
	Timeout time.Duration `mapstructure:"timeout"`
	MDNS    bool          `mapstructure:"mdns"`
	BLE     bool          `mapstructure:"ble"`
	CoIoT   bool          `mapstructure:"coiot"`
	Network string        `mapstructure:"network"` // Default subnet for scanning
}

// CloudConfig holds Shelly Cloud API settings.
type CloudConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Email        string `mapstructure:"email"`
	AccessToken  string `mapstructure:"access_token"`
	RefreshToken string `mapstructure:"refresh_token"`
	ServerURL    string `mapstructure:"server_url"`
}

// Alias represents a command alias.
type Alias struct {
	Name    string `mapstructure:"name"`
	Command string `mapstructure:"command"`
	Shell   bool   `mapstructure:"shell"` // If true, execute via shell
}

// Group represents a device group.
type Group struct {
	Name    string   `mapstructure:"name"`
	Devices []string `mapstructure:"devices"`
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

// Template represents a device configuration template.
type Template struct {
	Name         string         `mapstructure:"name" json:"name" yaml:"name"`
	Description  string         `mapstructure:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	Model        string         `mapstructure:"model" json:"model" yaml:"model"`
	App          string         `mapstructure:"app,omitempty" json:"app,omitempty" yaml:"app,omitempty"`
	Generation   int            `mapstructure:"generation" json:"generation" yaml:"generation"`
	Config       map[string]any `mapstructure:"config" json:"config" yaml:"config"`
	CreatedAt    string         `mapstructure:"created_at" json:"created_at" yaml:"created_at"`
	SourceDevice string         `mapstructure:"source_device,omitempty" json:"source_device,omitempty" yaml:"source_device,omitempty"`
}

// PluginsConfig holds plugin system settings.
type PluginsConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Path    []string `mapstructure:"path"` // Additional plugin search paths
}

var (
	cfg     *Config
	cfgOnce sync.Once
	cfgMu   sync.RWMutex
)

// Get returns the current configuration.
func Get() *Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	if cfg == nil {
		return &Config{
			Output:    "table",
			Color:     true,
			Theme:     "dracula",
			APIMode:   "local",
			Devices:   make(map[string]model.Device),
			Aliases:   make(map[string]Alias),
			Groups:    make(map[string]Group),
			Scenes:    make(map[string]Scene),
			Templates: make(map[string]Template),
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

// DefaultAliases are built-in aliases provided out of the box.
// These provide shortcuts for common operations.
var DefaultAliases = map[string]Alias{
	"api": {
		Name:    "api",
		Command: "debug rpc $@",
		Shell:   false,
	},
	"cron": {
		Name:    "cron",
		Command: "schedule create --timespec \"$1\" $2",
		Shell:   false,
	},
}

// Load reads configuration from file and environment.
func Load() (*Config, error) {
	var loadErr error

	cfgOnce.Do(func() {
		c := &Config{}
		if err := viper.Unmarshal(c); err != nil {
			loadErr = fmt.Errorf("failed to unmarshal config: %w", err)
			return
		}

		// Initialize maps if nil
		if c.Devices == nil {
			c.Devices = make(map[string]model.Device)
		}
		if c.Aliases == nil {
			c.Aliases = make(map[string]Alias)
		}
		if c.Groups == nil {
			c.Groups = make(map[string]Group)
		}
		if c.Scenes == nil {
			c.Scenes = make(map[string]Scene)
		}
		if c.Templates == nil {
			c.Templates = make(map[string]Template)
		}

		cfgMu.Lock()
		cfg = c
		cfgMu.Unlock()
	})

	if loadErr != nil {
		return nil, loadErr
	}

	return Get(), nil
}

// Reload forces a reload of the configuration.
func Reload() (*Config, error) {
	cfgOnce = sync.Once{} // Reset the once
	return Load()
}

// Save writes the current configuration to file (method version).
func (c *Config) Save() error {
	return Save()
}

// Save writes the current configuration to file.
func Save() error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	// Set all values in viper
	viper.Set("output", c.Output)
	viper.Set("color", c.Color)
	viper.Set("theme", c.Theme)
	viper.Set("api_mode", c.APIMode)
	viper.Set("verbose", c.Verbose)
	viper.Set("quiet", c.Quiet)
	viper.Set("discovery", c.Discovery)
	viper.Set("cloud", c.Cloud)
	viper.Set("devices", c.Devices)
	viper.Set("aliases", c.Aliases)
	viper.Set("groups", c.Groups)
	viper.Set("scenes", c.Scenes)
	viper.Set("plugins", c.Plugins)
	viper.Set("tui", c.TUI)

	// Get config file path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(home, ".config", "shelly", "config.yaml")
	}

	// Ensure directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Dir returns the configuration directory path.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "shelly"), nil
}

// PluginsDir returns the plugins directory path.
func PluginsDir() (string, error) {
	configDir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}
