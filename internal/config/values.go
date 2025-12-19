package config

import (
	"fmt"
	"strings"
)

// ParseValue attempts to parse a string value into an appropriate type.
// It handles boolean, null, integer, float, and string values.
func ParseValue(s string) any {
	// Remove surrounding quotes if present
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	// Check for boolean
	lower := strings.ToLower(s)
	if lower == "true" || lower == "on" || lower == "yes" {
		return true
	}
	if lower == "false" || lower == "off" || lower == "no" {
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
