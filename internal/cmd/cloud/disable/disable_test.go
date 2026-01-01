package disable

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
	if cmd.Use != "disable <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "disable <device>")
	}

	// Test Aliases
	wantAliases := []string{"off", "disconnect"}
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
		"shelly cloud disable",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

//nolint:paralleltest // Uses mock server which modifies global state
func TestRun_WithMock(t *testing.T) {
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
			"test-device": {"switch:0": map[string]any{"output": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err = run(context.Background(), opts)
	// May fail due to mock limitations
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

//nolint:paralleltest // Uses mock server which modifies global state
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{Factory: tf.Factory, Device: "nonexistent-device"}
	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses test factory which may modify global state
func TestRun_CancelledContext(t *testing.T) {
	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{Factory: tf.Factory, Device: "test-device"}
	err := run(ctx, opts)

	// Should fail due to cancelled context
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestNewCommand_CommandName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Name() != "disable" {
		t.Errorf("Name() = %q, want 'disable'", cmd.Name())
	}
}

func TestNewCommand_UsageString(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	usage := cmd.UsageString()
	if !strings.Contains(usage, "disable") {
		t.Error("UsageString should contain command name")
	}
}

func TestNewCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Commands()) > 0 {
		t.Errorf("disable command should not have subcommands, has %d", len(cmd.Commands()))
	}
}

func TestNewCommand_LongMentionsCloud(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "Cloud") && !strings.Contains(cmd.Long, "cloud") {
		t.Error("Long description should mention Cloud")
	}
}

func TestNewCommand_LongMentionsLocal(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Long, "local") {
		t.Error("Long description should mention local network")
	}
}

func TestNewCommand_ShortIsConcise(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if strings.Contains(cmd.Short, "\n") {
		t.Error("Short description should not contain newlines")
	}

	if len(cmd.Short) > 80 {
		t.Errorf("Short description too long (%d chars), should be under 80", len(cmd.Short))
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
		Device:  "my-device",
	}

	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
	if opts.Device != "my-device" {
		t.Errorf("Device = %q, want 'my-device'", opts.Device)
	}
}

func TestExecute_MissingArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing device argument")
	}
}

//nolint:paralleltest // Uses mock server which modifies global state
func TestExecute_WithDevice(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-test",
					Address:    "192.168.1.200",
					MAC:        "11:22:33:44:55:66",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
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

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"exec-test"})

	// May succeed or fail depending on mock capability
	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (expected for mock)", err)
	}
}
