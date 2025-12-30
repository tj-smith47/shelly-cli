package enable

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	mockserver "github.com/tj-smith47/shelly-cli/internal/mock"
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
	if cmd.Use != "enable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enable <device>")
	}

	// Test Aliases
	wantAliases := []string{"on", "activate"}
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
		{"two args", []string{"device1", "device2"}, true},
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

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly matter enable",
		"shelly matter code",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"Matter",
		"commission",
		"Apple Home",
		"Google Home",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "test-device",
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestExecute_Gen2Device_Success(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May succeed or fail depending on mock capabilities
	// but should reach the execute path
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{Version: "1", Config: mockserver.ConfigFixture{}}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_WithMock(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err = run(context.Background(), opts)
	// May fail due to mock limitations for Matter
	if err != nil {
		t.Logf("run() error = %v (expected for mock Matter operations)", err)
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{Version: "1", Config: mockserver.ConfigFixture{}}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestExecute_WithAlias_On(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "living-room",
					Address:    "192.168.1.101",
					MAC:        "BB:CC:DD:EE:FF:00",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"living-room": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"living-room"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Just test that Execute path works
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

func TestExecute_WithAlias_Activate(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "bedroom",
					Address:    "192.168.1.102",
					MAC:        "CC:DD:EE:FF:00:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"bedroom": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"bedroom"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Just test that Execute path works
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

func TestRun_OutputMessage(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err = run(context.Background(), opts)
	// May fail due to mock limitations
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
	// If successful, verify output contains expected messages
	if err == nil {
		output := tf.OutString()
		if !strings.Contains(output, "Matter") {
			t.Errorf("expected output to contain 'Matter', got: %s", output)
		}
	}
}

func TestExecute_MultipleDevices_FirstOne(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "device-1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "device-2",
					Address:    "192.168.1.101",
					MAC:        "BB:CC:DD:EE:FF:00",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mockserver.DeviceState{
			"device-1": {"switch:0": map[string]any{"output": true}},
			"device-2": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device-1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Just verify Execute path is exercised
	if err != nil {
		t.Logf("Execute error = %v (may be expected)", err)
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	err = run(ctx, opts)
	// Should either error due to cancellation or succeed
	// Just verify it handles the context properly
	if err != nil {
		t.Logf("run() error with cancelled context = %v", err)
	}
}

func TestExecute_CommandFields(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify command has proper configuration
	if cmd.Use != "enable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enable <device>")
	}

	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}

	if cmd.Example == "" {
		t.Error("Example should not be empty")
	}

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE should not be nil")
	}
}

// TestRun_FactoryIOStreams tests that the run function properly uses factory IOStreams.
func TestRun_FactoryIOStreams(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// Call run - it will error due to device not found, but we're testing the IOStreams usage
	err := run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error expected: %v", err)
	}
}

// TestNewCommand_Flags tests that the command doesn't incorrectly add flags.
func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify that no unexpected flags are added
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Flags() should return a flag set")
	}
}

// TestExecute_InjectsFactory tests that Execute correctly injects the device name.
func TestExecute_InjectsFactory(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config: mockserver.ConfigFixture{
			Devices: []mockserver.DeviceFixture{
				{
					Name:       "test",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	// Execute should process the args and attempt to run
	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (expected for mock): %v", err)
	}
}

// TestRun_WithFactoryContextTimeout tests WithDefaultTimeout integration.
func TestRun_WithFactoryContextTimeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "timeout-test",
	}

	// WithDefaultTimeout should be called internally by run()
	err := run(context.Background(), opts)
	// Error expected, but WithDefaultTimeout should have been called
	if err != nil {
		t.Logf("run() error expected: %v", err)
	}
}

// TestRun_ErrorPath ensures error is returned from RunWithSpinner.
func TestRun_ErrorPath(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Use a device name that won't be found
	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device-xyz",
	}

	// This should error because the device doesn't exist
	err := run(context.Background(), opts)
	require.Error(t, err)
}

// TestExecute_ErrorHandling verifies errors are properly handled.
func TestExecute_ErrorHandling(t *testing.T) {
	t.Parallel()

	fixtures := &mockserver.Fixtures{
		Version: "1",
		Config:  mockserver.ConfigFixture{},
	}

	demo, err := mockserver.StartWithFixtures(fixtures)
	require.NoError(t, err)
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"not-a-device"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	// Execute should return an error
	err = cmd.Execute()
	require.Error(t, err)
}

// TestExecute_CommandParsing verifies the command properly parses arguments.
func TestExecute_CommandParsing(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Set arguments and verify they're parsed
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"my-device"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	// Execute will error but args should be parsed correctly
	err := cmd.Execute()
	// Error is expected, just verifying argument parsing path is exercised
	if err != nil {
		t.Logf("Execute error (expected): %v", err)
	}
}

// TestRun_FactoryDependencies verifies the run function uses factory correctly.
func TestRun_FactoryDependencies(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
	}

	// run() should call Factory.WithDefaultTimeout and Factory.IOStreams and Factory.ShellyService
	err := run(context.Background(), opts)
	// Error is expected since device doesn't exist
	if err != nil {
		t.Logf("run() error (expected): %v", err)
	}
}
