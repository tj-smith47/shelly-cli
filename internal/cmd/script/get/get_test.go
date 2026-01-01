package get

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

	if cmd.Use != "get <device> <id>" {
		t.Errorf("Use = %q, want \"get <device> <id>\"", cmd.Use)
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

	expectedAliases := []string{"code"}
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
			name:    "two args valid",
			args:    []string{"device", "1"},
			wantErr: false,
		},
		{
			name:    "three args invalid",
			args:    []string{"device", "1", "extra"},
			wantErr: true,
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag == nil {
		t.Fatal("status flag not found")
	}

	if statusFlag.DefValue != "false" {
		t.Errorf("status default = %q, want \"false\"", statusFlag.DefValue)
	}

	// status flag should not have a shorthand based on the source
	if statusFlag.Shorthand != "" {
		t.Errorf("status shorthand = %q, want empty", statusFlag.Shorthand)
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
	cmd.SetArgs([]string{"device", "not-a-number"})

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

func TestNewCommand_StatusFlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag == nil {
		t.Fatal("status flag not found")
	}

	// Verify the flag has a meaningful usage description
	if statusFlag.Usage == "" {
		t.Error("status flag has no usage description")
	}

	if len(statusFlag.Usage) < 10 {
		t.Error("status flag usage seems too short to be meaningful")
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
func TestRun_GetCode(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "get-code-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"get-code-device": {},
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
		Factory:    tf.Factory,
		Device:     "get-code-device",
		ID:         1,
		ShowStatus: false,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_GetStatus(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "get-status-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"get-status-device": {},
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
		Factory:    tf.Factory,
		Device:     "get-status-device",
		ID:         1,
		ShowStatus: true,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Status output should contain memory info
	if !strings.Contains(output, "ID") && !strings.Contains(output, "Running") {
		t.Errorf("output should contain status info, got: %q", output)
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
					Name:       "exec-get-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"exec-get-device": {},
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
	cmd.SetArgs([]string{"exec-get-device", "1"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestNewCommand_ExecuteWithStatus(t *testing.T) {
	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-status-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"exec-status-device": {},
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
	cmd.SetArgs([]string{"exec-status-device", "1", "--status"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly script get",
		"--status",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}
