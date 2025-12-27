// Package model defines core domain types for the Shelly CLI.
package model

// ComponentType represents the type of a Shelly component.
type ComponentType string

// Component types.
const (
	ComponentSwitch ComponentType = "switch"
	ComponentCover  ComponentType = "cover"
	ComponentLight  ComponentType = "light"
	ComponentRGB    ComponentType = "rgb"
	ComponentInput  ComponentType = "input"
)

// Component represents a device component (switch, cover, light, etc.).
type Component struct {
	Type ComponentType
	ID   int
	Key  string // Original key from device (e.g., "switch:0")
}

// SwitchStatus represents the status of a switch component.
type SwitchStatus struct {
	ID          int
	Output      bool
	Source      string
	Power       *float64 // Active power in watts (nil if not available)
	Voltage     *float64
	Current     *float64
	Energy      *EnergyCounter
	Overtemp    bool
	Overpower   bool
	Overvoltage bool
}

// SwitchConfig represents the configuration of a switch component.
type SwitchConfig struct {
	ID           int
	Name         *string
	InitialState string
	AutoOn       bool
	AutoOnDelay  float64
	AutoOff      bool
	AutoOffDelay float64
	PowerLimit   *int
	VoltageLimit *int
	CurrentLimit *float64
}

// CoverStatus represents the status of a cover component.
type CoverStatus struct {
	ID              int
	State           string // "open", "closed", "opening", "closing", "stopped"
	Source          string
	CurrentPosition *int // 0-100 percent
	TargetPosition  *int
	MoveTimeout     bool
	Calibrating     bool
	Power           *float64
	Voltage         *float64
	Current         *float64
	Safety          *CoverSafety
}

// CoverSafety represents cover safety status.
type CoverSafety struct {
	Obstacle    bool
	Overpower   bool
	Overtemp    bool
	Overvoltage bool
}

// CoverConfig represents the configuration of a cover component.
type CoverConfig struct {
	ID               int
	Name             *string
	InitialState     string
	InvertDirections bool
	MaxTime          *float64
	MaxTimeOpen      *float64
	MaxTimeClose     *float64
	SwapInputs       bool
}

// LightStatus represents the status of a light component.
type LightStatus struct {
	ID         int
	Output     bool
	Brightness *int // 0-100
	Source     string
	Power      *float64
	Voltage    *float64
	Current    *float64
	Overtemp   bool
	Overpower  bool
}

// LightConfig represents the configuration of a light component.
type LightConfig struct {
	ID              int
	Name            *string
	InitialState    string
	AutoOn          bool
	AutoOnDelay     float64
	AutoOff         bool
	AutoOffDelay    float64
	DefaultBright   int
	NightModeEnable bool
	NightModeBright int
}

// RGBStatus represents the status of an RGB component.
type RGBStatus struct {
	ID         int
	Output     bool
	Brightness *int
	RGB        *RGBColor
	Source     string
	Power      *float64
	Voltage    *float64
	Current    *float64
	Overtemp   bool
	Overpower  bool
}

// RGBColor represents RGB color values.
type RGBColor struct {
	Red   int
	Green int
	Blue  int
}

// RGBConfig represents the configuration of an RGB component.
type RGBConfig struct {
	ID              int
	Name            *string
	InitialState    string
	AutoOn          bool
	AutoOnDelay     float64
	AutoOff         bool
	AutoOffDelay    float64
	DefaultBright   int
	NightModeEnable bool
	NightModeBright int
}

// EnergyCounter represents energy consumption data.
type EnergyCounter struct {
	Total    float64   // Total energy in Wh
	ByMinute []float64 // Energy by minute
	MinuteTs int64     // Timestamp of minute data
}

// InputStatus represents the status of an input component.
type InputStatus struct {
	ID    int
	State bool   // true = active (pressed/triggered)
	Type  string // "button", "switch", etc.
}

// InputConfig represents the configuration of an input component.
type InputConfig struct {
	ID     int
	Name   *string
	Type   string
	Invert bool
}

// ComponentListItem represents a component in list output (for power/energy list commands).
type ComponentListItem struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
