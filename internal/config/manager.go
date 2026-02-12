// Package config manages the CLI configuration using Viper.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-go/types"
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
		data = m.migrateSchemaURL(data)
		if err := yaml.Unmarshal(data, c); err != nil {
			return fmt.Errorf("unmarshal config: %w", err)
		}
	}

	initConfigMaps(c)
	deduplicateDevices(c)
	populateDerivedModels(c)

	m.config = c
	m.loaded = true
	return nil
}

// migrateSchemaURL fixes the schema URL from main→master in existing config files.
// Returns the (possibly modified) data. Persists the fix to disk if changed.
func (m *Manager) migrateSchemaURL(data []byte) []byte {
	if !bytes.Contains(data, []byte(oldSchemaURLFragment)) {
		return data
	}
	data = bytes.ReplaceAll(data,
		[]byte(oldSchemaURLFragment),
		[]byte("shelly-cli/master/cfg/config.schema.json"))
	// Persist the fix so editors pick it up immediately
	if err := afero.WriteFile(m.Fs(), m.path, data, 0o600); err != nil && viper.GetInt("verbosity") >= 1 {
		log.Printf("[debug] migrate schema URL: %v", err)
	}
	return data
}

// initConfigMaps initializes nil maps in a freshly-loaded config.
func initConfigMaps(c *Config) {
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
}

// deduplicateDevices removes duplicate device entries on config load.
// Two devices are considered duplicates if they share the same non-empty MAC address.
// When duplicates are found, the preferred key is kept using preferDeviceKey.
func deduplicateDevices(c *Config) {
	seen := make(map[string]string) // normalized MAC → preferred key
	for key := range c.Devices {
		mac := model.NormalizeMAC(c.Devices[key].MAC)
		if mac == "" {
			continue
		}
		if existingKey, exists := seen[mac]; exists {
			winner, loser := preferDeviceKey(existingKey, key)
			delete(c.Devices, loser)
			seen[mac] = winner
		} else {
			seen[mac] = key
		}
	}
}

// preferDeviceKey decides which device key to keep when two devices share a MAC.
// Prefers user-given names over auto-generated discovery names (which contain "shelly").
// Falls back to alphabetical order as the final tiebreaker.
func preferDeviceKey(a, b string) (winner, loser string) {
	aHasShelly := strings.Contains(strings.ToLower(a), "shelly")
	bHasShelly := strings.Contains(strings.ToLower(b), "shelly")

	switch {
	case aHasShelly && !bHasShelly:
		return b, a // prefer b (user-given name)
	case !aHasShelly && bHasShelly:
		return a, b // prefer a (user-given name)
	default:
		// Both or neither contain "shelly" — alphabetical tiebreaker
		if a < b {
			return a, b
		}
		return b, a
	}
}

// populateDerivedModels fills in empty Model fields from Type using shelly-go lookup.
func populateDerivedModels(c *Config) {
	for key, dev := range c.Devices {
		if dev.Model == "" && dev.Type != "" {
			dev.Model = types.ModelDisplayName(dev.Type)
			c.Devices[key] = dev
		}
	}
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
const yamlSchemaHeader = "# yaml-language-server: $schema=https://raw.githubusercontent.com/tj-smith47/shelly-cli/master/cfg/config.schema.json\n"

// oldSchemaURLFragment is the old schema URL fragment that needs to be auto-fixed in existing configs.
const oldSchemaURLFragment = "shelly-cli/main/cfg/config.schema.json"

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

	// Strip redundant Model fields before save (derivable from Type via shelly-go lookup)
	saveCfg := m.stripDerivedModels()

	data, err := marshalSorted(saveCfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Prepend YAML schema header for editor support (autocomplete, validation)
	fullData := append([]byte(yamlSchemaHeader), data...)

	if err := afero.WriteFile(fs, m.path, fullData, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Sync device registry to Viper so `config get devices` reflects changes
	m.syncDevicesToViper()

	return nil
}

// syncDevicesToViper syncs the Manager's device registry to Viper's in-memory state.
// This ensures `config get devices` reflects changes made via the Manager.
// Caller must hold m.mu.Lock().
func (m *Manager) syncDevicesToViper() {
	if m.config == nil {
		return
	}

	// Convert device map to mapstructure-compatible format for Viper
	deviceMap := make(map[string]any, len(m.config.Devices))
	for k, dev := range m.config.Devices {
		devMap := map[string]any{
			"address": dev.Address,
		}
		if dev.Name != "" {
			devMap["name"] = dev.Name
		}
		if dev.Generation > 0 {
			devMap["generation"] = dev.Generation
		}
		if dev.Type != "" {
			devMap["type"] = dev.Type
		}
		if dev.Model != "" {
			devMap["model"] = dev.Model
		}
		if dev.MAC != "" {
			devMap["mac"] = dev.MAC
		}
		if dev.Platform != "" {
			devMap["platform"] = dev.Platform
		}
		if dev.Auth != nil {
			devMap["auth"] = map[string]any{
				"username": dev.Auth.Username,
				"password": dev.Auth.Password,
			}
		}
		deviceMap[k] = devMap
	}
	viper.Set("devices", deviceMap)
}

// stripDerivedModels returns a copy of the config with Model fields removed
// where they can be derived from Type. The in-memory config is not modified.
func (m *Manager) stripDerivedModels() *Config {
	cfg := *m.config
	cfg.Devices = make(map[string]model.Device, len(m.config.Devices))
	for k, dev := range m.config.Devices {
		if dev.Type != "" && dev.Model != "" {
			if derived := types.ModelDisplayName(dev.Type); derived == dev.Model {
				dev.Model = "" // Don't persist — derivable from Type
			}
		}
		cfg.Devices[k] = dev
	}
	return &cfg
}

// marshalSorted marshals a value to YAML with top-level keys sorted alphabetically.
// Sub-keys in maps are already sorted by yaml.v3; struct field order is preserved
// for nested objects but the root-level mapping gets alphabetical sorting.
func marshalSorted(v any) ([]byte, error) {
	var node yaml.Node
	if err := node.Encode(v); err != nil {
		return nil, err
	}

	// node is a DocumentNode containing a MappingNode — sort the mapping keys
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		sortMappingKeys(node.Content[0])
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&node); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// sortMappingKeys sorts a YAML mapping node's keys alphabetically.
func sortMappingKeys(node *yaml.Node) {
	if node.Kind != yaml.MappingNode || len(node.Content) < 4 {
		return
	}

	type kvPair struct {
		key   *yaml.Node
		value *yaml.Node
	}
	pairs := make([]kvPair, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content)-1; i += 2 {
		pairs = append(pairs, kvPair{key: node.Content[i], value: node.Content[i+1]})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key.Value < pairs[j].key.Value
	})

	node.Content = make([]*yaml.Node, 0, len(pairs)*2)
	for _, p := range pairs {
		node.Content = append(node.Content, p.key, p.value)
	}
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
