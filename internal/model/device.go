// Package model defines core domain types for the Shelly CLI.
package model

// PlatformShelly is the default platform for native Shelly devices.
// Any other platform value (e.g., "tasmota") indicates the device is managed
// by a plugin named "shelly-{platform}" (e.g., "shelly-tasmota").
const PlatformShelly = "shelly"

// Device represents a Shelly device, either from the registry or ad-hoc.
type Device struct {
	Name       string `mapstructure:"name" json:"name,omitempty" yaml:"name,omitempty"`
	Address    string `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`
	Platform   string `mapstructure:"platform" json:"platform,omitempty" yaml:"platform,omitempty"`
	Generation int    `mapstructure:"generation" json:"generation,omitempty" yaml:"generation,omitempty"`
	Type       string `mapstructure:"type" json:"type,omitempty" yaml:"type,omitempty"`
	Model      string `mapstructure:"model" json:"model,omitempty" yaml:"model,omitempty"`
	Auth       *Auth  `mapstructure:"auth,omitempty" json:"auth,omitempty" yaml:"auth,omitempty"`
}

// Auth holds device authentication credentials.
type Auth struct {
	Username string `mapstructure:"username" json:"username,omitempty" yaml:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" yaml:"password,omitempty"`
}

// HasAuth returns true if the device has authentication configured.
func (d Device) HasAuth() bool {
	return d.Auth != nil && d.Auth.Password != ""
}

// DisplayName returns a human-readable name for the device.
func (d Device) DisplayName() string {
	if d.Name != "" {
		return d.Name
	}
	return d.Address
}

// IsShelly returns true if device is a native Shelly device.
// Empty platform defaults to Shelly for backward compatibility with existing configs.
func (d Device) IsShelly() bool {
	return d.Platform == "" || d.Platform == PlatformShelly
}

// IsPluginManaged returns true if device is managed by a plugin.
// Plugin-managed devices have a platform value that maps to a plugin name
// via the convention: platform "foo" → plugin "shelly-foo".
func (d Device) IsPluginManaged() bool {
	return d.Platform != "" && d.Platform != PlatformShelly
}

// GetPlatform returns the device platform, defaulting to "shelly" if empty.
func (d Device) GetPlatform() string {
	if d.Platform == "" {
		return PlatformShelly
	}
	return d.Platform
}

// PluginName returns the expected plugin name for this device's platform.
// Returns empty string for native Shelly devices.
// Example: platform "tasmota" → plugin name "shelly-tasmota".
func (d Device) PluginName() string {
	if d.IsShelly() {
		return ""
	}
	return "shelly-" + d.Platform
}
