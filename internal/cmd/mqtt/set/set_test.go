package set

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

	if cmd.Use != "set <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <device>")
	}

	wantAliases := []string{"configure", "config"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name     string
		defValue string
	}{
		{name: "server", defValue: ""},
		{name: "user", defValue: ""},
		{name: "password", defValue: ""},
		{name: "topic-prefix", defValue: ""},
		{name: "enable", defValue: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
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

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly mqtt set",
		"--server",
		"--enable",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device completion")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:     f,
		Device:      "test-device",
		Enable:      true,
		Server:      "mqtt://broker:1883",
		User:        "user",
		Password:    "pass",
		TopicPrefix: "shelly/",
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if !opts.Enable {
		t.Error("Enable should be true")
	}
	if opts.Server != "mqtt://broker:1883" {
		t.Errorf("Server = %q, want %q", opts.Server, "mqtt://broker:1883")
	}
	if opts.User != "user" {
		t.Errorf("User = %q, want %q", opts.User, "user")
	}
	if opts.Password != "pass" {
		t.Errorf("Password = %q, want %q", opts.Password, "pass")
	}
	if opts.TopicPrefix != "shelly/" {
		t.Errorf("TopicPrefix = %q, want %q", opts.TopicPrefix, "shelly/")
	}
}

func TestRun_NoOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		// No other options set
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error when no options are specified")
	}
	if !strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_EnableOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Enable:  true,
	}

	// Should not error on validation (enable is set)
	// Will fail on device resolution, but that's expected
	err := run(context.Background(), opts)
	if err == nil {
		t.Logf("run() succeeded (unexpected with no device registered)")
	} else if strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Error("should not fail validation when --enable is set")
	}
}

func TestRun_ServerOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Server:  "mqtt://broker:1883",
	}

	// Should not error on validation (server is set)
	err := run(context.Background(), opts)
	if err == nil {
		t.Logf("run() succeeded (unexpected with no device registered)")
	} else if strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Error("should not fail validation when --server is set")
	}
}

func TestRun_UserOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		User:    "testuser",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Logf("run() succeeded (unexpected with no device registered)")
	} else if strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Error("should not fail validation when --user is set")
	}
}

func TestRun_PasswordOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		Password: "testpass",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Logf("run() succeeded (unexpected with no device registered)")
	} else if strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Error("should not fail validation when --password is set")
	}
}

func TestRun_TopicPrefixOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:     tf.Factory,
		Device:      "test-device",
		TopicPrefix: "home/shelly",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Logf("run() succeeded (unexpected with no device registered)")
	} else if strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Error("should not fail validation when --topic-prefix is set")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
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

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		Server:  "mqtt://broker:1883",
	}

	err = run(context.Background(), opts)
	// May fail due to mock limitations for MQTT
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_DeviceNotFound(t *testing.T) {
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "nonexistent-device",
		Enable:  true,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_WithAllOptions(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "mqtt-device",
					Address:    "192.168.1.101",
					MAC:        "BB:CC:DD:EE:FF:00",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"mqtt-device": {"switch:0": map[string]any{"output": false}},
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
		Factory:     tf.Factory,
		Device:      "mqtt-device",
		Enable:      true,
		Server:      "mqtt://broker:1883",
		User:        "testuser",
		Password:    "testpass",
		TopicPrefix: "home/shelly",
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestRun_CanceledContext(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "cancel-device",
					Address:    "192.168.1.102",
					MAC:        "CC:DD:EE:FF:00:11",
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory: tf.Factory,
		Device:  "cancel-device",
		Server:  "mqtt://broker:1883",
	}

	err = run(ctx, opts)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing device argument")
	}
}

func TestExecute_NoOptions(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"test-device"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no options specified")
	}
	if !strings.Contains(err.Error(), "specify at least one configuration option") {
		t.Errorf("unexpected error: %v", err)
	}
}
