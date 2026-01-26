// Package diagram provides wiring diagram models and rendering for Shelly devices.
package diagram

import (
	"fmt"
	"strings"
)

// Topology represents the wiring topology of a Shelly device.
type Topology int

const (
	// SingleRelay is a single relay output (L, N, SW, I, O).
	SingleRelay Topology = iota
	// DualRelay is a dual relay output (L, N, S1, S2, O1, O2).
	DualRelay
	// QuadRelay is a quad relay output (L, N, S1-S4, O1-O4).
	QuadRelay
	// Dimmer is a dimmer output (L, N, SW1, SW2, O).
	Dimmer
	// InputOnly is an input-only device (SW1-SWn).
	InputOnly
	// Plug is a plug-in device (no wiring).
	Plug
	// EnergyMonitor is an energy monitoring device (L, N, CT clamps).
	EnergyMonitor
	// RGBW is an RGBW LED controller (L, N, R, G, B, W).
	RGBW
)

// Style represents the rendering style for a wiring diagram.
type Style int

const (
	// StyleSchematic renders left-to-right circuit-style diagrams.
	StyleSchematic Style = iota
	// StyleCompact renders minimal terminal-focused box layouts.
	StyleCompact
	// StyleDetailed renders top-down installer-friendly layouts with annotations.
	StyleDetailed
)

// DeviceSpecs holds electrical specifications for a device model.
type DeviceSpecs struct {
	Voltage         string
	MaxAmps         float64
	PowerMonitoring bool
	NeutralRequired bool
	Notes           string
}

// DeviceModel represents a known Shelly device model with its wiring topology.
type DeviceModel struct {
	Slug       string
	Name       string
	Generation int
	Topology   Topology
	Specs      DeviceSpecs
	AltSlugs   []string
}

// ParseStyle converts a string to a Style constant.
// Valid values: "schematic", "compact", "detailed" (case-insensitive).
func ParseStyle(s string) (Style, error) {
	switch strings.ToLower(s) {
	case "schematic":
		return StyleSchematic, nil
	case "compact":
		return StyleCompact, nil
	case "detailed":
		return StyleDetailed, nil
	default:
		return 0, fmt.Errorf("invalid style %q: valid styles are %s", s, strings.Join(ValidStyles(), ", "))
	}
}

// ValidStyles returns all valid style names for flag completion.
func ValidStyles() []string {
	return []string{"schematic", "compact", "detailed"}
}

// ParseGeneration converts a string to a generation number.
// Valid values: "1", "2", "3", "4", "gen1", "gen2", "gen3", "gen4" (case-insensitive).
func ParseGeneration(s string) (int, error) {
	switch strings.ToLower(s) {
	case "1", "gen1":
		return 1, nil
	case "2", "gen2":
		return 2, nil
	case "3", "gen3":
		return 3, nil
	case "4", "gen4":
		return 4, nil
	default:
		return 0, fmt.Errorf("invalid generation %q: valid generations are %s", s, strings.Join(ValidGenerations(), ", "))
	}
}

// ValidGenerations returns all valid generation values for flag completion.
func ValidGenerations() []string {
	return []string{"1", "2", "3", "4", "gen1", "gen2", "gen3", "gen4"}
}
