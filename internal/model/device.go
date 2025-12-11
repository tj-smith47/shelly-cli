// Package model defines core domain types for the Shelly CLI.
package model

// Device represents a Shelly device, either from the registry or ad-hoc.
type Device struct {
	Name       string
	Address    string
	Generation int
	Type       string
	Model      string
	Auth       *Auth
}

// Auth holds device authentication credentials.
type Auth struct {
	Username string
	Password string
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
