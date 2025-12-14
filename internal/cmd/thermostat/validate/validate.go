// Package validate provides validation utilities for thermostat commands.
package validate

import "fmt"

// ValidModes contains the valid thermostat operating modes.
var ValidModes = map[string]bool{
	"heat": true,
	"cool": true,
	"auto": true,
}

// ValidateMode validates that a thermostat mode is one of: heat, cool, auto.
// If allowEmpty is true, an empty string is also valid (used when mode is optional).
func ValidateMode(mode string, allowEmpty bool) error {
	if mode == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("mode is required, must be one of: heat, cool, auto")
	}

	if !ValidModes[mode] {
		return fmt.Errorf("invalid mode %q, must be one of: heat, cool, auto", mode)
	}

	return nil
}
