// Package model defines core domain types for the Shelly CLI.
package model

import "encoding/json"

// CloudEvent represents a parsed cloud WebSocket event.
type CloudEvent struct {
	Event     string          `json:"event"`
	DeviceID  string          `json:"device_id,omitempty"`
	Device    string          `json:"device,omitempty"`
	Status    json.RawMessage `json:"status,omitempty"`
	Settings  json.RawMessage `json:"settings,omitempty"`
	Online    *int            `json:"online,omitempty"`
	Timestamp int64           `json:"ts,omitempty"`
}

// GetDeviceID returns the device identifier from the event.
// It prefers DeviceID over Device field for compatibility.
func (e *CloudEvent) GetDeviceID() string {
	if e.DeviceID != "" {
		return e.DeviceID
	}
	return e.Device
}
