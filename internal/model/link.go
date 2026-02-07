// Package model defines core domain types for the Shelly CLI.
package model

// LinkInfo represents a device link for listing.
type LinkInfo struct {
	ChildDevice  string `json:"child_device" yaml:"child_device"`
	ParentDevice string `json:"parent_device" yaml:"parent_device"`
	SwitchID     int    `json:"switch_id" yaml:"switch_id"`
}

// LinkStatus represents the resolved status of a linked device.
type LinkStatus struct {
	ChildDevice  string `json:"child_device" yaml:"child_device"`
	ParentDevice string `json:"parent_device" yaml:"parent_device"`
	SwitchID     int    `json:"switch_id" yaml:"switch_id"`
	ParentOnline bool   `json:"parent_online" yaml:"parent_online"`
	SwitchOutput bool   `json:"switch_output" yaml:"switch_output"`
	State        string `json:"state" yaml:"state"`
}
