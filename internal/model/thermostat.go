// Package model defines core domain types for the Shelly CLI.
package model

// ThermostatInfo holds basic thermostat information for listing.
type ThermostatInfo struct {
	ID int `json:"id"`
	// Enabled is the configured enable flag (config "enable"), not the live
	// valve state. A thermostat can be enabled while its valve is closed.
	Enabled bool `json:"enabled"`
	// Heating reports the live valve output (status "output") — true while the
	// thermostat is actively calling for heat.
	Heating bool    `json:"heating"`
	Mode    string  `json:"mode,omitempty"`
	TargetC float64 `json:"target_c,omitempty"`
}
