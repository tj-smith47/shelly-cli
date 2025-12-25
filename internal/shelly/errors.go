// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"errors"
	"fmt"
)

// Sentinel errors for plugin dispatch.
var (
	// ErrCommandNotSupported indicates a command is not supported for the device's platform.
	ErrCommandNotSupported = errors.New("command not supported for this platform")

	// ErrPluginNotFound indicates no plugin is installed for the device's platform.
	ErrPluginNotFound = errors.New("no plugin found for device platform")

	// ErrPluginHookMissing indicates the plugin doesn't implement the required hook.
	ErrPluginHookMissing = errors.New("plugin does not implement required hook")
)

// PlatformError provides detailed error information for platform-specific failures.
// It includes the platform name, command, and an optional hint for the user.
type PlatformError struct {
	Platform string
	Command  string
	Hint     string
}

// Error implements the error interface.
func (e *PlatformError) Error() string {
	msg := fmt.Sprintf("%q command is not supported for %s devices", e.Command, e.Platform)
	if e.Hint != "" {
		msg += fmt.Sprintf("\nHint: %s", e.Hint)
	}
	return msg
}

// NewPlatformError creates a new PlatformError.
// Hints should be provided by the plugin, not hardcoded in core.
func NewPlatformError(platform, command string) *PlatformError {
	return &PlatformError{
		Platform: platform,
		Command:  command,
	}
}

// NewPlatformErrorWithHint creates a new PlatformError with a hint.
// This allows plugins to provide platform-specific hints via their hooks.
func NewPlatformErrorWithHint(platform, command, hint string) *PlatformError {
	return &PlatformError{
		Platform: platform,
		Command:  command,
		Hint:     hint,
	}
}

// PluginNotFoundError provides detailed error when a plugin is not installed.
type PluginNotFoundError struct {
	Platform   string
	PluginName string
}

// Error implements the error interface.
func (e *PluginNotFoundError) Error() string {
	return fmt.Sprintf("no plugin found for platform %q (expected plugin: %s)\n"+
		"Install it with: shelly plugin install %s", e.Platform, e.PluginName, e.PluginName)
}

// Is allows errors.Is matching.
func (e *PluginNotFoundError) Is(target error) bool {
	return target == ErrPluginNotFound
}

// NewPluginNotFoundError creates a new PluginNotFoundError.
func NewPluginNotFoundError(platform string) *PluginNotFoundError {
	return &PluginNotFoundError{
		Platform:   platform,
		PluginName: "shelly-" + platform,
	}
}

// PluginHookMissingError provides detailed error when a plugin hook is missing.
type PluginHookMissingError struct {
	PluginName string
	Hook       string
}

// Error implements the error interface.
func (e *PluginHookMissingError) Error() string {
	return fmt.Sprintf("plugin %q does not implement the %q hook", e.PluginName, e.Hook)
}

// Is allows errors.Is matching.
func (e *PluginHookMissingError) Is(target error) bool {
	return target == ErrPluginHookMissing
}

// NewPluginHookMissingError creates a new PluginHookMissingError.
func NewPluginHookMissingError(pluginName, hook string) *PluginHookMissingError {
	return &PluginHookMissingError{
		PluginName: pluginName,
		Hook:       hook,
	}
}
