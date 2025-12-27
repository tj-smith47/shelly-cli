// Package model defines core domain types for the Shelly CLI.
package model

// DeviceListItem represents device information for list command output.
type DeviceListItem struct {
	Name             string `json:"name" yaml:"name"`
	Address          string `json:"address" yaml:"address"`
	Platform         string `json:"platform" yaml:"platform"`
	Model            string `json:"model" yaml:"model"`
	Type             string `json:"type,omitempty" yaml:"type,omitempty"`
	Generation       int    `json:"generation" yaml:"generation"`
	Auth             bool   `json:"auth" yaml:"auth"`
	CurrentVersion   string `json:"current_version,omitempty" yaml:"current_version,omitempty"`
	AvailableVersion string `json:"available_version,omitempty" yaml:"available_version,omitempty"`
	HasUpdate        bool   `json:"has_update,omitempty" yaml:"has_update,omitempty"`
}
