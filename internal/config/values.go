package config

import (
	"fmt"
	"strings"
)

// Boolean literal values recognized when parsing and formatting settings.
const (
	valTrue  = "true"
	valFalse = "false"
	valOn    = "on"
	valYes   = "yes"
)

// unquote removes a single pair of surrounding double quotes if present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseBoolLenient parses a human-friendly boolean for a value already known to
// target a boolean field: true/false, on/off, yes/no, y/n, 1/0 (case-insensitive).
// Returns ok=false when the value is not a recognized boolean. Unlike ParseValue,
// it treats 1/0 as booleans because the destination type is known to be bool.
func parseBoolLenient(s string) (val, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case valTrue, valOn, valYes, "y", "1":
		return true, true
	case valFalse, cmdOff, "no", "n", "0":
		return false, true
	default:
		return false, false
	}
}

// ParseValue attempts to parse a string value into an appropriate type.
// It handles boolean, null, integer, float, and string values. This heuristic
// is used where the target type is not known ahead of time (e.g. device
// component config). For typed CLI settings, prefer CoerceSettingValue.
func ParseValue(s string) any {
	// Remove surrounding quotes if present
	if unq := unquote(s); unq != s {
		return unq
	}

	// Check for boolean
	lower := strings.ToLower(s)
	if lower == valTrue || lower == valOn || lower == valYes {
		return true
	}
	if lower == valFalse || lower == cmdOff || lower == "no" {
		return false
	}

	// Check for null
	if lower == "null" || lower == "nil" {
		return nil
	}

	// Try to parse as integer
	var i int64
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return i
	}

	// Try to parse as float
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err == nil {
		return f
	}

	// Return as string
	return s
}
