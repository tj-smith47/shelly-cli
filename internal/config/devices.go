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
