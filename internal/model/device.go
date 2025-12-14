// Package model defines core domain types for the Shelly CLI.
package model

// Device represents a Shelly device, either from the registry or ad-hoc.
type Device struct {
	Name       string `mapstructure:"name" json:"name,omitempty" yaml:"name,omitempty"`
	Address    string `mapstructure:"address" json:"address,omitempty" yaml:"address,omitempty"`
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
