// Package config manages the CLI configuration using Viper.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// defaultFs is the package-level filesystem used for directory operations.
// This can be replaced in tests with an in-memory filesystem.
var (
	defaultFs   afero.Fs = afero.NewOsFs()
	defaultFsMu sync.RWMutex
)

// SetFs sets the package-level filesystem for testing.
// Pass nil to reset to the real OS filesystem.
func SetFs(fs afero.Fs) {
	defaultFsMu.Lock()
	defer defaultFsMu.Unlock()
	if fs == nil {
		defaultFs = afero.NewOsFs()
	} else {
		defaultFs = fs
	}
}

// getFs returns the current package-level filesystem.
func getFs() afero.Fs {
	defaultFsMu.RLock()
	defer defaultFsMu.RUnlock()
	return defaultFs
}

// Fs returns the package-level filesystem for use by other packages.
// In production, this returns the real OS filesystem.
// In tests, this can be replaced with an in-memory filesystem via SetFs.
func Fs() afero.Fs {
	return getFs()
}

// IsTestFs returns true if the package-level filesystem is a test filesystem
// (i.e., not the real OS filesystem). This is used to skip certain operations
// during tests, such as creating plugin directories.
func IsTestFs() bool {
	fs := getFs()
	_, isOsFs := fs.(*afero.OsFs)
	return !isOsFs
}

// Manager handles config loading, saving, and access with proper locking.
// It replaces the package-level global singleton to eliminate deadlocks
// and enable parallel testing.
//
// Entity-specific operations are organized in separate files:
//   - devices.go: Device and device alias operations
//   - groups.go: Group operations
//   - aliases.go: Command alias operations
//   - scenes.go: Scene operations
//   - template.go: Device and script template operations
//   - alerts.go: Alert operations
type Manager struct {
	mu     sync.RWMutex
	config *Config
	path   string
	loaded bool
	fs     afero.Fs // filesystem for file operations (nil uses package default)
}

// Fs returns the filesystem used by this manager.
// Returns the manager's fs if set, otherwise the package default.
func (m *Manager) Fs() afero.Fs {
	if m.fs != nil {
		return m.fs
	}
	return getFs()
}

// NewManager creates a config manager for the given path.
// If path is empty, uses default config directory from Dir().
func NewManager(path string) *Manager {
	if path == "" {
		dir, err := Dir()
		if err != nil {
			// Fall back to current directory if config dir is unavailable
			path = "config.yaml"
		} else {
			path = filepath.Join(dir, "config.yaml")
		}
	}
	return &Manager{path: path}
}

// NewTestManager creates a config manager with a pre-populated config for testing.
// The config is marked as loaded and won't be overwritten by Load().
func NewTestManager(cfg *Config) *Manager {
	if cfg.Devices == nil {
		cfg.Devices = make(map[string]model.Device)
	}
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]Alias)
	}
	if cfg.Groups == nil {
		cfg.Groups = make(map[string]Group)
	}
	if cfg.Links == nil {
		cfg.Links = make(map[string]Link)
	}
	if cfg.Scenes == nil {
		cfg.Scenes = make(map[string]Scene)
	}
	if cfg.Templates.Device == nil {
		cfg.Templates.Device = make(map[string]DeviceTemplate)
	}
	if cfg.Templates.Script == nil {
		cfg.Templates.Script = make(map[string]ScriptTemplate)
	}
	if cfg.Alerts == nil {
		cfg.Alerts = make(map[string]Alert)
	}
	return &Manager{
		config: cfg,
		loaded: true,
	}
}

// Load reads config from file. Safe to call multiple times.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		return nil
	}

	c := &Config{}

	// Read config file directly (not via viper) to enable parallel tests
	if data, err := afero.ReadFile(m.Fs(), m.path); err == nil {
		if err := yaml.Unmarshal(data, c); err != nil {
			return fmt.Errorf("unmarshal config: %w", err)
		}
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
	if c.Links == nil {
		c.Links = make(map[string]Link)
	}
	if c.Scenes == nil {
		c.Scenes = make(map[string]Scene)
	}
	if c.Templates.Device == nil {
		c.Templates.Device = make(map[string]DeviceTemplate)
	}
	if c.Templates.Script == nil {
		c.Templates.Script = make(map[string]ScriptTemplate)
	}
	if c.Alerts == nil {
		c.Alerts = make(map[string]Alert)
	}

	m.config = c
	m.loaded = true
	return nil
}

// Get returns the current config.
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Save persists config to file.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveWithoutLock()
}

// yamlSchemaHeader is prepended to the config file for YAML language server support.
// This provides autocomplete and validation in editors like VS Code.
const yamlSchemaHeader = "# yaml-language-server: $schema=https://raw.githubusercontent.com/tj-smith47/shelly-cli/main/cfg/config.schema.json\n"

// saveWithoutLock writes config to file. Caller must hold m.mu.Lock().
// For test managers (path is empty), this is a no-op.
func (m *Manager) saveWithoutLock() error {
	if m.config == nil {
		return errors.New("config not loaded")
	}

	// Skip disk write for test managers (no path set)
	if m.path == "" {
		return nil
	}

	fs := m.Fs()
	if err := fs.MkdirAll(filepath.Dir(m.path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Prepend YAML schema header for editor support (autocomplete, validation)
	fullData := append([]byte(yamlSchemaHeader), data...)

	if err := afero.WriteFile(fs, m.path, fullData, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// Reload forces a fresh load from file.
func (m *Manager) Reload() error {
	m.mu.Lock()
	m.loaded = false
	m.config = nil
	m.mu.Unlock()
	return m.Load()
}

// Path returns the config file path.
func (m *Manager) Path() string {
	return m.path
}
