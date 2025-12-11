// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"strings"
)

// RegisterDevice adds a device to the registry.
func RegisterDevice(name, address string, generation int, deviceType, model string, auth *Auth) error {
	if err := ValidateDeviceName(name); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	c.Devices[name] = Device{
		Name:       name,
		Address:    address,
		Generation: generation,
		Type:       deviceType,
		Model:      model,
		Auth:       auth,
	}
	cfgMu.Unlock()

	return Save()
}

// UnregisterDevice removes a device from the registry.
func UnregisterDevice(name string) error {
	c := Get()

	cfgMu.Lock()
	if _, ok := c.Devices[name]; !ok {
		cfgMu.Unlock()
		return fmt.Errorf("device %q not found", name)
	}
	delete(c.Devices, name)
	cfgMu.Unlock()

	// Also remove from any groups
	for groupName, group := range c.Groups {
		newDevices := make([]string, 0, len(group.Devices))
		for _, d := range group.Devices {
			if d != name {
				newDevices = append(newDevices, d)
			}
		}
		if len(newDevices) != len(group.Devices) {
			group.Devices = newDevices
			c.Groups[groupName] = group
		}
	}

	return Save()
}

// GetDevice returns a device by name.
func GetDevice(name string) (Device, bool) {
	c := Get()
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	device, ok := c.Devices[name]
	return device, ok
}

// ListDevices returns all registered devices.
func ListDevices() map[string]Device {
	c := Get()
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]Device, len(c.Devices))
	for k, v := range c.Devices {
		result[k] = v
	}
	return result
}

// RenameDevice renames a device.
func RenameDevice(oldName, newName string) error {
	if err := ValidateDeviceName(newName); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	device, ok := c.Devices[oldName]
	if !ok {
		cfgMu.Unlock()
		return fmt.Errorf("device %q not found", oldName)
	}

	if _, exists := c.Devices[newName]; exists {
		cfgMu.Unlock()
		return fmt.Errorf("device %q already exists", newName)
	}

	// Update device
	device.Name = newName
	delete(c.Devices, oldName)
	c.Devices[newName] = device

	// Update group references
	for groupName, group := range c.Groups {
		for i, d := range group.Devices {
			if d == oldName {
				group.Devices[i] = newName
				c.Groups[groupName] = group
				break
			}
		}
	}
	cfgMu.Unlock()

	return Save()
}

// ValidateDeviceName checks if a device name is valid.
func ValidateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	if strings.ContainsAny(name, " \t\n/\\:") {
		return fmt.Errorf("device name contains invalid characters")
	}

	return nil
}

// ResolveDevice resolves a device identifier to a Device.
// The identifier can be a device name (from registry) or an address.
func ResolveDevice(identifier string) (Device, error) {
	// First, check if it's a registered device name
	if device, ok := GetDevice(identifier); ok {
		return device, nil
	}

	// Otherwise, treat it as an address
	return Device{
		Name:    identifier,
		Address: identifier,
	}, nil
}

// CreateGroup creates a new device group.
func CreateGroup(name string) error {
	if err := ValidateGroupName(name); err != nil {
		return err
	}

	c := Get()

	cfgMu.Lock()
	if _, exists := c.Groups[name]; exists {
		cfgMu.Unlock()
		return fmt.Errorf("group %q already exists", name)
	}

	c.Groups[name] = Group{
		Name:    name,
		Devices: []string{},
	}
	cfgMu.Unlock()

	return Save()
}

// DeleteGroup deletes a device group.
func DeleteGroup(name string) error {
	c := Get()

	cfgMu.Lock()
	if _, ok := c.Groups[name]; !ok {
		cfgMu.Unlock()
		return fmt.Errorf("group %q not found", name)
	}
	delete(c.Groups, name)
	cfgMu.Unlock()

	return Save()
}

// GetGroup returns a group by name.
func GetGroup(name string) (Group, bool) {
	c := Get()
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	group, ok := c.Groups[name]
	return group, ok
}

// ListGroups returns all groups.
func ListGroups() map[string]Group {
	c := Get()
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	result := make(map[string]Group, len(c.Groups))
	for k, v := range c.Groups {
		result[k] = v
	}
	return result
}

// AddDeviceToGroup adds a device to a group.
func AddDeviceToGroup(groupName, deviceName string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	group, ok := c.Groups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found", groupName)
	}

	// Note: We allow adding unregistered devices (by address) to groups.
	// This provides flexibility for ad-hoc grouping.

	// Check if already in group
	for _, d := range group.Devices {
		if d == deviceName {
			return fmt.Errorf("device %q already in group %q", deviceName, groupName)
		}
	}

	group.Devices = append(group.Devices, deviceName)
	c.Groups[groupName] = group

	return Save()
}

// RemoveDeviceFromGroup removes a device from a group.
func RemoveDeviceFromGroup(groupName, deviceName string) error {
	c := Get()

	cfgMu.Lock()
	defer cfgMu.Unlock()

	group, ok := c.Groups[groupName]
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
	c.Groups[groupName] = group

	return Save()
}

// ValidateGroupName checks if a group name is valid.
func ValidateGroupName(name string) error {
	if name == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if strings.ContainsAny(name, " \t\n/\\:") {
		return fmt.Errorf("group name contains invalid characters")
	}

	return nil
}

// GetGroupDevices returns all devices in a group as Device structs.
func GetGroupDevices(groupName string) ([]Device, error) {
	group, ok := GetGroup(groupName)
	if !ok {
		return nil, fmt.Errorf("group %q not found", groupName)
	}

	devices := make([]Device, 0, len(group.Devices))
	for _, name := range group.Devices {
		device, err := ResolveDevice(name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve device %q: %w", name, err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}
