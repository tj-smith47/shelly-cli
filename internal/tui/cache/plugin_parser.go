// Package cache provides a shared device data cache for the TUI.
// This file contains parsing for plugin device status results.
package cache

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/debug"
)

// ParsePluginStatus converts a plugin DeviceStatusResult into a ParsedStatus.
// Components are keyed by "type:id" (e.g., "switch:0", "light:1").
func ParsePluginStatus(name string, result *plugins.DeviceStatusResult) *ParsedStatus {
	if result == nil {
		return nil
	}

	parsed := &ParsedStatus{
		SwitchPowers: make(map[int]float64),
		LightPowers:  make(map[int]float64),
		CoverPowers:  make(map[int]float64),
	}

	// Parse components using the plugin contract's generic field names.
	// Plugins translate their native format into the contract format (e.g., "output", "state", "power").
	// Supported component types: switch, light, cover, input (per plugin contract).
	for key, value := range result.Components {
		compType, compID, ok := pluginParseComponentKey(key)
		if !ok {
			debug.TraceEvent("cache: plugin %s: skipping unknown component key %q", name, key)
			continue
		}

		compMap, ok := value.(map[string]any)
		if !ok {
			debug.TraceEvent("cache: plugin %s: component %q value is not a map", name, key)
			continue
		}

		switch compType {
		case ComponentSwitch:
			sw := pluginParseSwitchMap(compID, compMap)
			parsed.Switches = append(parsed.Switches, sw)
			if power, ok := pluginGetFloat(compMap, "power"); ok {
				parsed.SwitchPowers[compID] = power
			}
		case ComponentLight:
			lt := pluginParseLightMap(compID, compMap)
			parsed.Lights = append(parsed.Lights, lt)
			if power, ok := pluginGetFloat(compMap, "power"); ok {
				parsed.LightPowers[compID] = power
			}
		case ComponentCover:
			parsed.Covers = append(parsed.Covers, pluginParseCoverMap(compID, compMap))
		case ComponentInput:
			parsed.Inputs = append(parsed.Inputs, pluginParseInputMap(compID, compMap))
		default:
			debug.TraceEvent("cache: plugin %s: unhandled component type %q", name, compType)
		}
	}

	// Parse energy
	if result.Energy != nil {
		parsed.Power = result.Energy.Power
		parsed.Voltage = result.Energy.Voltage
		parsed.Current = result.Energy.Current
		parsed.TotalEnergy = result.Energy.Total
	}

	// Parse sensors — plugin contract uses nested maps: {"value": float64, "unit": string}
	if temp, ok := pluginGetSensorValue(result.Sensors, "temperature"); ok {
		parsed.Temperature = temp
	}

	return parsed
}

// pluginParseComponentKey splits "switch:0" into ("switch", 0, true).
func pluginParseComponentKey(key string) (compType string, compID int, ok bool) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return "", 0, false
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, false
	}
	return parts[0], id, true
}

// pluginParseSwitchMap extracts switch state from a plugin component map.
func pluginParseSwitchMap(id int, m map[string]any) SwitchState {
	s := SwitchState{ID: id}
	if on, ok := m["output"].(bool); ok {
		s.On = on
	} else if state, ok := m["state"].(string); ok {
		// Some plugins use "state" string instead of "output" bool
		s.On = state == "on" || state == "ON"
	}
	if name, ok := m["name"].(string); ok {
		s.Name = name
	}
	if source, ok := m["source"].(string); ok {
		s.Source = source
	}
	return s
}

// pluginParseLightMap extracts light state from a plugin component map.
func pluginParseLightMap(id int, m map[string]any) LightState {
	l := LightState{ID: id}
	if on, ok := m["output"].(bool); ok {
		l.On = on
	} else if state, ok := m["state"].(string); ok {
		l.On = state == "on" || state == "ON"
	}
	if name, ok := m["name"].(string); ok {
		l.Name = name
	}
	return l
}

// pluginParseCoverMap extracts cover state from a plugin component map.
func pluginParseCoverMap(id int, m map[string]any) CoverState {
	c := CoverState{ID: id, State: CoverStateStopped}
	if state, ok := m["state"].(string); ok {
		c.State = state
	}
	if name, ok := m["name"].(string); ok {
		c.Name = name
	}
	return c
}

// pluginParseInputMap extracts input state from a plugin component map.
func pluginParseInputMap(id int, m map[string]any) InputState {
	i := InputState{ID: id}
	if state, ok := m["state"].(bool); ok {
		i.State = state
	}
	if typ, ok := m["type"].(string); ok {
		i.Type = typ
	}
	if name, ok := m["name"].(string); ok {
		i.Name = name
	}
	return i
}

// pluginGetFloat extracts a float64 from a map by key.
func pluginGetFloat(m map[string]any, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

// pluginGetSensorValue extracts a float64 from a plugin sensor map.
// Plugin sensors use nested maps: {"value": 23.5, "unit": "C"}.
// Falls back to reading the value directly as a float for simpler plugins.
func pluginGetSensorValue(sensors map[string]any, key string) (float64, bool) {
	if sensors == nil {
		return 0, false
	}
	v, ok := sensors[key]
	if !ok {
		return 0, false
	}
	// Plugin contract format: nested map with "value" field
	if m, ok := v.(map[string]any); ok {
		return pluginGetFloat(m, "value")
	}
	// Fallback: raw numeric value
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}

// BuildPluginDeviceInfo constructs a minimal DeviceInfo for plugin devices
// from config data, since plugins don't provide Shelly DeviceInfo.
func BuildPluginDeviceInfo(name string, device model.Device) *shelly.DeviceInfo {
	return &shelly.DeviceInfo{
		ID:         name,
		MAC:        device.MAC,
		Model:      device.Model,
		Generation: device.Generation,
		Firmware:   fmt.Sprintf("%s (plugin)", device.GetPlatform()),
		App:        device.GetPlatform(),
	}
}
