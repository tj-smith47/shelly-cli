// Package model defines core domain types for the Shelly CLI.
package model

import "strings"

// PlatformShelly is the default platform for native Shelly devices.
// Any other platform value (e.g., "tasmota") indicates the device is managed
// by a plugin named "shelly-{platform}" (e.g., "shelly-tasmota").
const PlatformShelly = "shelly"

// Device represents a Shelly device, either from the registry or ad-hoc.
type Device struct {
	Name       string   `mapstructure:"name" json:"name,omitempty" yaml:"name,omitempty"`
	Address    string   `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`
	MAC        string   `mapstructure:"mac" json:"mac,omitempty" yaml:"mac,omitempty"`
	Aliases    []string `mapstructure:"aliases" json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Platform   string   `mapstructure:"platform" json:"platform,omitempty" yaml:"platform,omitempty"`
	Generation int      `mapstructure:"generation" json:"generation,omitempty" yaml:"generation,omitempty"`
	Type       string   `mapstructure:"type" json:"type,omitempty" yaml:"type,omitempty"`
	Model      string   `mapstructure:"model" json:"model,omitempty" yaml:"model,omitempty"`
	Auth       *Auth    `mapstructure:"auth,omitempty" json:"auth,omitempty" yaml:"auth,omitempty"`
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

// NormalizedMAC returns the MAC address in a consistent format (uppercase, colon-separated).
// Example: "aa:bb:cc:dd:ee:ff" or "AA-BB-CC-DD-EE-FF" → "AA:BB:CC:DD:EE:FF".
// Returns empty string if MAC is not set.
func (d Device) NormalizedMAC() string {
	return NormalizeMAC(d.MAC)
}

// HasAlias returns true if the device has the given alias (case-insensitive).
func (d Device) HasAlias(alias string) bool {
	for _, a := range d.Aliases {
		if strings.EqualFold(a, alias) {
			return true
		}
	}
	return false
}

// NormalizeMAC normalizes a MAC address to uppercase colon-separated format.
// Accepts formats: "aa:bb:cc:dd:ee:ff", "AA-BB-CC-DD-EE-FF", "aabbccddeeff".
// Returns empty string if the input is not a valid MAC address.
func NormalizeMAC(mac string) string {
	if mac == "" {
		return ""
	}

	// Remove common separators and convert to uppercase
	cleaned := strings.ToUpper(mac)
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")

	// Must be exactly 12 hex characters
	if len(cleaned) != 12 {
		return ""
	}

	// Validate all characters are hex
	for _, c := range cleaned {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return ""
		}
	}

	// Format as colon-separated
	return cleaned[0:2] + ":" + cleaned[2:4] + ":" + cleaned[4:6] + ":" +
		cleaned[6:8] + ":" + cleaned[8:10] + ":" + cleaned[10:12]
}
