// Package bthomeutil provides shared utilities for BTHome commands.
package bthomeutil

import (
	"encoding/json"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DeviceStatus represents BTHome device status from Shelly.GetStatus.
type DeviceStatus struct {
	ID         int     `json:"id"`
	RSSI       *int    `json:"rssi"`
	Battery    *int    `json:"battery"`
	LastUpdate float64 `json:"last_updated_ts"`
}

// CollectDevices collects BTHomeDevice components from device status.
// Returns a slice of DeviceStatus for all bthomedevice:* keys in the status map.
func CollectDevices(status map[string]json.RawMessage, ios *iostreams.IOStreams) []DeviceStatus {
	var devices []DeviceStatus

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomedevice:") {
			continue
		}

		var devStatus DeviceStatus
		if err := json.Unmarshal(raw, &devStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome device %s: %v", key, err)
			}
			continue
		}

		devices = append(devices, devStatus)
	}

	return devices
}

// SensorStatus represents BTHome sensor status from Shelly.GetStatus.
type SensorStatus struct {
	ID         int     `json:"id"`
	RSSI       *int    `json:"rssi"`
	Battery    *int    `json:"battery"`
	LastUpdate float64 `json:"last_updated_ts"`
}

// CollectSensors collects BTHomeSensor components from device status.
// Returns a slice of SensorStatus for all bthomesensor:* keys in the status map.
func CollectSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []SensorStatus {
	var sensors []SensorStatus

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomesensor:") {
			continue
		}

		var sensorStatus SensorStatus
		if err := json.Unmarshal(raw, &sensorStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome sensor %s: %v", key, err)
			}
			continue
		}

		sensors = append(sensors, sensorStatus)
	}

	return sensors
}
