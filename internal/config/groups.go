package config

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// =============================================================================
// Package-level Group Functions (delegate to default manager)
// =============================================================================

// CreateGroup creates a new device group.
func CreateGroup(name string) error {
	return getDefaultManager().CreateGroup(name)
}

// DeleteGroup deletes a device group.
func DeleteGroup(name string) error {
	return getDefaultManager().DeleteGroup(name)
}

// GetGroup returns a group by name.
func GetGroup(name string) (Group, bool) {
	return getDefaultManager().GetGroup(name)
}

// ListGroups returns all groups.
func ListGroups() map[string]Group {
	return getDefaultManager().ListGroups()
}

// AddDeviceToGroup adds a device to a group.
func AddDeviceToGroup(groupName, deviceName string) error {
	return getDefaultManager().AddDeviceToGroup(groupName, deviceName)
}

// RemoveDeviceFromGroup removes a device from a group.
func RemoveDeviceFromGroup(groupName, deviceName string) error {
	return getDefaultManager().RemoveDeviceFromGroup(groupName, deviceName)
}

// GetGroupDevices returns all devices in a group as Device structs.
func GetGroupDevices(groupName string) ([]model.Device, error) {
	return getDefaultManager().GetGroupDevices(groupName)
}

// =============================================================================
// Manager Group Methods
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
