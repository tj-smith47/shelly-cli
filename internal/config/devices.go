// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// nonAlphanumericRegex matches any character that isn't alphanumeric or dash.
var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9-]+`)

// NormalizeDeviceName converts a device name to a normalized key.
// Examples: "Master Bathroom" becomes "master-bathroom",
// "Living Room Light" becomes "living-room-light".
func NormalizeDeviceName(name string) string {
	// Lowercase
	normalized := strings.ToLower(name)
	// Replace spaces and underscores with dashes
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	// Remove any remaining invalid characters
	normalized = nonAlphanumericRegex.ReplaceAllString(normalized, "")
	// Collapse multiple dashes
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	// Trim leading/trailing dashes
	normalized = strings.Trim(normalized, "-")
	return normalized
}

// ValidateDeviceName checks if a device name is valid.
// Names can contain spaces (will be normalized to dashes for storage).
func ValidateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	// Only reject truly problematic characters for file paths/URLs
	if strings.ContainsAny(name, "/\\:") {
		return fmt.Errorf("device name cannot contain path separators or colons")
	}

	// Check that normalized name isn't empty
	if NormalizeDeviceName(name) == "" {
		return fmt.Errorf("device name must contain at least one alphanumeric character")
	}

	return nil
}

// ValidateGroupName checks if a group name is valid.
// Names can contain spaces (will be normalized to dashes for storage).
func ValidateGroupName(name string) error {
	if name == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	// Only reject truly problematic characters
	if strings.ContainsAny(name, "/\\:") {
		return fmt.Errorf("group name cannot contain path separators or colons")
	}

	// Check that normalized name isn't empty
	if NormalizeDeviceName(name) == "" {
		return fmt.Errorf("group name must contain at least one alphanumeric character")
	}

	return nil
}

// deviceAliasRegex validates device alias format: alphanumeric, hyphens, underscores.
var deviceAliasRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateDeviceAlias checks if a device alias is valid.
// Aliases must be 1-32 characters, starting with alphanumeric,
// containing only letters, numbers, hyphens, and underscores.
func ValidateDeviceAlias(alias string) error {
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}

	if len(alias) > 32 {
		return fmt.Errorf("alias cannot exceed 32 characters")
	}

	if !deviceAliasRegex.MatchString(alias) {
		return fmt.Errorf("alias must start with a letter or number and contain only letters, numbers, hyphens, and underscores")
	}

	return nil
}

// CheckAliasConflict checks if an alias conflicts with existing device names, keys, or aliases.
// excludeDevice is the device name to exclude from the check (for updates to existing device).
// Returns an error describing the conflict, or nil if no conflict.
func CheckAliasConflict(alias, excludeDevice string) error {
	return getDefaultManager().CheckAliasConflict(alias, excludeDevice)
}

// AddDeviceAlias adds an alias to a device.
func AddDeviceAlias(deviceName, alias string) error {
	return getDefaultManager().AddDeviceAlias(deviceName, alias)
}

// RemoveDeviceAlias removes an alias from a device.
func RemoveDeviceAlias(deviceName, alias string) error {
	return getDefaultManager().RemoveDeviceAlias(deviceName, alias)
}

// GetDeviceAliases returns all aliases for a device.
func GetDeviceAliases(deviceName string) ([]string, error) {
	return getDefaultManager().GetDeviceAliases(deviceName)
}

// Package-level functions delegate to the default manager.

// RegisterDevice adds a device to the registry.
func RegisterDevice(name, address string, generation int, deviceType, deviceModel string, auth *model.Auth) error {
	return getDefaultManager().RegisterDevice(name, address, generation, deviceType, deviceModel, auth)
}

// RegisterDeviceWithPlatform adds a device to the registry with platform support.
// Use this for plugin-managed devices (e.g., Tasmota, ESPHome).
func RegisterDeviceWithPlatform(name, address string, generation int, deviceType, deviceModel, platform string, auth *model.Auth) error {
	return getDefaultManager().RegisterDeviceWithPlatform(name, address, generation, deviceType, deviceModel, platform, auth)
}

// UpdateDeviceInfo updates device info fields without requiring full re-registration.
// Only non-empty/non-zero values are applied; empty values preserve existing data.
func UpdateDeviceInfo(name string, updates DeviceUpdates) error {
	return getDefaultManager().UpdateDeviceInfo(name, updates)
}

// UpdateDeviceAddress updates a device's IP address (for IP remapping).
func UpdateDeviceAddress(name, newAddress string) error {
	return getDefaultManager().UpdateDeviceAddress(name, newAddress)
}

// UnregisterDevice removes a device from the registry.
func UnregisterDevice(name string) error {
	return getDefaultManager().UnregisterDevice(name)
}

// GetDevice returns a device by name.
func GetDevice(name string) (model.Device, bool) {
	return getDefaultManager().GetDevice(name)
}

// ListDevices returns all registered devices.
func ListDevices() map[string]model.Device {
	return getDefaultManager().ListDevices()
}

// RenameDevice renames a device.
func RenameDevice(oldName, newName string) error {
	return getDefaultManager().RenameDevice(oldName, newName)
}

// ResolveDevice resolves a device identifier to a Device.
func ResolveDevice(identifier string) (model.Device, error) {
	return getDefaultManager().ResolveDevice(identifier)
}

// =============================================================================
// Manager Device Methods
// =============================================================================

// DeviceUpdates holds partial device info updates.
// Empty strings and zero values are ignored (existing values preserved).
type DeviceUpdates struct {
	Type       string // Device type/SKU (e.g., "SPSW-001PE16EU")
	Model      string // Human-readable model name (e.g., "Shelly Pro 1PM")
	Generation int    // Device generation (1, 2, 3, etc.)
	MAC        string // Device MAC address (normalized on save)
}

// RegisterDevice adds a device to the registry.
// The name is normalized for use as a key (e.g., "Master Bathroom" → "master-bathroom")
// but the original display name is preserved in the Device struct.
// For plugin-managed devices, use RegisterDeviceWithPlatform instead.
func (m *Manager) RegisterDevice(name, address string, generation int, deviceType, deviceModel string, auth *model.Auth) error {
	return m.RegisterDeviceWithPlatform(name, address, generation, deviceType, deviceModel, "", auth)
}

// RegisterDeviceWithPlatform adds a device to the registry with platform support.
// The name is normalized for use as a key (e.g., "Master Bathroom" → "master-bathroom")
// but the original display name is preserved in the Device struct.
// Empty platform defaults to "shelly" for native Shelly devices.
func (m *Manager) RegisterDeviceWithPlatform(name, address string, generation int, deviceType, deviceModel, platform string, auth *model.Auth) error {
	if err := ValidateDeviceName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	key := NormalizeDeviceName(name)
	m.config.Devices[key] = model.Device{
		Name:       name, // Preserve original display name
		Address:    address,
		Platform:   platform,
		Generation: generation,
		Type:       deviceType,
		Model:      deviceModel,
		Auth:       auth,
	}
	return m.saveWithoutLock()
}

// UpdateDeviceInfo updates device info fields without requiring full re-registration.
// Only non-empty/non-zero values are applied; empty values preserve existing data.
// This is more efficient than RegisterDevice for partial updates discovered at runtime.
func (m *Manager) UpdateDeviceInfo(name string, updates DeviceUpdates) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try exact match first, then normalized
	key := name
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(name)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", name)
		}
	}

	// Apply non-empty updates
	changed := false
	if updates.Type != "" && dev.Type != updates.Type {
		dev.Type = updates.Type
		changed = true
	}
	if updates.Model != "" && dev.Model != updates.Model {
		dev.Model = updates.Model
		changed = true
	}
	if updates.Generation > 0 && dev.Generation != updates.Generation {
		dev.Generation = updates.Generation
		changed = true
	}
	if updates.MAC != "" {
		// Normalize MAC address before storing
		normalizedMAC := model.NormalizeMAC(updates.MAC)
		if normalizedMAC != "" && dev.MAC != normalizedMAC {
			dev.MAC = normalizedMAC
			changed = true
		}
	}

	if !changed {
		return nil // No changes, skip disk write
	}

	m.config.Devices[key] = dev
	return m.saveWithoutLock()
}

// UpdateDeviceAddress updates a device's IP address (for IP remapping).
// This is a specialized method for silent IP remapping when a device's DHCP address changes.
func (m *Manager) UpdateDeviceAddress(name, newAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try exact match first, then normalized
	key := name
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(name)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", name)
		}
	}

	if dev.Address == newAddress {
		return nil // No change needed
	}

	dev.Address = newAddress
	m.config.Devices[key] = dev
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

// CheckAliasConflict checks if an alias conflicts with existing device names, keys, or aliases.
// excludeDevice is the device key to exclude from the check (for updates to existing device).
// Returns an error describing the conflict, or nil if no conflict.
func (m *Manager) CheckAliasConflict(alias, excludeDevice string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	aliasLower := strings.ToLower(alias)
	normalizedAlias := NormalizeDeviceName(alias)
	excludeKey := NormalizeDeviceName(excludeDevice)

	for key, dev := range m.config.Devices {
		// Skip the device being updated
		if key == excludeKey || key == excludeDevice {
			continue
		}

		// Check if alias matches device key
		if key == normalizedAlias || strings.EqualFold(key, alias) {
			return fmt.Errorf("alias %q conflicts with device key %q", alias, key)
		}

		// Check if alias matches device name
		if strings.EqualFold(dev.Name, alias) {
			return fmt.Errorf("alias %q conflicts with device name %q", alias, dev.Name)
		}

		// Check if alias matches any existing alias on other devices
		for _, existingAlias := range dev.Aliases {
			if strings.EqualFold(existingAlias, aliasLower) {
				return fmt.Errorf("alias %q already used by device %q", alias, dev.Name)
			}
		}
	}

	return nil
}

// AddDeviceAlias adds an alias to a device.
// Validates the alias format and checks for conflicts with other devices.
func (m *Manager) AddDeviceAlias(deviceName, alias string) error {
	// Validate alias format
	if err := ValidateDeviceAlias(alias); err != nil {
		return err
	}

	// Check for conflicts (excluding the target device)
	if err := m.CheckAliasConflict(alias, deviceName); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the device
	key := deviceName
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(deviceName)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", deviceName)
		}
	}

	// Check if alias already exists on this device
	for _, existing := range dev.Aliases {
		if strings.EqualFold(existing, alias) {
			return fmt.Errorf("alias %q already exists on device %q", alias, dev.Name)
		}
	}

	dev.Aliases = append(dev.Aliases, alias)
	m.config.Devices[key] = dev
	return m.saveWithoutLock()
}

// RemoveDeviceAlias removes an alias from a device.
func (m *Manager) RemoveDeviceAlias(deviceName, alias string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the device
	key := deviceName
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(deviceName)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", deviceName)
		}
	}

	// Find and remove the alias
	found := false
	newAliases := make([]string, 0, len(dev.Aliases))
	for _, existing := range dev.Aliases {
		if strings.EqualFold(existing, alias) {
			found = true
		} else {
			newAliases = append(newAliases, existing)
		}
	}

	if !found {
		return fmt.Errorf("alias %q not found on device %q", alias, dev.Name)
	}

	dev.Aliases = newAliases
	m.config.Devices[key] = dev
	return m.saveWithoutLock()
}

// GetDeviceAliases returns all aliases for a device.
func (m *Manager) GetDeviceAliases(deviceName string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the device
	dev, ok := m.config.Devices[deviceName]
	if !ok {
		dev, ok = m.config.Devices[NormalizeDeviceName(deviceName)]
		if !ok {
			return nil, fmt.Errorf("device %q not found", deviceName)
		}
	}

	// Return a copy to prevent mutation
	result := make([]string, len(dev.Aliases))
	copy(result, dev.Aliases)
	return result, nil
}

// ResolveDevice resolves a device identifier to a Device.
// Resolution order (stops at first match):
//  1. Exact key match ("master-bathroom")
//  2. Normalized key match (e.g., "Master Bathroom" → "master-bathroom")
//  3. Display name match (case-insensitive)
//  4. Alias match (any in device.Aliases array)
//  5. MAC address match
//  6. Fallback: treat as direct address
func (m *Manager) ResolveDevice(identifier string) (model.Device, error) {
	// 1-2. Exact or normalized key match (handled by GetDevice)
	if device, ok := m.GetDevice(identifier); ok {
		return device, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Normalize identifier for comparisons
	identifierLower := strings.ToLower(identifier)
	normalizedMAC := model.NormalizeMAC(identifier)

	for _, dev := range m.config.Devices {
		// 3. Display name match (case-insensitive)
		if strings.EqualFold(dev.Name, identifier) {
			return dev, nil
		}

		// 4. Alias match
		if dev.HasAlias(identifier) {
			return dev, nil
		}

		// 5. MAC address match (normalized comparison)
		if normalizedMAC != "" && dev.MAC != "" {
			if model.NormalizeMAC(dev.MAC) == normalizedMAC {
				return dev, nil
			}
		}

		// Also check if identifier matches normalized MAC format directly
		if dev.MAC != "" && strings.EqualFold(dev.MAC, identifierLower) {
			return dev, nil
		}
	}

	// 6. Fallback: treat as direct address
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

// UpdateDeviceComponents updates the cached component names for a device.
// The components map structure is: component type ("switch", "light", etc.) -> map of ID -> name.
func UpdateDeviceComponents(deviceName string, components map[string]map[int]string) error {
	return getDefaultManager().UpdateDeviceComponents(deviceName, components)
}

// UpdateDeviceComponents updates the cached component names for a device.
// The components map structure is: component type ("switch", "light", etc.) -> map of ID -> name.
// Example: {"switch": {0: "Kitchen Light", 1: "Living Room"}}.
func (m *Manager) UpdateDeviceComponents(deviceName string, components map[string]map[int]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find the device (try exact match first, then normalized)
	key := deviceName
	dev, ok := m.config.Devices[key]
	if !ok {
		key = NormalizeDeviceName(deviceName)
		dev, ok = m.config.Devices[key]
		if !ok {
			return fmt.Errorf("device %q not found", deviceName)
		}
	}

	// Check if components changed to avoid unnecessary disk writes
	if componentsEqual(dev.Components, components) {
		return nil
	}

	// Deep copy the components map
	dev.Components = make(map[string]map[int]string)
	for typeKey, idMap := range components {
		dev.Components[typeKey] = make(map[int]string)
		for id, name := range idMap {
			dev.Components[typeKey][id] = name
		}
	}

	m.config.Devices[key] = dev
	return m.saveWithoutLock()
}

// componentsEqual compares two component name maps for equality.
func componentsEqual(a, b map[string]map[int]string) bool {
	if len(a) != len(b) {
		return false
	}
	for typeKey, aIDMap := range a {
		bIDMap, ok := b[typeKey]
		if !ok || len(aIDMap) != len(bIDMap) {
			return false
		}
		for id, name := range aIDMap {
			if bIDMap[id] != name {
				return false
			}
		}
	}
	return true
}

// GetDeviceComponents returns the cached component names for a device.
// Returns nil if no components are cached.
func GetDeviceComponents(deviceName string) map[string]map[int]string {
	return getDefaultManager().GetDeviceComponents(deviceName)
}

// GetDeviceComponents returns the cached component names for a device.
// Returns nil if no components are cached.
func (m *Manager) GetDeviceComponents(deviceName string) map[string]map[int]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the device
	dev, ok := m.config.Devices[deviceName]
	if !ok {
		dev, ok = m.config.Devices[NormalizeDeviceName(deviceName)]
		if !ok {
			return nil
		}
	}

	if dev.Components == nil {
		return nil
	}

	// Deep copy to prevent mutation
	result := make(map[string]map[int]string)
	for typeKey, idMap := range dev.Components {
		result[typeKey] = make(map[int]string)
		for id, name := range idMap {
			result[typeKey][id] = name
		}
	}
	return result
}
