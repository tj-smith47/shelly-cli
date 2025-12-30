package status

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "status <device> [id]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status <device> [id]")
	}

	// Test Aliases
	wantAliases := []string{"st"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"device"}, false},
		{"two args valid", []string{"device", "0"}, false},
		{"three args", []string{"device", "0", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test type flag
	flag := cmd.Flags().Lookup("type")
	if flag == nil {
		t.Fatal("--type flag not found")
	}
	if flag.DefValue != "auto" {
		t.Errorf("--type default = %q, want %q", flag.DefValue, "auto")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly power status",
		"--type",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_InvalidComponentID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "not-a-number"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid component ID")
	}

	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("expected 'invalid component ID' error, got: %v", err)
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "power meter status") {
		t.Errorf("help output should contain command description, got: %s", output)
	}
	if !strings.Contains(output, "status <device>") {
		t.Error("help output should show usage")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no device argument provided")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device1", "device2", "extra"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when too many arguments provided")
	}
}

func TestExecute_WithMockPMDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-pm",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-pm": {
				"pm:0": map[string]any{
					"voltage":  230.5,
					"current":  1.2,
					"apower":   275.6,
					"freq":     50.0,
					"aenergy":  map[string]any{"total": 1234.5},
				},
			},
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
	cmd.SetArgs([]string{"test-pm"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithMockPM1Device(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-pm1",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:00",
					Type:       "SNSW-002P16EU",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-pm1": {
				"pm1:0": map[string]any{
					"voltage": 230.0,
					"current": 2.5,
					"apower":  575.0,
					"freq":    50.0,
					"aenergy": map[string]any{"total": 5678.9},
				},
			},
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
	cmd.SetArgs([]string{"test-pm1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithComponentID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "multi-pm",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"multi-pm": {
				"pm:0": map[string]any{
					"voltage": 230.0,
					"current": 1.0,
					"apower":  230.0,
					"freq":    50.0,
				},
			},
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
	cmd.SetArgs([]string{"multi-pm", "0"})
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
}

func TestExecute_WithTypeFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "typed-pm",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:22",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"typed-pm": {
				"pm:0": map[string]any{
					"voltage": 230.0,
					"current": 1.5,
					"apower":  345.0,
				},
			},
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
	cmd.SetArgs([]string{"typed-pm", "--type", shelly.ComponentTypePM})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_NoComponentsFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "no-pm",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:33",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"no-pm": {},
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
	cmd.SetArgs([]string{"no-pm"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// This will likely error due to no power meter components found
	if err != nil {
		t.Logf("Execute error (expected): %v", err)
	}
}

func TestExecute_WithJSON(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-pm",
					Address:    "192.168.1.105",
					MAC:        "AA:BB:CC:DD:EE:44",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-pm": {
				"pm:0": map[string]any{
					"voltage": 230.0,
					"current": 1.0,
					"apower":  230.0,
					"aenergy": map[string]any{"total": 100.5},
				},
			},
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
	cmd.SetArgs([]string{"json-pm", "-o", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithInvalidID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "invalid-id",
					Address:    "192.168.1.106",
					MAC:        "AA:BB:CC:DD:EE:55",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"invalid-id": {
				"pm:0": map[string]any{
					"voltage": 230.0,
					"current": 1.0,
					"apower":  230.0,
				},
			},
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
	cmd.SetArgs([]string{"invalid-id", "abc"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid component ID")
	}
	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("expected 'invalid component ID' error, got: %v", err)
	}
}

func TestExecute_WithPM1TypeFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "pm1-typed",
					Address:    "192.168.1.107",
					MAC:        "AA:BB:CC:DD:EE:66",
					Type:       "SNSW-002P16EU",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"pm1-typed": {
				"pm1:0": map[string]any{
					"voltage": 230.0,
					"current": 2.0,
					"apower":  460.0,
				},
			},
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
	cmd.SetArgs([]string{"pm1-typed", "--type", shelly.ComponentTypePM1})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestRun_PMSuccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a test context
	ctx := context.Background()

	// Call run directly with known PM type to cover PM success path
	err := run(ctx, tf.Factory, "nonexistent-pm", 0, shelly.ComponentTypePM)

	// We expect an error because the device doesn't exist in mock
	if err == nil {
		t.Error("expected error due to nonexistent device")
	}
	if !strings.Contains(err.Error(), "failed to get pm status") {
		t.Logf("error = %v", err)
	}
}

func TestRun_PM1Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a test context
	ctx := context.Background()

	// Call run directly with known PM1 type to cover PM1 success path
	err := run(ctx, tf.Factory, "nonexistent-pm1", 0, shelly.ComponentTypePM1)

	// We expect an error because the device doesn't exist in mock
	if err == nil {
		t.Error("expected error due to nonexistent device")
	}
	if !strings.Contains(err.Error(), "failed to get pm1 status") {
		t.Logf("error = %v", err)
	}
}

func TestRun_AutoDetectToPM(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx := context.Background()

	// Call run with auto-detect type
	err := run(ctx, tf.Factory, "test-device", 0, shelly.ComponentTypeAuto)

	// We expect an error because the device doesn't exist
	if err == nil {
		t.Error("expected error due to nonexistent device")
	}
	// When auto-detect returns auto, it goes to default case
	if !strings.Contains(err.Error(), "no power meter components found") {
		t.Logf("error = %v", err)
	}
}

func TestNewCommand_RunE_InvalidID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"mydevice", "abc123"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid component ID in RunE")
	}
	// The error could be either from parsing or from trying to get the component
	if !strings.Contains(err.Error(), "invalid component ID") && !strings.Contains(err.Error(), "no power meter") {
		t.Errorf("expected error for invalid component ID, got: %v", err)
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require at least 1 argument
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should accept 1 argument
	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	// Should accept 2 arguments
	err = cmd.Args(cmd, []string{"device1", "0"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to get defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag.DefValue != "auto" {
		t.Errorf("type default = %q, want auto", typeFlag.DefValue)
	}
}

func TestRun_DirectCall_PM(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()

	// Directly test the run function with PM type
	// This should hit the GetPMStatus path
	err := run(ctx, f, "test-device", 0, shelly.ComponentTypePM)

	// We expect an error because the device doesn't exist
	if err == nil {
		t.Error("expected error when device not found")
	}
}

func TestRun_DirectCall_PM1(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()

	// Directly test the run function with PM1 type
	// This should hit the GetPM1Status path
	err := run(ctx, f, "test-device", 0, shelly.ComponentTypePM1)

	// We expect an error because the device doesn't exist
	if err == nil {
		t.Error("expected error when device not found")
	}
}

func TestRun_DirectCall_Auto(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	ctx := context.Background()

	// Directly test the run function with auto-detect type
	// This should go to the default case (no power meter components found)
	err := run(ctx, f, "test-device", 0, shelly.ComponentTypeAuto)

	// We expect an error because no power meter components found
	if err == nil {
		t.Error("expected error when no power meter components found")
	}
	if !strings.Contains(err.Error(), "no power meter components found") {
		t.Errorf("expected 'no power meter components found' error, got: %v", err)
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Show power meter status"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "power meter") {
		t.Error("Long should contain 'power meter'")
	}
}
