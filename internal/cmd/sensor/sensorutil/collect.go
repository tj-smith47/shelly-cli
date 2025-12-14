// Package sensorutil provides shared utilities for sensor commands.
package sensorutil

import (
	"encoding/json"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CollectByPrefix collects sensor data from a status map by component prefix.
// It iterates through all keys in the status map, finds those matching the prefix,
// and unmarshals them into the target type T.
//
// Example prefixes: "temperature:", "humidity:", "flood:", "smoke:", "illuminance:", "voltmeter:".
func CollectByPrefix[T any](status map[string]json.RawMessage, prefix string, ios *iostreams.IOStreams) []T {
	var sensors []T
	for key, raw := range status {
		if strings.HasPrefix(key, prefix) {
			var s T
			if err := json.Unmarshal(raw, &s); err != nil {
				if ios != nil {
					ios.Debug("failed to unmarshal sensor %s: %v", key, err)
				}
				continue
			}
			sensors = append(sensors, s)
		}
	}
	return sensors
}

// CollectByPrefixSilent is like CollectByPrefix but doesn't log unmarshal errors.
// Use this when you don't have access to IOStreams or errors are expected.
func CollectByPrefixSilent[T any](status map[string]json.RawMessage, prefix string) []T {
	return CollectByPrefix[T](status, prefix, nil)
}

// AlarmReading represents a sensor with alarm and mute state (flood, smoke).
type AlarmReading struct {
	ID    int  `json:"id"`
	Alarm bool `json:"alarm"`
	Mute  bool `json:"mute"`
}

// StyleFunc represents a styling function compatible with lipgloss Style.Render.
type StyleFunc func(...string) string

// DisplayAlarmSensors displays alarm-type sensors (flood, smoke) with a consistent format.
// Returns true if any sensors were displayed.
func DisplayAlarmSensors(ios *iostreams.IOStreams, sensors []AlarmReading, sensorType, alarmMsg string, okStyle, errorStyle, dimStyle StyleFunc) bool {
	if len(sensors) == 0 {
		return false
	}
	for _, s := range sensors {
		status := okStyle("Clear")
		if s.Alarm {
			status = errorStyle(alarmMsg)
		}
		muteStr := ""
		if s.Mute {
			muteStr = " " + dimStyle("(muted)")
		}
		ios.Printf("    Sensor %d: %s%s\n", s.ID, status, muteStr)
	}
	return true
}
