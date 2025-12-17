// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"encoding/json"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// CollectSensorsByPrefix collects sensor data from a status map by component prefix.
// It iterates through all keys in the status map, finds those matching the prefix,
// and unmarshals them into the target type T.
//
// Example prefixes: "temperature:", "humidity:", "flood:", "smoke:", "illuminance:", "voltmeter:".
func CollectSensorsByPrefix[T any](status map[string]json.RawMessage, prefix string, ios *iostreams.IOStreams) []T {
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

// CollectSensorsByPrefixSilent is like CollectSensorsByPrefix but doesn't log unmarshal errors.
// Use this when you don't have access to IOStreams or errors are expected.
func CollectSensorsByPrefixSilent[T any](status map[string]json.RawMessage, prefix string) []T {
	return CollectSensorsByPrefix[T](status, prefix, nil)
}
