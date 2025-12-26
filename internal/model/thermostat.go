// Package model defines core domain types for the Shelly CLI.
package model

// ThermostatInfo holds basic thermostat information for listing.
type ThermostatInfo struct {
	ID      int     `json:"id"`
	Enabled bool    `json:"enabled"`
	Mode    string  `json:"mode,omitempty"`
	TargetC float64 `json:"target_c,omitempty"`
}
