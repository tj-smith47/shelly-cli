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
	ComponentRGBW   ComponentType = "rgbw"
	ComponentInput  ComponentType = "input"
)

// Component represents a device component (switch, cover, light, etc.).
type Component struct {
	Type ComponentType `json:"type" yaml:"type"`
	ID   int           `json:"id" yaml:"id"`
	Key  string        `json:"key" yaml:"key"` // Original key from device (e.g., "switch:0")
}

// SwitchStatus represents the status of a switch component.
type SwitchStatus struct {
	ID          int            `json:"id" yaml:"id"`
	Output      bool           `json:"output" yaml:"output"`
	Source      string         `json:"source" yaml:"source"`
	Power       *float64       `json:"power" yaml:"power"` // Active power in watts (nil if not available)
	Voltage     *float64       `json:"voltage" yaml:"voltage"`
	Current     *float64       `json:"current" yaml:"current"`
	Energy      *EnergyCounter `json:"energy" yaml:"energy"`
	Overtemp    bool           `json:"overtemp" yaml:"overtemp"`
	Overpower   bool           `json:"overpower" yaml:"overpower"`
	Overvoltage bool           `json:"overvoltage" yaml:"overvoltage"`
}

// SwitchConfig represents the configuration of a switch component.
type SwitchConfig struct {
	ID           int      `json:"id" yaml:"id"`
	Name         *string  `json:"name" yaml:"name"`
	InitialState string   `json:"initial_state" yaml:"initial_state"`
	AutoOn       bool     `json:"auto_on" yaml:"auto_on"`
	AutoOnDelay  float64  `json:"auto_on_delay" yaml:"auto_on_delay"`
	AutoOff      bool     `json:"auto_off" yaml:"auto_off"`
	AutoOffDelay float64  `json:"auto_off_delay" yaml:"auto_off_delay"`
	PowerLimit   *int     `json:"power_limit" yaml:"power_limit"`
	VoltageLimit *int     `json:"voltage_limit" yaml:"voltage_limit"`
	CurrentLimit *float64 `json:"current_limit" yaml:"current_limit"`
}

// CoverStatus represents the status of a cover component.
type CoverStatus struct {
	ID              int          `json:"id" yaml:"id"`
	State           string       `json:"state" yaml:"state"` // "open", "closed", "opening", "closing", "stopped"
	Source          string       `json:"source" yaml:"source"`
	CurrentPosition *int         `json:"current_pos" yaml:"current_pos"` // 0-100 percent
	TargetPosition  *int         `json:"target_pos" yaml:"target_pos"`
	MoveTimeout     bool         `json:"move_timeout" yaml:"move_timeout"`
	Calibrating     bool         `json:"calibrating" yaml:"calibrating"`
	Power           *float64     `json:"power" yaml:"power"`
	Voltage         *float64     `json:"voltage" yaml:"voltage"`
	Current         *float64     `json:"current" yaml:"current"`
	Safety          *CoverSafety `json:"safety" yaml:"safety"`
}

// CoverSafety represents cover safety status.
type CoverSafety struct {
	Obstacle    bool `json:"obstacle" yaml:"obstacle"`
	Overpower   bool `json:"overpower" yaml:"overpower"`
	Overtemp    bool `json:"overtemp" yaml:"overtemp"`
	Overvoltage bool `json:"overvoltage" yaml:"overvoltage"`
}

// CoverConfig represents the configuration of a cover component.
type CoverConfig struct {
	ID               int      `json:"id" yaml:"id"`
	Name             *string  `json:"name" yaml:"name"`
	InitialState     string   `json:"initial_state" yaml:"initial_state"`
	InvertDirections bool     `json:"invert_directions" yaml:"invert_directions"`
	MaxTime          *float64 `json:"max_time" yaml:"max_time"`
	MaxTimeOpen      *float64 `json:"max_time_open" yaml:"max_time_open"`
	MaxTimeClose     *float64 `json:"max_time_close" yaml:"max_time_close"`
	SwapInputs       bool     `json:"swap_inputs" yaml:"swap_inputs"`
}

// LightStatus represents the status of a light component.
type LightStatus struct {
	ID         int      `json:"id" yaml:"id"`
	Output     bool     `json:"output" yaml:"output"`
	Brightness *int     `json:"brightness" yaml:"brightness"` // 0-100
	Temp       *int     `json:"temp" yaml:"temp"`             // white color temperature in Kelvin (Gen1 white-temp bulbs)
	Source     string   `json:"source" yaml:"source"`
	Power      *float64 `json:"power" yaml:"power"`
	Voltage    *float64 `json:"voltage" yaml:"voltage"`
	Current    *float64 `json:"current" yaml:"current"`
	Overtemp   bool     `json:"overtemp" yaml:"overtemp"`
	Overpower  bool     `json:"overpower" yaml:"overpower"`
}

// LightConfig represents the configuration of a light component.
type LightConfig struct {
	ID              int     `json:"id" yaml:"id"`
	Name            *string `json:"name" yaml:"name"`
	InitialState    string  `json:"initial_state" yaml:"initial_state"`
	AutoOn          bool    `json:"auto_on" yaml:"auto_on"`
	AutoOnDelay     float64 `json:"auto_on_delay" yaml:"auto_on_delay"`
	AutoOff         bool    `json:"auto_off" yaml:"auto_off"`
	AutoOffDelay    float64 `json:"auto_off_delay" yaml:"auto_off_delay"`
	DefaultBright   int     `json:"default_bright" yaml:"default_bright"`
	NightModeEnable bool    `json:"night_mode_enable" yaml:"night_mode_enable"`
	NightModeBright int     `json:"night_mode_bright" yaml:"night_mode_bright"`
}

// RGBStatus represents the status of an RGB component.
type RGBStatus struct {
	ID         int       `json:"id" yaml:"id"`
	Output     bool      `json:"output" yaml:"output"`
	Brightness *int      `json:"brightness" yaml:"brightness"`
	RGB        *RGBColor `json:"rgb" yaml:"rgb"`
	Source     string    `json:"source" yaml:"source"`
	Power      *float64  `json:"power" yaml:"power"`
	Voltage    *float64  `json:"voltage" yaml:"voltage"`
	Current    *float64  `json:"current" yaml:"current"`
	Overtemp   bool      `json:"overtemp" yaml:"overtemp"`
	Overpower  bool      `json:"overpower" yaml:"overpower"`
}

// RGBColor represents RGB color values.
type RGBColor struct {
	Red   int `json:"red" yaml:"red"`
	Green int `json:"green" yaml:"green"`
	Blue  int `json:"blue" yaml:"blue"`
}

// RGBConfig represents the configuration of an RGB component.
type RGBConfig struct {
	ID              int     `json:"id" yaml:"id"`
	Name            *string `json:"name" yaml:"name"`
	InitialState    string  `json:"initial_state" yaml:"initial_state"`
	AutoOn          bool    `json:"auto_on" yaml:"auto_on"`
	AutoOnDelay     float64 `json:"auto_on_delay" yaml:"auto_on_delay"`
	AutoOff         bool    `json:"auto_off" yaml:"auto_off"`
	AutoOffDelay    float64 `json:"auto_off_delay" yaml:"auto_off_delay"`
	DefaultBright   int     `json:"default_bright" yaml:"default_bright"`
	NightModeEnable bool    `json:"night_mode_enable" yaml:"night_mode_enable"`
	NightModeBright int     `json:"night_mode_bright" yaml:"night_mode_bright"`
}

// RGBWStatus represents the status of an RGBW component.
type RGBWStatus struct {
	ID         int       `json:"id" yaml:"id"`
	Output     bool      `json:"output" yaml:"output"`
	Brightness *int      `json:"brightness" yaml:"brightness"`
	RGB        *RGBColor `json:"rgb" yaml:"rgb"`
	White      *int      `json:"white" yaml:"white"`
	Source     string    `json:"source" yaml:"source"`
	Power      *float64  `json:"power" yaml:"power"`
	Voltage    *float64  `json:"voltage" yaml:"voltage"`
	Current    *float64  `json:"current" yaml:"current"`
	Overtemp   bool      `json:"overtemp" yaml:"overtemp"`
	Overpower  bool      `json:"overpower" yaml:"overpower"`
}

// RGBWConfig represents the configuration of an RGBW component.
type RGBWConfig struct {
	ID              int     `json:"id" yaml:"id"`
	Name            *string `json:"name" yaml:"name"`
	InitialState    string  `json:"initial_state" yaml:"initial_state"`
	AutoOn          bool    `json:"auto_on" yaml:"auto_on"`
	AutoOnDelay     float64 `json:"auto_on_delay" yaml:"auto_on_delay"`
	AutoOff         bool    `json:"auto_off" yaml:"auto_off"`
	AutoOffDelay    float64 `json:"auto_off_delay" yaml:"auto_off_delay"`
	DefaultBright   int     `json:"default_bright" yaml:"default_bright"`
	DefaultWhite    int     `json:"default_white" yaml:"default_white"`
	NightModeEnable bool    `json:"night_mode_enable" yaml:"night_mode_enable"`
	NightModeBright int     `json:"night_mode_bright" yaml:"night_mode_bright"`
	NightModeWhite  int     `json:"night_mode_white" yaml:"night_mode_white"`
}

// EnergyCounter represents energy consumption data.
type EnergyCounter struct {
	Total    float64   `json:"total" yaml:"total"`         // Total energy in Wh
	ByMinute []float64 `json:"by_minute" yaml:"by_minute"` // Energy by minute
	MinuteTs int64     `json:"minute_ts" yaml:"minute_ts"` // Timestamp of minute data
}

// InputStatus represents the status of an input component.
type InputStatus struct {
	ID    int    `json:"id" yaml:"id"`
	State bool   `json:"state" yaml:"state"` // true = active (pressed/triggered)
	Type  string `json:"type" yaml:"type"`   // "button", "switch", etc.
}

// InputConfig represents the configuration of an input component.
type InputConfig struct {
	ID           int     `json:"id" yaml:"id"`
	Name         *string `json:"name" yaml:"name"`
	Type         string  `json:"type" yaml:"type"`
	Enable       bool    `json:"enable" yaml:"enable"`
	Invert       bool    `json:"invert" yaml:"invert"`
	FactoryReset bool    `json:"factory_reset" yaml:"factory_reset"`
}

// ComponentListItem represents a component in list output (for power/energy list commands).
type ComponentListItem struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
