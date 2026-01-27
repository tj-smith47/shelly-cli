// Package wireless provides wireless protocol operations for Shelly devices.
package wireless

import (
	"github.com/tj-smith47/shelly-go/zwave"
)

// ZWaveInclusionMode represents a method for including a device in a Z-Wave network.
type ZWaveInclusionMode string

const (
	// ZWaveInclusionSmartStart uses the device QR code for automatic inclusion.
	ZWaveInclusionSmartStart ZWaveInclusionMode = "smart_start"

	// ZWaveInclusionButton uses the device S button for manual inclusion.
	ZWaveInclusionButton ZWaveInclusionMode = "button"

	// ZWaveInclusionSwitch uses the connected switch for manual inclusion.
	ZWaveInclusionSwitch ZWaveInclusionMode = "switch"
)

// ZWaveInclusionModes returns all available inclusion modes.
func ZWaveInclusionModes() []ZWaveInclusionMode {
	return []ZWaveInclusionMode{
		ZWaveInclusionSmartStart,
		ZWaveInclusionButton,
		ZWaveInclusionSwitch,
	}
}

// ZWaveInclusionModeName returns a human-readable name for a mode.
func ZWaveInclusionModeName(mode ZWaveInclusionMode) string {
	switch mode {
	case ZWaveInclusionSmartStart:
		return "SmartStart (QR Code)"
	case ZWaveInclusionButton:
		return "S Button"
	case ZWaveInclusionSwitch:
		return "Connected Switch"
	default:
		return string(mode)
	}
}

// ZWaveInclusionSteps returns inclusion instructions for the given mode.
func ZWaveInclusionSteps(mode ZWaveInclusionMode) []string {
	info := zwave.GetInclusionInfo(nil, zwave.InclusionMode(mode))
	return info.Instructions
}

// ZWaveExclusionSteps returns exclusion instructions for the given mode.
func ZWaveExclusionSteps(mode ZWaveInclusionMode) []string {
	info := zwave.GetExclusionInfo(nil, zwave.InclusionMode(mode))
	return info.Instructions
}

// ZWaveFactoryResetWarning returns the factory reset warning text.
func ZWaveFactoryResetWarning() string {
	info := zwave.GetFactoryResetInfo(nil)
	return info.Warning
}

// ZWaveFactoryResetSteps returns the factory reset steps.
func ZWaveFactoryResetSteps() []string {
	info := zwave.GetFactoryResetInfo(nil)
	return info.Instructions
}

// ZWaveConfigParam represents a Z-Wave configuration parameter.
type ZWaveConfigParam struct {
	Name         string
	Description  string
	Number       int
	Size         int
	DefaultValue int
	MinValue     int
	MaxValue     int
}

// ZWaveCommonConfigParams returns common Z-Wave configuration parameters.
func ZWaveCommonConfigParams() []ZWaveConfigParam {
	params := zwave.CommonConfigParameters()
	result := make([]ZWaveConfigParam, len(params))
	for i, p := range params {
		result[i] = ZWaveConfigParam{
			Name:         p.Name,
			Description:  p.Description,
			Number:       p.Number,
			Size:         p.Size,
			DefaultValue: p.DefaultValue,
			MinValue:     p.MinValue,
			MaxValue:     p.MaxValue,
		}
	}
	return result
}
