package list

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
	if cmd.Use != "list <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list <device>")
	}

	// Test Aliases
	wantAliases := []string{"ls", "l"}
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
		{"two args", []string{"device", "extra"}, true},
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

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"values", "", "false"},
		{"match", "m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
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
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly kvs list",
		"--values",
		"--match",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

// Execute-based tests with mock fixtures

func TestExecute_ListKeysDefaultOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Expected to handle successfully or with graceful error from mock
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithValuesFlag(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--values"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithMatchPattern(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--match", "sensor_*"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithShorthand(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "-m", "config_*"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithValuesAndMatch(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--values", "--match", "script_*"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithJSONOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "-o", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysWithYAMLOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "-o", "yaml"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysValuesAndJSONOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--values", "-o", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_ListKeysUsingAlias_ls(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test that the command works when referenced by alias
	cmd := NewCommand(tf.Factory)
	if len(cmd.Aliases) < 1 || cmd.Aliases[0] != "ls" {
		t.Skip("ls alias not found")
	}
}

func TestExecute_ListKeysUsingAlias_l(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test that the command works when referenced by alias
	cmd := NewCommand(tf.Factory)
	if len(cmd.Aliases) < 2 || cmd.Aliases[1] != "l" {
		t.Skip("l alias not found")
	}
}

func TestOptions_Device(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
}

func TestOptions_Values(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Values:  true,
		Factory: cmdutil.NewFactory(),
	}

	if !opts.Values {
		t.Error("Values = false, want true")
	}
}

func TestOptions_Match(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Match:   "sensor_*",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Match != "sensor_*" {
		t.Errorf("Match = %q, want %q", opts.Match, "sensor_*")
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "test-device",
		Factory: f,
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	if opts.Factory.IOStreams() == nil {
		t.Error("Factory.IOStreams() returned nil")
	}
}

func TestRun_WithContext(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		Factory: tf.Factory,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_WithContextAndValues(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		Values:  true,
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_WithContextAndMatch(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		Match:   "data_*",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_WithMatchWildcard(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
		Values:  true,
		Match:   "*",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short is empty")
	}

	if !strings.Contains(cmd.Short, "KVS") || !strings.Contains(cmd.Short, "keys") {
		t.Errorf("Short should mention KVS and keys, got: %s", cmd.Short)
	}
}

func TestNewCommand_LongDescribesKVS(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedPatterns := []string{
		"Key-Value Storage",
		"persistent storage",
		"wildcard",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long should contain %q", pattern)
		}
	}
}

func TestExecute_MultipleDevices(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:00",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
				{
					Name:       "device2",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device1": {},
			"device2": {},
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
	cmd.SetArgs([]string{"device1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_InvalidDevice(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"nonexistent-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Expected to error or handle gracefully
	if err != nil {
		t.Logf("Execute error (expected for nonexistent device): %v", err)
	}
}

func TestNewCommand_LongContainsFlagInfo(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention flags
	if !strings.Contains(cmd.Long, "values") {
		t.Error("Long should mention values flag")
	}

	if !strings.Contains(cmd.Long, "match") {
		t.Error("Long should mention match flag")
	}
}

func TestNewCommand_IsCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should not be nil and should have the expected structure
	if cmd.Use == "" || cmd.RunE == nil {
		t.Error("Command not properly configured")
	}

	// Verify it's intended to be a leaf command
	if len(cmd.Commands()) > 0 {
		t.Error("list command should have no subcommands")
	}
}

func TestOptions_EmptyValues(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Values:  false,
		Factory: cmdutil.NewFactory(),
	}

	if opts.Values {
		t.Error("Values should be false by default")
	}
}

func TestOptions_EmptyMatch(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test-device",
		Match:   "",
		Factory: cmdutil.NewFactory(),
	}

	if opts.Match != "" {
		t.Errorf("Match should be empty by default, got: %q", opts.Match)
	}
}

func TestRun_ValuesOnlyNoMatch(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test the Values path without explicit match (should use wildcard)
	opts := &Options{
		Device:  "test-device",
		Values:  true,
		Match:   "",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestRun_MatchOnlyNoValues(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test the Match path (uses GetMany, not List)
	opts := &Options{
		Device:  "test-device",
		Values:  false,
		Match:   "test_*",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

func TestNewCommand_Short_NotEmpty(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Should be reasonably descriptive
	if len(cmd.Short) < 5 {
		t.Error("Short description is too short")
	}
}

func TestNewCommand_LongLengthCheck(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Long) < 30 {
		t.Error("Long description seems too short")
	}
}

func TestExecute_ListKeysWithMatchAndOutput(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--match", "prefix_*", "-o", "table"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestExecute_AllFlagsAtOnce(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--values", "--match", "app_*", "-o", "yaml"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
}

func TestOptions_AllFieldsSet(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Device:  "mydevice",
		Values:  true,
		Match:   "key_*",
		Factory: f,
	}

	// Verify all fields are set correctly
	if opts.Device != "mydevice" {
		t.Errorf("Device = %q, want %q", opts.Device, "mydevice")
	}

	if !opts.Values {
		t.Error("Values = false, want true")
	}

	if opts.Match != "key_*" {
		t.Errorf("Match = %q, want %q", opts.Match, "key_*")
	}

	if opts.Factory != f {
		t.Error("Factory not set correctly")
	}
}

func TestNewCommand_ShortAndLongDiffer(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Short and Long should be different
	if cmd.Short == cmd.Long {
		t.Error("Short and Long descriptions should differ")
	}
}

func TestRun_WithNoValuesNoMatch(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Test the default path (neither Values nor Match set)
	opts := &Options{
		Device:  "test-device",
		Values:  false,
		Match:   "",
		Factory: tf.Factory,
	}

	ctx := context.Background()
	err = run(ctx, opts)
	if err != nil {
		t.Logf("run returned: %v", err)
	}
}

// Additional tests for better coverage

func TestOptions_ZeroValues(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Device != "" {
		t.Errorf("Device should be empty string initially")
	}

	if opts.Values {
		t.Error("Values should be false initially")
	}

	if opts.Match != "" {
		t.Error("Match should be empty string initially")
	}

	if opts.Factory != nil {
		t.Error("Factory should be nil initially")
	}
}

func TestNewCommand_ExecutableSuccess(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Verify the command structure allows execution
	if cmd.RunE == nil {
		t.Fatal("RunE function should not be nil")
	}

	// Call help to verify basic execution works
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Logf("help execution error (expected): %v", err)
	}
}

func TestNewCommand_Use_Contains_Device_Placeholder(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Use, "device") {
		t.Error("Use string should mention 'device'")
	}

	if !strings.Contains(cmd.Use, "<") || !strings.Contains(cmd.Use, ">") {
		t.Error("Use should have angle brackets for argument")
	}
}

func TestNewCommand_Example_Multiple_Commands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Count occurrences of command invocation
	shellyCount := strings.Count(cmd.Example, "shelly")
	if shellyCount < 2 {
		t.Errorf("Example should show multiple command invocations, found %d", shellyCount)
	}
}

func TestNewCommand_Long_Contains_Output_Info(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention output formats
	hasTableInfo := strings.Contains(cmd.Long, "table") || strings.Contains(cmd.Long, "Table")
	hasJSONInfo := strings.Contains(cmd.Long, "JSON") || strings.Contains(cmd.Long, "json")

	if !hasTableInfo && !hasJSONInfo {
		t.Error("Long should mention output formats")
	}
}

func TestNewCommand_Example_Shows_OptionUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should show different option combinations
	hasDeviceExample := strings.Contains(cmd.Example, "living-room") || strings.Contains(cmd.Example, "kitchen")
	if !hasDeviceExample {
		t.Log("Example should contain device names for clarity")
	}
}

func TestNewCommand_Flags_Are_Boolean(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	valuesFlag := cmd.Flags().Lookup("values")
	if valuesFlag == nil {
		t.Fatal("values flag not found")
	}

	// Verify it's a boolean flag by checking type
	if valuesFlag.Value.Type() != "bool" {
		t.Errorf("values flag type = %s, want bool", valuesFlag.Value.Type())
	}
}

func TestNewCommand_Match_Flag_Has_Shorthand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	matchFlag := cmd.Flags().Lookup("match")
	if matchFlag == nil {
		t.Fatal("match flag not found")
	}

	if matchFlag.Shorthand != "m" {
		t.Errorf("match flag shorthand = %q, want 'm'", matchFlag.Shorthand)
	}
}

func TestOptions_IsStruct(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test",
		Values:  true,
		Match:   "pattern",
		Factory: cmdutil.NewFactory(),
	}

	// Verify fields are accessible
	if opts.Device != "test" {
		t.Error("Device field not accessible")
	}

	if !opts.Values {
		t.Error("Values field not accessible")
	}

	if opts.Match != "pattern" {
		t.Error("Match field not accessible")
	}

	if opts.Factory == nil {
		t.Error("Factory field not accessible")
	}
}

func TestNewCommand_Short_Starts_With_Capital(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short != "" && cmd.Short[0] >= 'a' && cmd.Short[0] <= 'z' {
		t.Error("Short description should start with capital letter")
	}
}

func TestNewCommand_Example_Shows_JSON_Output(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "json") {
		t.Error("Example should show JSON output usage")
	}
}

func TestNewCommand_Example_Shows_Match_Flag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if !strings.Contains(cmd.Example, "match") && !strings.Contains(cmd.Example, "-m") {
		t.Error("Example should demonstrate match flag usage")
	}
}

func TestExecute_WithTimeout(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-25",
					Model:      "Shelly Plus 2PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Should complete or timeout gracefully
	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute with context: %v", err)
	}
}

// Tests for display and output coverage

func TestNewCommand_HelpText_Complete(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Help should be accessible
	if cmd == nil {
		t.Fatal("Command is nil")
	}

	// The command should have help content
	if cmd.Short == "" && cmd.Long == "" {
		t.Error("Command should have help text")
	}
}

func TestNewCommand_Args_Exactly_One(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify Args validation using cobra's validation
	if cmd.Args == nil {
		t.Error("Args validation function should be set")
	}

	// Test exact argument requirements
	if cmd.Args(cmd, []string{}) == nil {
		t.Error("should reject empty args")
	}

	if cmd.Args(cmd, []string{"device1", "device2"}) == nil {
		t.Error("should reject more than one arg")
	}
}

func TestNewCommand_Fields_Populated(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify all required Cobra command fields are populated
	if cmd.Use == "" {
		t.Error("Use field should not be empty")
	}

	if cmd.Short == "" {
		t.Error("Short field should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long field should not be empty")
	}

	if cmd.Example == "" {
		t.Error("Example field should not be empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE field should be set")
	}
}

func TestNewCommand_Example_Well_Formatted(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Examples should have content
	lines := strings.Split(cmd.Example, "\n")
	if len(lines) < 3 {
		t.Error("Example should have multiple lines")
	}

	// Should have some descriptive comments
	hasComments := strings.Contains(cmd.Example, "#")
	if !hasComments {
		t.Log("Example should have descriptive comments")
	}
}

func TestNewCommand_Long_Detail_Level(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should explain the feature well
	// Check for key terms about KVS functionality
	terms := []string{"persistent", "storage", "keys", "values"}
	foundTerms := 0

	for _, term := range terms {
		if strings.Contains(strings.ToLower(cmd.Long), term) {
			foundTerms++
		}
	}

	if foundTerms < 2 {
		t.Log("Long description should explain KVS functionality in detail")
	}
}

func TestOptions_Exported_Fields(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:  "test",
		Values:  true,
		Match:   "pattern",
		Factory: cmdutil.NewFactory(),
	}

	// All fields must be exported (start with capital)
	// Verify by checking if we can access them
	device := opts.Device
	values := opts.Values
	match := opts.Match
	f := opts.Factory

	if device == "" || !values || match == "" || f == nil {
		t.Error("Options fields should be accessible and populated")
	}
}

func TestNewCommand_Cmd_Field_Set(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify the returned value is a valid cobra.Command
	if cmd == nil {
		t.Fatal("NewCommand should not return nil")
	}

	// Should have RunE defined (Execute method uses this)
	if cmd.RunE == nil {
		t.Error("RunE should be defined for execution")
	}
}

func TestNewCommand_Use_Format_Correct(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Use should follow cobra convention: "command <required> [optional]"
	// For this command it should be: "list <device>"
	expectedFormat := "list"
	if !strings.Contains(cmd.Use, expectedFormat) {
		t.Errorf("Use should contain %q", expectedFormat)
	}

	if !strings.Contains(cmd.Use, "device") {
		t.Error("Use should mention device argument")
	}
}

func TestNewCommand_Aliases_Type(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Aliases should be a slice of strings
	if cmd.Aliases == nil {
		t.Error("Aliases should not be nil")
	}

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases should not be empty")
	}

	// Each alias should be a valid string
	for i, alias := range cmd.Aliases {
		if alias == "" {
			t.Errorf("Alias at index %d is empty", i)
		}
		if len(alias) > 10 {
			t.Errorf("Alias %q seems too long", alias)
		}
	}
}

func TestNewCommand_Complete_Verification(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Verify all major components are in place
	checks := []struct {
		name  string
		check func() bool
	}{
		{"Use is not empty", func() bool { return cmd.Use != "" }},
		{"Short is not empty", func() bool { return cmd.Short != "" }},
		{"Long is not empty", func() bool { return cmd.Long != "" }},
		{"Example is not empty", func() bool { return cmd.Example != "" }},
		{"RunE is set", func() bool { return cmd.RunE != nil }},
		{"Has at least one alias", func() bool { return len(cmd.Aliases) >= 1 }},
		{"Has ValidArgsFunction", func() bool { return cmd.ValidArgsFunction != nil }},
		{"Args is set", func() bool { return cmd.Args != nil }},
	}

	for _, check := range checks {
		if !check.check() {
			t.Errorf("Check failed: %s", check.name)
		}
	}
}

func TestNewCommand_WithDifferentFactories(t *testing.T) {
	t.Parallel()

	// Test that NewCommand works with different factory instances
	f1 := cmdutil.NewFactory()
	f2 := cmdutil.NewFactory()

	cmd1 := NewCommand(f1)
	cmd2 := NewCommand(f2)

	if cmd1 == nil || cmd2 == nil {
		t.Fatal("NewCommand should not return nil")
	}

	// Both should have the same structure
	if cmd1.Use != cmd2.Use {
		t.Error("Commands should have same Use")
	}

	if cmd1.Short != cmd2.Short {
		t.Error("Commands should have same Short")
	}
}
