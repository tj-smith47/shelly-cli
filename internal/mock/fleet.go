package mock

import (
	"time"

	"github.com/tj-smith47/shelly-go/integrator"
)

// FleetDeviceStatus represents a fleet device status for demo mode.
// This mirrors integrator.DeviceStatus from shelly-go.
type FleetDeviceStatus struct {
	DeviceID  string
	Name      string
	Model     string
	Online    bool
	Firmware  string
	LastSeen  time.Time
	Host      string
	GroupName string
}

// GetFleetDevices returns fleet device statuses from the current demo fixtures.
// Returns nil if not in demo mode or if fixtures aren't loaded.
func GetFleetDevices() []FleetDeviceStatus {
	demo := GetCurrentDemo()
	if demo == nil || demo.Fixtures == nil {
		return nil
	}

	devices := make([]FleetDeviceStatus, len(demo.Fixtures.Fleet.Devices))
	for i, d := range demo.Fixtures.Fleet.Devices {
		// Use Name as DeviceID for display since the table shows DeviceID
		deviceID := d.Name
		if deviceID == "" {
			deviceID = d.ID
		}
		devices[i] = FleetDeviceStatus{
			DeviceID: deviceID,
			Name:     d.Name,
			Model:    d.Model,
			Online:   d.Online,
			Firmware: d.Firmware,
			LastSeen: time.Now().Add(-time.Duration(i) * time.Minute), // Stagger last seen times
			Host:     "wss://api.shelly.cloud",
		}
	}
	return devices
}

// GetFleetOrganization returns the fleet organization name from demo fixtures.
func GetFleetOrganization() string {
	demo := GetCurrentDemo()
	if demo == nil || demo.Fixtures == nil {
		return ""
	}
	return demo.Fixtures.Fleet.Organization
}

// GetFleetDeviceStatuses returns fleet device statuses as integrator.DeviceStatus
// for direct use in display functions. Returns nil if not in demo mode.
func GetFleetDeviceStatuses() []*integrator.DeviceStatus {
	demo := GetCurrentDemo()
	if demo == nil || demo.Fixtures == nil {
		return nil
	}

	statuses := make([]*integrator.DeviceStatus, len(demo.Fixtures.Fleet.Devices))
	for i, d := range demo.Fixtures.Fleet.Devices {
		// Use Name as DeviceID for display since the table shows DeviceID
		deviceID := d.Name
		if deviceID == "" {
			deviceID = d.ID
		}
		statuses[i] = &integrator.DeviceStatus{
			DeviceID: deviceID,
			Online:   d.Online,
			LastSeen: time.Now().Add(-time.Duration(i) * time.Minute),
			Host:     "wss://api.shelly.cloud",
		}
	}
	return statuses
}
