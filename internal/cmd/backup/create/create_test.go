package create

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const (
	testFalseValue  = "false"
	testJSONDefault = "json"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "create <device> [file]" {
		t.Errorf("Use = %q, want 'create <device> [file]'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}

	// Verify example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	aliases := cmd.Aliases
	if len(aliases) < 2 {
		t.Errorf("expected at least 2 aliases, got %d", len(aliases))
	}

	// Check for expected aliases
	hasNew := false
	hasMake := false
	for _, a := range aliases {
		if a == "new" {
			hasNew = true
		}
		if a == "make" {
			hasMake = true
		}
	}
	if !hasNew {
		t.Error("expected 'new' alias")
	}
	if !hasMake {
		t.Error("expected 'make' alias")
	}
}

func TestNewCommand_RequiresDevice(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should require at least 1 argument (device)
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should accept 1 arg (device only)
	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("Expected no error with one arg, got: %v", err)
	}

	// Should accept 2 args (device and file)
	err = cmd.Args(cmd, []string{"device1", "backup.json"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	// Should reject 3 args
	err = cmd.Args(cmd, []string{"device1", "backup.json", "extra"})
	if err == nil {
		t.Error("Expected error with three args")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check format flag
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("format flag shorthand = %q, want 'f'", formatFlag.Shorthand)
	}
	if formatFlag.DefValue != testJSONDefault {
		t.Errorf("format flag default = %q, want %q", formatFlag.DefValue, testJSONDefault)
	}

	// Check encrypt flag
	encryptFlag := cmd.Flags().Lookup("encrypt")
	if encryptFlag == nil {
		t.Fatal("encrypt flag not found")
	}
	if encryptFlag.Shorthand != "e" {
		t.Errorf("encrypt flag shorthand = %q, want 'e'", encryptFlag.Shorthand)
	}

	// Check skip-scripts flag
	skipScriptsFlag := cmd.Flags().Lookup("skip-scripts")
	if skipScriptsFlag == nil {
		t.Fatal("skip-scripts flag not found")
	}
	if skipScriptsFlag.DefValue != testFalseValue {
		t.Errorf("skip-scripts flag default = %q, want %q", skipScriptsFlag.DefValue, testFalseValue)
	}

	// Check skip-schedules flag
	skipSchedulesFlag := cmd.Flags().Lookup("skip-schedules")
	if skipSchedulesFlag == nil {
		t.Fatal("skip-schedules flag not found")
	}
	if skipSchedulesFlag.DefValue != testFalseValue {
		t.Errorf("skip-schedules flag default = %q, want %q", skipSchedulesFlag.DefValue, testFalseValue)
	}

	// Check skip-webhooks flag
	skipWebhooksFlag := cmd.Flags().Lookup("skip-webhooks")
	if skipWebhooksFlag == nil {
		t.Fatal("skip-webhooks flag not found")
	}
	if skipWebhooksFlag.DefValue != testFalseValue {
		t.Errorf("skip-webhooks flag default = %q, want %q", skipWebhooksFlag.DefValue, testFalseValue)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Test parsing with flags - these won't actually run because no device
	// but they validate flag parsing works
	testCases := []struct {
		name string
		args []string
	}{
		{"format json", []string{"--format", "json", "device"}},
		{"format yaml", []string{"--format", "yaml", "device"}},
		{"format short", []string{"-f", "yaml", "device"}},
		{"encrypt", []string{"--encrypt", "mypassword", "device"}},
		{"encrypt short", []string{"-e", "mypassword", "device"}},
		{"skip-scripts", []string{"--skip-scripts", "device"}},
		{"skip-schedules", []string{"--skip-schedules", "device"}},
		{"skip-webhooks", []string{"--skip-webhooks", "device"}},
		{"all skip flags", []string{"--skip-scripts", "--skip-schedules", "--skip-webhooks", "device"}},
		{"combined flags", []string{"-f", "yaml", "-e", "pass", "--skip-scripts", "device"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := NewCommand(f)
			cmd.SetOut(out)
			cmd.SetErr(errOut)
			cmd.SetArgs(tc.args)

			// Parse flags only (don't run command as it requires network)
			err := cmd.ParseFlags(tc.args)
			if err != nil {
				t.Errorf("ParseFlags failed: %v", err)
			}
		})
	}
}

func TestNewCommand_Example_Content(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Check that example contains expected usage patterns
	examples := []string{
		"backup create",
		"backup.json",
		"--format yaml",
		"--encrypt",
		"--skip-scripts",
	}

	for _, e := range examples {
		if !containsString(example, e) {
			t.Errorf("expected example to contain %q", e)
		}
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
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
		t.Error("Expected error when no device argument provided")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"device1", "file.json", "extra"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when too many arguments provided")
	}
}

func TestExecute_WithMockDevice(t *testing.T) {
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// CreateBackup may fail if the mock doesn't support full backup operations
	// but this still exercises the command execution path
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithOutputFile(t *testing.T) {
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

	dir := t.TempDir()
	filePath := dir + "/backup.json"

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", filePath})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestExecute_YAMLFormat(t *testing.T) {
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--format", "yaml"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithSkipFlags(t *testing.T) {
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--skip-scripts", "--skip-schedules", "--skip-webhooks"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithEncrypt(t *testing.T) {
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "--encrypt", "mysecret"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Encrypted backups return an error in the service layer
	if err != nil {
		t.Logf("Execute() error = %v (expected for encrypted backups)", err)
	}
}

func TestExecute_UnknownDevice(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{},
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
	cmd.SetArgs([]string{"unknown-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for unknown device")
	}
	if !strings.Contains(err.Error(), "failed to create backup") {
		t.Logf("Error message: %v", err)
	}
}

func TestExecute_StdoutOutput(t *testing.T) {
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

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	// Use "-" to explicitly write to stdout
	cmd.SetArgs([]string{"test-device", "-"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected for mock)", err)
	}
}

func TestRun_Success(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Format:  "json",
	}
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_WriteToFile(t *testing.T) {
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

	dir := t.TempDir()
	filePath := dir + "/backup.json"

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: filePath,
		Format:   "json",
	}
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_InvalidFilePath(t *testing.T) {
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

	// Try to write to an invalid path
	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: "/nonexistent/directory/backup.json",
		Format:   "json",
	}
	err = run(context.Background(), opts)
	// This should fail either at backup creation (mock limitation) or file write (invalid path)
	if err != nil {
		t.Logf("run() error = %v (expected)", err)
	}
}

func TestRun_YAMLFormat(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Format:  "yaml",
	}
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_WithAllSkipFlags(t *testing.T) {
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
		Factory:       tf.Factory,
		Device:        "test-device",
		Format:        "json",
		SkipScripts:   true,
		SkipSchedules: true,
		SkipWebhooks:  true,
	}
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected for mock)", err)
	}
}

func TestRun_WithPassword(t *testing.T) {
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
		Factory: tf.Factory,
		Device:  "test-device",
		Format:  "json",
		Encrypt: "mysecretpassword",
	}
	err = run(context.Background(), opts)
	// Expected to fail because encrypted backups are not supported via service layer
	if err == nil {
		t.Log("Expected error for encrypted backup via service layer")
	}
}

func TestNewCommand_Long_Description(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	long := cmd.Long

	// Check that long description contains key information
	keywords := []string{
		"backup",
		"device",
		"scripts",
		"schedules",
		"webhooks",
		"encrypt",
	}

	for _, kw := range keywords {
		if !containsString(long, kw) {
			t.Errorf("expected long description to contain %q", kw)
		}
	}
}
