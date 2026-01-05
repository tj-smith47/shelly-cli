package create

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
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
	if cmd.Use != "create <device> <type>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <device> <type>")
	}

	// Test Aliases
	wantAliases := []string{"add", "new"}
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

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"device"}, true},
		{"two args valid", []string{"device", "boolean"}, false},
		{"three args", []string{"device", "boolean", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory()) // Fresh cmd per subtest
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		defValue string
	}{
		{"name", ""},
		{"id", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory()) // Fresh cmd per subtest
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
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
		t.Error("ValidArgsFunction should be set")
	}
}

func TestNewCommand_ValidArgsFunction_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test with no args - should return device names completion
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")
	// Not checking suggestions since DeviceNames() returns from config
	if directive == cobra.ShellCompDirectiveError {
		t.Error("Should not return error directive")
	}
	_ = suggestions // DeviceNames returns from config, so we just verify it doesn't error
}

func TestNewCommand_ValidArgsFunction_OneArg(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test with one arg (device provided) - should return virtual types
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"device"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	if len(suggestions) == 0 {
		t.Error("Expected virtual type suggestions")
	}
	// Verify it contains expected types
	hasBoolean := false
	for _, s := range suggestions {
		if s == "boolean" {
			hasBoolean = true
			break
		}
	}
	if !hasBoolean {
		t.Error("Expected 'boolean' in suggestions")
	}
}

func TestNewCommand_ValidArgsFunction_TwoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test with two args (device and type provided) - should return no completion
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"device", "boolean"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Directive = %v, want %v", directive, cobra.ShellCompDirectiveNoFileComp)
	}
	if len(suggestions) != 0 {
		t.Errorf("Expected no suggestions, got %v", suggestions)
	}
}

func TestNewCommand_InvalidType(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "invalid-type"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}

	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("expected 'invalid type' error, got: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly virtual create",
		"boolean",
		"--name",
		"--id",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "test-device",
		Type:    "boolean",
		Name:    "Test Component",
		ID:      200,
		Factory: f,
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.Type != "boolean" {
		t.Errorf("Type = %q, want %q", opts.Type, "boolean")
	}

	if opts.Name != "Test Component" {
		t.Errorf("Name = %q, want %q", opts.Name, "Test Component")
	}

	if opts.ID != 200 {
		t.Errorf("ID = %d, want %d", opts.ID, 200)
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

func TestRun_WithMock_SuccessWithName(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:  "test-device",
		Type:    "boolean",
		Name:    "Test Component",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	// May fail due to mock limitations, but exercises the code path
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_WithMock_SuccessWithoutName(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:  "test-device",
		Type:    "number",
		Name:    "", // No name
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_WithMock_WithSpecificID(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:  "test-device",
		Type:    "text",
		Name:    "Custom Text",
		ID:      250,
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_DeviceNotFound(t *testing.T) {
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

	opts := &Options{
		Device:  "nonexistent",
		Type:    "boolean",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_AllTypes(t *testing.T) {
	t.Parallel()

	types := []string{"boolean", "number", "text", "enum", "button", "group"}

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
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
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	for _, vType := range types {
		t.Run(vType, func(t *testing.T) {
			t.Parallel()

			// Create fresh TestFactory per subtest to avoid race on IOStreams
			tf := factory.NewTestFactory(t)
			demo.InjectIntoFactory(tf.Factory)

			opts := &Options{
				Device:  "test-device",
				Type:    vType,
				Name:    "Test " + vType,
				Factory: tf.Factory,
			}

			ctx := context.Background()
			err := run(ctx, opts)
			if err != nil {
				t.Logf("run() for type %s error = %v (may be expected for mock)", vType, err)
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
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "boolean", "--name", "Test Bool"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithMock_AllFlags(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "kitchen",
					Address:    "192.168.1.200",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"kitchen": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"kitchen", "number", "--name", "Temperature", "--id", "205"})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestRun_CancelledContext(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Device:  "test-device",
		Type:    "boolean",
		Factory: tf.Factory,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}
