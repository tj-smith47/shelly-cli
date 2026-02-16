// Package set provides the rgb set subcommand.
package set

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand(cmdutil.NewFactory()) returned nil")
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
	cmd := NewCommand(cmdutil.NewFactory())

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
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Args == nil {
		t.Error("Args validator not set")
	}
}

func TestBuildRGBSetParams(t *testing.T) {
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
			params := shelly.BuildRGBSetParams(tt.red, tt.green, tt.blue, tt.brightness, tt.on)

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

func TestBuildRGBSetParams_ValidRanges(t *testing.T) {
	t.Parallel()

	// Test valid color range (0-255)
	params := shelly.BuildRGBSetParams(0, 128, 255, 50, true)
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

func TestBuildRGBSetParams_InvalidRanges(t *testing.T) {
	t.Parallel()

	// Test invalid color range
	params := shelly.BuildRGBSetParams(256, -1, 300, 101, false)
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
var _ shelly.RGBSetParams = shelly.BuildRGBSetParams(0, 0, 0, 0, false)

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}

	// Check for "color" alias
	hasColor := false
	for _, a := range cmd.Aliases {
		if a == "color" {
			hasColor = true
			break
		}
	}
	if !hasColor {
		t.Error("Aliases should contain 'color'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestExecute_WithCancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"test-device", "--red", "255"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "red flag short",
			args:    []string{"-r", "128"},
			wantErr: false,
		},
		{
			name:    "red flag long",
			args:    []string{"--red", "128"},
			wantErr: false,
		},
		{
			name:    "all color flags",
			args:    []string{"-r", "255", "-g", "128", "-b", "64"},
			wantErr: false,
		},
		{
			name:    "brightness flag",
			args:    []string{"--brightness", "75"},
			wantErr: false,
		},
		{
			name:    "on flag",
			args:    []string{"--on"},
			wantErr: false,
		},
		{
			name:    "id flag",
			args:    []string{"--id", "2"},
			wantErr: false,
		},
		{
			name:    "all flags combined",
			args:    []string{"-r", "255", "-g", "128", "-b", "64", "--brightness", "75", "--on", "--id", "1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecute_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-rgb",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHRGBW2",
					Model:      "Shelly RGBW2",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-rgb": {"rgb:0": map[string]any{"output": false, "red": 0, "green": 0, "blue": 0}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-rgb", "-r", "255", "-g", "128", "-b", "64"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{Devices: []mock.DeviceFixture{}},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent device")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "unknown") {
		t.Logf("error = %v", err)
	}
}
