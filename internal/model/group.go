// Package model defines core domain types for the Shelly CLI.
package model

// GroupInfo represents a device group for listing.
type GroupInfo struct {
	Name        string   `json:"name" yaml:"name"`
	DeviceCount int      `json:"device_count" yaml:"device_count"`
	Devices     []string `json:"devices" yaml:"devices"`
}
