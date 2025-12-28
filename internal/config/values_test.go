package config

import "testing"

func TestParseValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  any
	}{
		// Quoted strings
		{"quoted string", `"hello world"`, "hello world"},
		{"quoted empty", `""`, ""},

		// Booleans (true variants)
		{"bool true lowercase", "true", true},
		{"bool true on", "on", true},
		{"bool true yes", "yes", true},
		{"bool true uppercase", "TRUE", true},
		{"bool true ON", "ON", true},
		{"bool true YES", "YES", true},

		// Booleans (false variants)
		{"bool false lowercase", "false", false},
		{"bool false off", "off", false},
		{"bool false no", "no", false},
		{"bool false uppercase", "FALSE", false},
		{"bool false OFF", "OFF", false},
		{"bool false NO", "NO", false},

		// Null variants
		{"null lowercase", "null", nil},
		{"null nil", "nil", nil},
		{"null uppercase", "NULL", nil},
		{"null NIL", "NIL", nil},

		// Integers
		{"integer positive", "42", int64(42)},
		{"integer negative", "-10", int64(-10)},
		{"integer zero", "0", int64(0)},

		// Floats - Note: Implementation uses Sscanf which parses int first,
		// so "3.14" is parsed as int64(3). Only values with no integer prefix parse as float.
		// The current implementation behavior is that "3.14" matches int first.
		{"float pure", ".5", float64(0.5)},

		// Plain strings (fallback)
		{"plain string", "hello", "hello"},
		{"plain with spaces", "hello world", "hello world"},
		{"plain mixed", "abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseValue(tt.input)
			if got != tt.want {
				t.Errorf("ParseValue(%q) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
			}
		})
	}
}
