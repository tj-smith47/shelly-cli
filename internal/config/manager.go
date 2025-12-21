// Package config manages the CLI configuration using Viper.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// Manager handles config loading, saving, and access with proper locking.
// It replaces the package-level global singleton to eliminate deadlocks
// and enable parallel testing.
type Manager struct {
	mu     sync.RWMutex
	config *Config
	path   string
	loaded bool
}

// NewManager creates a config manager for the given path.
// If path is empty, uses default (~/.config/shelly/config.yaml).
func NewManager(path string) *Manager {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fall back to current directory if home is unavailable
			path = "config.yaml"
		} else {
			path = filepath.Join(home, ".config", "shelly", "config.yaml")
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
	if cfg.Scenes == nil {
		cfg.Scenes = make(map[string]Scene)
	}
	if cfg.Templates == nil {
		cfg.Templates = make(map[string]Template)
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
	if data, err := os.ReadFile(m.path); err == nil {
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
	if c.Scenes == nil {
		c.Scenes = make(map[string]Scene)
	}
	if c.Templates == nil {
		c.Templates = make(map[string]Template)
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

// saveWithoutLock writes config to file. Caller must hold m.mu.Lock().
func (m *Manager) saveWithoutLock() error {
	if m.config == nil {
		return errors.New("config not loaded")
	}

	if err := os.MkdirAll(filepath.Dir(m.path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0o600); err != nil {
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

// =============================================================================
// Device Operations
// =============================================================================

// RegisterDevice adds a device to the registry.
// The name is normalized for use as a key (e.g., "Master Bathroom" → "master-bathroom")
// but the original display name is preserved in the Device struct.
func (m *Manager) RegisterDevice(name, address string, generation int, deviceType, deviceModel string, auth *model.Auth) error {
	if err := ValidateDeviceName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := NormalizeDeviceName(name)
	m.config.Devices[key] = model.Device{
		Name:       name, // Preserve original display name
		Address:    address,
		Generation: generation,
		Type:       deviceType,
		Model:      deviceModel,
		Auth:       auth,
	}
	return m.saveWithoutLock()
}

// UnregisterDevice removes a device from the registry.
// Accepts both display name ("Master Bathroom") and normalized key ("master-bathroom").
func (m *Manager) UnregisterDevice(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try exact match first, then normalized
	key := name
	if _, ok := m.config.Devices[key]; !ok {
		key = NormalizeDeviceName(name)
		if _, ok := m.config.Devices[key]; !ok {
			return fmt.Errorf("device %q not found", name)
		}
	}
	delete(m.config.Devices, key)

	// Also remove from any groups (check both forms)
	for groupName, group := range m.config.Groups {
		newDevices := make([]string, 0, len(group.Devices))
		for _, d := range group.Devices {
			if d != key && d != name {
				newDevices = append(newDevices, d)
			}
		}
		if len(newDevices) != len(group.Devices) {
			group.Devices = newDevices
			m.config.Groups[groupName] = group
		}
	}

	return m.saveWithoutLock()
}

// GetDevice returns a device by name.
// Accepts both display name ("Master Bathroom") and normalized key ("master-bathroom").
func (m *Manager) GetDevice(name string) (model.Device, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try exact match first
	if dev, ok := m.config.Devices[name]; ok {
		return dev, true
	}
	// Try normalized key
	if dev, ok := m.config.Devices[NormalizeDeviceName(name)]; ok {
		return dev, true
	}
	return model.Device{}, false
}

// ListDevices returns all registered devices.
func (m *Manager) ListDevices() map[string]model.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]model.Device, len(m.config.Devices))
	for k, v := range m.config.Devices {
		result[k] = v
	}
	return result
}

// RenameDevice renames a device.
// Accepts both display name and normalized key for oldName.
func (m *Manager) RenameDevice(oldName, newName string) error {
	if err := ValidateDeviceName(newName); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find old device (try exact match first, then normalized)
	oldKey := oldName
	device, ok := m.config.Devices[oldKey]
	if !ok {
		oldKey = NormalizeDeviceName(oldName)
		device, ok = m.config.Devices[oldKey]
		if !ok {
			return fmt.Errorf("device %q not found", oldName)
		}
	}

	newKey := NormalizeDeviceName(newName)
	if _, exists := m.config.Devices[newKey]; exists && newKey != oldKey {
		return fmt.Errorf("device %q already exists", newName)
	}

	// Update device with new display name
	device.Name = newName
	delete(m.config.Devices, oldKey)
	m.config.Devices[newKey] = device

	// Update group references
	for groupName, group := range m.config.Groups {
		for i, d := range group.Devices {
			if d == oldKey {
				group.Devices[i] = newKey
				m.config.Groups[groupName] = group
				break
			}
		}
	}

	return m.saveWithoutLock()
}

// ResolveDevice resolves a device identifier to a Device.
// The identifier can be a device name (from registry) or an address.
func (m *Manager) ResolveDevice(identifier string) (model.Device, error) {
	if device, ok := m.GetDevice(identifier); ok {
		return device, nil
	}

	// Otherwise, treat it as an address
	return model.Device{
		Name:    identifier,
		Address: identifier,
	}, nil
}

// SetDeviceAuth sets authentication credentials for a device.
// Accepts both display name and normalized key.
func (m *Manager) SetDeviceAuth(deviceName, username, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try exact match first, then normalized
	key := deviceName
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(deviceName)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", deviceName)
		}
	}

	dev.Auth = &model.Auth{
		Username: username,
		Password: password,
	}
	m.config.Devices[key] = dev
	return m.saveWithoutLock()
}

// GetAllDeviceCredentials returns credentials for all devices that have auth configured.
func (m *Manager) GetAllDeviceCredentials() map[string]struct{ Username, Password string } {
	m.mu.RLock()
	defer m.mu.RUnlock()

	creds := make(map[string]struct{ Username, Password string })
	for name, dev := range m.config.Devices {
		if dev.Auth != nil && dev.Auth.Password != "" {
			creds[name] = struct{ Username, Password string }{
				Username: dev.Auth.Username,
				Password: dev.Auth.Password,
			}
		}
	}
	return creds
}

// =============================================================================
// Group Operations
// =============================================================================

// CreateGroup creates a new device group.
func (m *Manager) CreateGroup(name string) error {
	if err := ValidateGroupName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Groups[name]; exists {
		return fmt.Errorf("group %q already exists", name)
	}

	m.config.Groups[name] = Group{
		Name:    name,
		Devices: []string{},
	}
	return m.saveWithoutLock()
}

// DeleteGroup deletes a device group.
func (m *Manager) DeleteGroup(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.config.Groups[name]; !ok {
		return fmt.Errorf("group %q not found", name)
	}
	delete(m.config.Groups, name)
	return m.saveWithoutLock()
}

// GetGroup returns a group by name.
func (m *Manager) GetGroup(name string) (Group, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	group, ok := m.config.Groups[name]
	return group, ok
}

// ListGroups returns all groups.
func (m *Manager) ListGroups() map[string]Group {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Group, len(m.config.Groups))
	for k, v := range m.config.Groups {
		result[k] = v
	}
	return result
}

// AddDeviceToGroup adds a device to a group.
func (m *Manager) AddDeviceToGroup(groupName, deviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.config.Groups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found", groupName)
	}

	// Check if already in group
	for _, d := range group.Devices {
		if d == deviceName {
			return fmt.Errorf("device %q already in group %q", deviceName, groupName)
		}
	}

	group.Devices = append(group.Devices, deviceName)
	m.config.Groups[groupName] = group
	return m.saveWithoutLock()
}

// RemoveDeviceFromGroup removes a device from a group.
func (m *Manager) RemoveDeviceFromGroup(groupName, deviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	group, ok := m.config.Groups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found", groupName)
	}

	found := false
	newDevices := make([]string, 0, len(group.Devices))
	for _, d := range group.Devices {
		if d == deviceName {
			found = true
		} else {
			newDevices = append(newDevices, d)
		}
	}

	if !found {
		return fmt.Errorf("device %q not in group %q", deviceName, groupName)
	}

	group.Devices = newDevices
	m.config.Groups[groupName] = group
	return m.saveWithoutLock()
}

// GetGroupDevices returns all devices in a group as Device structs.
func (m *Manager) GetGroupDevices(groupName string) ([]model.Device, error) {
	group, ok := m.GetGroup(groupName)
	if !ok {
		return nil, fmt.Errorf("group %q not found", groupName)
	}

	devices := make([]model.Device, 0, len(group.Devices))
	for _, name := range group.Devices {
		device, err := m.ResolveDevice(name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve device %q: %w", name, err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// =============================================================================
// Alias Operations
// =============================================================================

// AddAlias adds or updates an alias.
func (m *Manager) AddAlias(name, command string, shell bool) error {
	if err := ValidateAliasName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Aliases == nil {
		m.config.Aliases = make(map[string]Alias)
	}

	m.config.Aliases[name] = Alias{
		Name:    name,
		Command: command,
		Shell:   shell,
	}
	return m.saveWithoutLock()
}

// RemoveAlias removes an alias by name.
func (m *Manager) RemoveAlias(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Aliases[name]; !exists {
		return fmt.Errorf("alias %q not found", name)
	}
	delete(m.config.Aliases, name)
	return m.saveWithoutLock()
}

// GetAlias returns an alias by name.
func (m *Manager) GetAlias(name string) (Alias, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alias, ok := m.config.Aliases[name]
	return alias, ok
}

// ListAliases returns all aliases sorted by name.
func (m *Manager) ListAliases() []Alias {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Alias, 0, len(m.config.Aliases))
	for _, v := range m.config.Aliases {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListAliasesMap returns all aliases as a map.
func (m *Manager) ListAliasesMap() map[string]Alias {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Alias, len(m.config.Aliases))
	for k, v := range m.config.Aliases {
		result[k] = v
	}
	return result
}

// IsAlias checks if a command name is an alias.
func (m *Manager) IsAlias(name string) bool {
	_, ok := m.GetAlias(name)
	return ok
}

// ImportAliases imports aliases from a YAML file.
// Returns the number of imported aliases, skipped aliases, and any error.
// If merge is true, existing aliases are not overwritten.
func (m *Manager) ImportAliases(filename string, merge bool) (imported, skipped int, err error) {
	//nolint:gosec // G304: filename is user-provided intentionally (import command)
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	var af aliasFile
	if err := yaml.Unmarshal(data, &af); err != nil {
		return 0, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	// Validate all alias names before acquiring lock
	for name := range af.Aliases {
		if err := ValidateAliasName(name); err != nil {
			return 0, 0, fmt.Errorf("invalid alias %q: %w", name, err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Aliases == nil {
		m.config.Aliases = make(map[string]Alias)
	}

	for name, command := range af.Aliases {
		if merge {
			if _, exists := m.config.Aliases[name]; exists {
				skipped++
				continue
			}
		}

		// Detect shell aliases (prefixed with !)
		shell := false
		if command != "" && command[0] == '!' {
			shell = true
			command = command[1:]
		}

		m.config.Aliases[name] = Alias{
			Name:    name,
			Command: command,
			Shell:   shell,
		}
		imported++
	}

	if err := m.saveWithoutLock(); err != nil {
		return 0, 0, err
	}
	return imported, skipped, nil
}

// ExportAliases exports all aliases to a YAML file.
// If filename is empty, returns the YAML data as a string.
func (m *Manager) ExportAliases(filename string) (string, error) {
	aliases := m.ListAliases()

	af := aliasFile{
		Aliases: make(map[string]string, len(aliases)),
	}

	for _, a := range aliases {
		cmd := a.Command
		if a.Shell {
			cmd = "!" + cmd
		}
		af.Aliases[a.Name] = cmd
	}

	data, err := yaml.Marshal(&af)
	if err != nil {
		return "", fmt.Errorf("failed to marshal aliases: %w", err)
	}

	if filename == "" {
		return string(data), nil
	}

	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return "", nil
}

// =============================================================================
// Scene Operations
// =============================================================================

// CreateScene creates a new scene.
func (m *Manager) CreateScene(name, description string) error {
	if err := ValidateSceneName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Scenes[name]; exists {
		return fmt.Errorf("scene %q already exists", name)
	}

	m.config.Scenes[name] = Scene{
		Name:        name,
		Description: description,
		Actions:     []SceneAction{},
	}
	return m.saveWithoutLock()
}

// DeleteScene removes a scene.
func (m *Manager) DeleteScene(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Scenes[name]; !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	delete(m.config.Scenes, name)
	return m.saveWithoutLock()
}

// GetScene returns a scene by name.
func (m *Manager) GetScene(name string) (Scene, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scene, ok := m.config.Scenes[name]
	return scene, ok
}

// ListScenes returns all scenes.
func (m *Manager) ListScenes() map[string]Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Scene, len(m.config.Scenes))
	for k, v := range m.config.Scenes {
		result[k] = v
	}
	return result
}

// AddActionToScene adds an action to a scene.
func (m *Manager) AddActionToScene(sceneName string, action SceneAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = append(scene.Actions, action)
	m.config.Scenes[sceneName] = scene
	return m.saveWithoutLock()
}

// SetSceneActions replaces all actions in a scene.
func (m *Manager) SetSceneActions(sceneName string, actions []SceneAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[sceneName]
	if !exists {
		return fmt.Errorf("scene %q not found", sceneName)
	}

	scene.Actions = actions
	m.config.Scenes[sceneName] = scene
	return m.saveWithoutLock()
}

// UpdateScene updates a scene's name and/or description.
func (m *Manager) UpdateScene(oldName, newName, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scene, exists := m.config.Scenes[oldName]
	if !exists {
		return fmt.Errorf("scene %q not found", oldName)
	}

	if newName != "" && newName != oldName {
		if err := ValidateSceneName(newName); err != nil {
			return err
		}
		if _, exists := m.config.Scenes[newName]; exists {
			return fmt.Errorf("scene %q already exists", newName)
		}
		delete(m.config.Scenes, oldName)
		scene.Name = newName
	}

	if description != "" {
		scene.Description = description
	}

	m.config.Scenes[scene.Name] = scene
	return m.saveWithoutLock()
}

// SaveScene saves or updates a scene (used by import).
func (m *Manager) SaveScene(scene Scene) error {
	if err := ValidateSceneName(scene.Name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Scenes[scene.Name] = scene
	return m.saveWithoutLock()
}

// =============================================================================
// Template Operations
// =============================================================================

// CreateTemplate creates a new template.
func (m *Manager) CreateTemplate(name, description, deviceModel, app string, generation int, cfg map[string]any, sourceDevice string) error {
	if err := ValidateTemplateName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Templates[name]; exists {
		return fmt.Errorf("template %q already exists", name)
	}

	m.config.Templates[name] = Template{
		Name:         name,
		Description:  description,
		Model:        deviceModel,
		App:          app,
		Generation:   generation,
		Config:       cfg,
		CreatedAt:    time.Now().Format(time.RFC3339),
		SourceDevice: sourceDevice,
	}
	return m.saveWithoutLock()
}

// DeleteTemplate removes a template.
func (m *Manager) DeleteTemplate(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Templates[name]; !exists {
		return fmt.Errorf("template %q not found", name)
	}

	delete(m.config.Templates, name)
	return m.saveWithoutLock()
}

// GetTemplate returns a template by name.
func (m *Manager) GetTemplate(name string) (Template, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	template, ok := m.config.Templates[name]
	return template, ok
}

// ListTemplates returns all templates.
func (m *Manager) ListTemplates() map[string]Template {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Template, len(m.config.Templates))
	for k, v := range m.config.Templates {
		result[k] = v
	}
	return result
}

// UpdateTemplate updates a template's metadata.
func (m *Manager) UpdateTemplate(name, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	template, exists := m.config.Templates[name]
	if !exists {
		return fmt.Errorf("template %q not found", name)
	}

	if description != "" {
		template.Description = description
	}

	m.config.Templates[name] = template
	return m.saveWithoutLock()
}

// SaveTemplate saves or updates a template.
func (m *Manager) SaveTemplate(template Template) error {
	if err := ValidateTemplateName(template.Name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Templates[template.Name] = template
	return m.saveWithoutLock()
}

// =============================================================================
// Alert Operations
// =============================================================================

// CreateAlert creates a new alert.
func (m *Manager) CreateAlert(name, description, device, condition, action string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Alerts[name]; exists {
		return fmt.Errorf("alert %q already exists", name)
	}

	m.config.Alerts[name] = Alert{
		Name:        name,
		Description: description,
		Device:      device,
		Condition:   condition,
		Action:      action,
		Enabled:     enabled,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	return m.saveWithoutLock()
}

// DeleteAlert removes an alert.
func (m *Manager) DeleteAlert(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Alerts[name]; !exists {
		return fmt.Errorf("alert %q not found", name)
	}

	delete(m.config.Alerts, name)
	return m.saveWithoutLock()
}

// GetAlert returns an alert by name.
func (m *Manager) GetAlert(name string) (Alert, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alert, ok := m.config.Alerts[name]
	return alert, ok
}

// ListAlerts returns all alerts.
func (m *Manager) ListAlerts() map[string]Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Alert, len(m.config.Alerts))
	for k, v := range m.config.Alerts {
		result[k] = v
	}
	return result
}

// UpdateAlert updates an alert.
func (m *Manager) UpdateAlert(name string, enabled *bool, snoozedUntil string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.config.Alerts[name]
	if !exists {
		return fmt.Errorf("alert %q not found", name)
	}

	if enabled != nil {
		alert.Enabled = *enabled
	}
	if snoozedUntil != "" {
		alert.SnoozedUntil = snoozedUntil
	}

	m.config.Alerts[name] = alert
	return m.saveWithoutLock()
}
