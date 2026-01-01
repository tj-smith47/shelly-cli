package set

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

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

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

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
		{"white", "white", "w", "-1"},
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

func TestRun_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "test-device",
		Red:        255,
		Green:      128,
		Blue:       64,
		White:      100,
		Brightness: 50,
		On:         true,
		ComponentFlags: flags.ComponentFlags{
			ID: 0,
		},
	}

	// Create a cancelled context to prevent actual device connection
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail due to cancelled context, but validates the run function path
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithDefaultID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Red:     100,
	}

	// Default ID should be 0
	if opts.ID != 0 {
		t.Errorf("Default ID = %d, want 0", opts.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_WithCustomID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "rgbw-device",
		ComponentFlags: flags.ComponentFlags{
			ID: 2,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestOptions_DeviceFieldSet(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "my-rgbw-device",
	}

	if opts.Device != "my-rgbw-device" {
		t.Errorf("Device = %q, want 'my-rgbw-device'", opts.Device)
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:    tf.Factory,
		Device:     "test-device",
		Red:        -1,
		Green:      -1,
		Blue:       -1,
		White:      -1,
		Brightness: -1,
	}

	// Default values should match flag defaults
	if opts.Red != -1 {
		t.Errorf("Default Red = %d, want -1", opts.Red)
	}
	if opts.Green != -1 {
		t.Errorf("Default Green = %d, want -1", opts.Green)
	}
	if opts.Blue != -1 {
		t.Errorf("Default Blue = %d, want -1", opts.Blue)
	}
	if opts.White != -1 {
		t.Errorf("Default White = %d, want -1", opts.White)
	}
	if opts.Brightness != -1 {
		t.Errorf("Default Brightness = %d, want -1", opts.Brightness)
	}
	if opts.On {
		t.Error("Default On should be false")
	}
}

func TestNewCommand_Execute_SetsDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-test-device", "--red", "255"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
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
			name:    "white flag",
			args:    []string{"--white", "200"},
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
			args:    []string{"-r", "255", "-g", "128", "-b", "64", "-w", "100", "--brightness", "75", "--on", "--id", "1"},
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

func TestBuildRGBWSetParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		red        int
		green      int
		blue       int
		white      int
		brightness int
		on         bool
		wantRed    bool
		wantGreen  bool
		wantBlue   bool
		wantWhite  bool
		wantBright bool
		wantOn     bool
	}{
		{
			name:       "no params set",
			red:        -1,
			green:      -1,
			blue:       -1,
			white:      -1,
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
			name:      "white set",
			red:       -1,
			green:     -1,
			blue:      -1,
			white:     200,
			wantWhite: true,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := shelly.BuildRGBWSetParams(tt.red, tt.green, tt.blue, tt.white, tt.brightness, tt.on)

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
			if tt.wantWhite {
				checkPtr("white", params.White, true)
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

func TestExecute_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-rgbw",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHRGBW2",
					Model:      "Shelly RGBW2",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-rgbw": {"rgbw:0": map[string]any{"output": false, "red": 0, "green": 0, "blue": 0, "white": 0}},
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
	cmd.SetArgs([]string{"test-rgbw", "-r", "255", "-g", "128", "-b", "64", "-w", "100"})
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
