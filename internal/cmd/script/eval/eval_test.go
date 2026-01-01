package eval

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "eval <device> <id> <code>" {
		t.Errorf("Use = %q, want \"eval <device> <id> <code>\"", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"exec"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, expected := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias %q", expected)
			continue
		}
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"device"},
			wantErr: true,
		},
		{
			name:    "two args",
			args:    []string{"device", "1"},
			wantErr: true,
		},
		{
			name:    "three args valid (minimum)",
			args:    []string{"device", "1", "code"},
			wantErr: false,
		},
		{
			name:    "four args valid (code with spaces)",
			args:    []string{"device", "1", "print", "'hello'"},
			wantErr: false,
		},
		{
			name:    "five args valid (longer code)",
			args:    []string{"device", "1", "1", "+", "2"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction is nil, expected completion function")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_RunE_InvalidScriptID(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	cmd.SetArgs([]string{"device", "not-a-number", "1 + 2"})

	// Execute should fail because script ID is not a number
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid script ID, got nil")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	if len(cmd.Long) < 30 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestNewCommand_ArgsMinimum(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Eval uses MinimumNArgs(3), meaning at least 3 args required
	// This allows code to be multiple words joined together

	// Two args should fail
	err := cmd.Args(cmd, []string{"device", "1"})
	if err == nil {
		t.Error("expected error with only 2 args")
	}

	// Three args should succeed
	err = cmd.Args(cmd, []string{"device", "1", "code"})
	if err != nil {
		t.Errorf("unexpected error with 3 args: %v", err)
	}

	// Many args should succeed (code can be multiple words)
	err = cmd.Args(cmd, []string{"device", "1", "print", "'hello", "world'"})
	if err != nil {
		t.Errorf("unexpected error with 5 args: %v", err)
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

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_SuccessfulEval(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "eval-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"eval-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "eval-device",
		ID:      1,
		Code:    "1 + 2",
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{}, // No devices
		},
		DeviceStates: map[string]mock.DeviceState{},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
		ID:      1,
		Code:    "test()",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestNewCommand_Execute(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-eval-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"exec-eval-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"exec-eval-device", "1", "print('hello')"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestNewCommand_ExecuteMultiWordCode(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "multiword-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"multiword-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	// Code as multiple arguments that get joined
	cmd.SetArgs([]string{"multiword-device", "1", "1", "+", "2", "+", "3"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly script eval",
		"living-room",
		"myFunction()",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}
