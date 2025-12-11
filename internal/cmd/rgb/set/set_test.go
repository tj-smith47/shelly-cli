// Package set provides the rgb set subcommand.
package set

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd == nil {
		t.Fatal("NewCommand() returned nil")
	}

	if cmd.Use != "set <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
		defValue  string
	}{
		{"id", "id", "i", "0"},
		{"red", "red", "r", "-1"},
		{"green", "green", "g", "-1"},
		{"blue", "blue", "b", "-1"},
		{"brightness", "brightness", "", "-1"},
		{"on", "on", "", "false"},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.flagName)
		if flag == nil {
			t.Errorf("%s flag not found", tt.flagName)
			continue
		}
		if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
			t.Errorf("%s shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
		}
		if flag.DefValue != tt.defValue {
			t.Errorf("%s default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand()

	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestBuildParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		red        int
		green      int
		blue       int
		brightness int
		on         bool
		wantRed    bool
		wantGreen  bool
		wantBlue   bool
		wantBright bool
		wantOn     bool
	}{
		{
			name:       "no params set",
			red:        -1,
			green:      -1,
			blue:       -1,
			brightness: -1,
			on:         false,
		},
		{
			name:      "all colors set",
			red:       255,
			green:     128,
			blue:      64,
			wantRed:   true,
			wantGreen: true,
			wantBlue:  true,
		},
		{
			name:       "brightness set",
			red:        -1,
			green:      -1,
			blue:       -1,
			brightness: 50,
			wantBright: true,
		},
		{
			name:   "on set",
			red:    -1,
			green:  -1,
			blue:   -1,
			on:     true,
			wantOn: true,
		},
		{
			name: "out of range values ignored",
			red:  256,
		},
		{
			name:  "zero is valid",
			red:   0,
			green: 0,
			blue:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := buildParams(tt.red, tt.green, tt.blue, tt.brightness, tt.on)

			checkPtr := func(name string, got *int, want bool) {
				if want && got == nil {
					t.Errorf("%s: expected value, got nil", name)
				} else if !want && got != nil {
					t.Errorf("%s: expected nil, got %v", name, *got)
				}
			}

			if tt.wantRed {
				checkPtr("red", params.Red, true)
			}
			if tt.wantGreen {
				checkPtr("green", params.Green, true)
			}
			if tt.wantBlue {
				checkPtr("blue", params.Blue, true)
			}
			if tt.wantBright {
				checkPtr("brightness", params.Brightness, true)
			}
			if tt.wantOn && params.On == nil {
				t.Error("on: expected value, got nil")
			}
		})
	}
}

func TestBuildParams_ValidRanges(t *testing.T) {
	t.Parallel()

	// Test valid color range (0-255)
	params := buildParams(0, 128, 255, 50, true)
	if params.Red == nil || *params.Red != 0 {
		t.Error("Red 0 should be valid")
	}
	if params.Green == nil || *params.Green != 128 {
		t.Error("Green 128 should be valid")
	}
	if params.Blue == nil || *params.Blue != 255 {
		t.Error("Blue 255 should be valid")
	}
	if params.Brightness == nil || *params.Brightness != 50 {
		t.Error("Brightness 50 should be valid")
	}
	if params.On == nil || !*params.On {
		t.Error("On true should be valid")
	}
}

func TestBuildParams_InvalidRanges(t *testing.T) {
	t.Parallel()

	// Test invalid color range
	params := buildParams(256, -1, 300, 101, false)
	if params.Red != nil {
		t.Error("Red 256 should be invalid")
	}
	if params.Green != nil {
		t.Error("Green -1 should be invalid")
	}
	if params.Blue != nil {
		t.Error("Blue 300 should be invalid")
	}
	if params.Brightness != nil {
		t.Error("Brightness 101 should be invalid")
	}
	if params.On != nil {
		t.Error("On false should not set the pointer")
	}
}

// Verify RGBSetParams type is compatible.
var _ shelly.RGBSetParams = buildParams(0, 0, 0, 0, false)
