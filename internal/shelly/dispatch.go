// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// PluginQuickResult holds the result of a plugin quick operation.
// This is the plugin-equivalent of QuickResult.
type PluginQuickResult struct {
	// Device is the name/identifier of the device.
	Device string
	// Component is the component type that was controlled.
	Component string
	// State is the resulting state (e.g., "on", "off").
	State string
	// Success indicates if the operation was successful.
	Success bool
}

// dispatchToPlugin executes a control action through a plugin.
// This is called when a device is plugin-managed (device.IsPluginManaged() returns true).
func (s *Service) dispatchToPlugin(ctx context.Context, device model.Device, action, component string, id *int) (*PluginQuickResult, error) {
	// Check if plugin registry is configured
	if s.pluginRegistry == nil {
		return nil, NewPluginNotFoundError(device.Platform)
	}

	// Find the plugin for this platform
	plugin, err := s.pluginRegistry.FindByPlatform(device.Platform)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, NewPluginNotFoundError(device.Platform)
	}

	// Check if plugin has control hook
	if plugin.Manifest == nil || plugin.Manifest.Hooks == nil || plugin.Manifest.Hooks.Control == "" {
		return nil, NewPluginHookMissingError(plugin.Name, "control")
	}

	// Execute the control hook
	executor := plugins.NewHookExecutor(plugin)
	compID := 0
	if id != nil {
		compID = *id
	}

	result, err := executor.ExecuteControl(ctx, device.Address, device.Auth, action, component, compID)
	if err != nil {
		return nil, err
	}

	// Convert to PluginQuickResult
	return &PluginQuickResult{
		Device:    device.DisplayName(),
		Component: component,
		State:     result.State,
		Success:   result.Success,
	}, nil
}

// GetPluginDeviceStatus retrieves status from a plugin-managed device.
func (s *Service) GetPluginDeviceStatus(ctx context.Context, device model.Device) (*plugins.DeviceStatusResult, error) {
	// Check if plugin registry is configured
	if s.pluginRegistry == nil {
		return nil, NewPluginNotFoundError(device.Platform)
	}

	// Find the plugin for this platform
	plugin, err := s.pluginRegistry.FindByPlatform(device.Platform)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, NewPluginNotFoundError(device.Platform)
	}

	// Check if plugin has status hook
	if plugin.Manifest == nil || plugin.Manifest.Hooks == nil || plugin.Manifest.Hooks.Status == "" {
		return nil, NewPluginHookMissingError(plugin.Name, "status")
	}

	// Execute the status hook
	executor := plugins.NewHookExecutor(plugin)
	return executor.ExecuteStatus(ctx, device.Address, device.Auth)
}

// SupportsPluginCommand checks if a plugin supports a specific command.
// Returns a PlatformError if the command is not supported by the plugin.
// Command support is determined by the plugin's hooks - if a plugin doesn't
// have the appropriate hook, the command is not supported.
func (s *Service) SupportsPluginCommand(device model.Device, command string) error {
	if !device.IsPluginManaged() {
		// Shelly devices support all commands
		return nil
	}

	// For plugin-managed devices, command support is checked at dispatch time
	// via hook availability. This function provides an early check.
	if s.pluginRegistry == nil {
		return NewPluginNotFoundError(device.Platform)
	}

	plugin, err := s.pluginRegistry.FindByPlatform(device.Platform)
	if err != nil {
		return err
	}
	if plugin == nil {
		return NewPluginNotFoundError(device.Platform)
	}

	// Check if the plugin has the required hook for this command type
	if plugin.Manifest == nil || plugin.Manifest.Hooks == nil {
		return NewPlatformError(device.Platform, command)
	}

	// Map command to hook
	var hookAvailable bool
	switch command {
	case "on", "off", "toggle", "switch", "light", "cover":
		hookAvailable = plugin.Manifest.Hooks.Control != ""
	case "status":
		hookAvailable = plugin.Manifest.Hooks.Status != ""
	case "firmware", "update":
		hookAvailable = plugin.Manifest.Hooks.CheckUpdates != "" || plugin.Manifest.Hooks.ApplyUpdate != ""
	default:
		// Unknown commands are assumed unsupported for plugins
		hookAvailable = false
	}

	if !hookAvailable {
		// Look up hint from plugin manifest if available
		hint := ""
		if plugin.Manifest.Capabilities != nil && plugin.Manifest.Capabilities.Hints != nil {
			hint = plugin.Manifest.Capabilities.Hints[command]
		}
		if hint != "" {
			return NewPlatformErrorWithHint(device.Platform, command, hint)
		}
		return NewPlatformError(device.Platform, command)
	}

	return nil
}
