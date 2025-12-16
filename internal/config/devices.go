// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

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

// Package-level functions delegate to the default manager.

// RegisterDevice adds a device to the registry.
func RegisterDevice(name, address string, generation int, deviceType, deviceModel string, auth *model.Auth) error {
	return getDefaultManager().RegisterDevice(name, address, generation, deviceType, deviceModel, auth)
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
