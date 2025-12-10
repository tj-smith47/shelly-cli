// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// Config holds all CLI configuration.
type Config struct {
	// Global settings
	Output  string `mapstructure:"output"`
	Color   bool   `mapstructure:"color"`
	Theme   string `mapstructure:"theme"`
	APIMode string `mapstructure:"api_mode"`
	Verbose bool   `mapstructure:"verbose"`
	Quiet   bool   `mapstructure:"quiet"`

	// Discovery settings
	Discovery DiscoveryConfig `mapstructure:"discovery"`

	// Cloud settings
	Cloud CloudConfig `mapstructure:"cloud"`

	// Device registry
	Devices map[string]Device `mapstructure:"devices"`

	// Aliases
	Aliases map[string]Alias `mapstructure:"aliases"`

	// Groups
	Groups map[string]Group `mapstructure:"groups"`

	// Plugin settings
	Plugins PluginsConfig `mapstructure:"plugins"`
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

// Device represents a registered device.
type Device struct {
	Name       string `mapstructure:"name"`
	Address    string `mapstructure:"address"`
	Generation int    `mapstructure:"generation"`
	Type       string `mapstructure:"type"`
	Model      string `mapstructure:"model"`
	Auth       *Auth  `mapstructure:"auth,omitempty"`
}

// Auth holds device authentication credentials.
type Auth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
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
			Output:  "table",
			Color:   true,
			Theme:   "dracula",
			APIMode: "local",
			Devices: make(map[string]Device),
			Aliases: make(map[string]Alias),
			Groups:  make(map[string]Group),
		}
	}
	return cfg
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
			c.Devices = make(map[string]Device)
		}
		if c.Aliases == nil {
			c.Aliases = make(map[string]Alias)
		}
		if c.Groups == nil {
			c.Groups = make(map[string]Group)
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
	viper.Set("plugins", c.Plugins)

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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// ConfigDir returns the configuration directory path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "shelly"), nil
}

// PluginsDir returns the plugins directory path.
func PluginsDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "plugins"), nil
}
