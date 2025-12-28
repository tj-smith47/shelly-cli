package config

import (
	"testing"
)

func TestKnownSettingKeys(t *testing.T) {
	t.Parallel()

	keys := KnownSettingKeys()
	if len(keys) == 0 {
		t.Error("KnownSettingKeys() returned empty slice")
	}

	// Check for expected keys
	expectedKeys := []string{
		"defaults.timeout",
		"defaults.output",
		"editor",
		"theme.name",
	}

	for _, expected := range expectedKeys {
		found := false
		for _, k := range keys {
			if k == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected key %q not found in KnownSettingKeys()", expected)
		}
	}
}

func TestFilterSettingKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prefix   string
		wantMin  int // minimum expected matches
		contains string
	}{
		{"defaults prefix", "defaults.", 2, "defaults.timeout"},
		{"theme prefix", "theme.", 2, "theme.name"},
		{"log prefix", "log.", 1, "log.json"},
		{"no match", "nonexistent.", 0, ""},
		{"empty prefix", "", 5, ""}, // should return all
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FilterSettingKeys(tt.prefix)
			if len(result) < tt.wantMin {
				t.Errorf("FilterSettingKeys(%q) returned %d items, want at least %d", tt.prefix, len(result), tt.wantMin)
			}
			if tt.contains != "" {
				found := false
				for _, k := range result {
					if k == tt.contains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FilterSettingKeys(%q) should contain %q", tt.prefix, tt.contains)
				}
			}
		})
	}
}

func TestFormatSettingValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string value", "hello", "hello"},
		{"empty string", "", ""},
		{"true bool", true, "true"},
		{"false bool", false, "false"},
		{"nil value", nil, "(not set)"},
		{"int value", 42, "42"},
		{"float value", 3.14, "3.14"},
		{"slice value", []string{"a", "b"}, "[a b]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatSettingValue(tt.input)
			if got != tt.want {
				t.Errorf("FormatSettingValue(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
