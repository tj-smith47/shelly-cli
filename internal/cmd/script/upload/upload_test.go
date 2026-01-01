package upload

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
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
	if cmd.Use != "upload <device> <id> <file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "upload <device> <id> <file>")
	}

	// Test Aliases
	wantAliases := []string{"put"}
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
		{"one arg", []string{"device"}, true},
		{"two args", []string{"device", "1"}, true},
		{"three args valid", []string{"device", "1", "file.js"}, false},
		{"four args", []string{"device", "1", "file.js", "extra"}, true},
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

	// Test append flag
	flag := cmd.Flags().Lookup("append")
	if flag == nil {
		t.Fatal("--append flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("--append default = %q, want %q", flag.DefValue, "false")
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
		t.Error("ValidArgsFunction should be set for completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly script upload",
		"--append",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_InvalidScriptID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "not-a-number", "file.js"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid script ID")
	}

	if !strings.Contains(err.Error(), "invalid script ID") {
		t.Errorf("expected 'invalid script ID' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global filesystem
func TestRun_FileNotFound(t *testing.T) {
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		ID:      1,
		File:    "/nonexistent/script.js",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("error = %v, want to contain 'failed to read file'", err)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_SuccessfulUpload(t *testing.T) {
	// Set up in-memory filesystem with script file
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	scriptContent := `// Test script
print("Hello World!");`
	if err := afero.WriteFile(memFs, "/scripts/test.js", []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
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
		t.Fatalf("failed to start demo: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory: tf.Factory,
		Device:  "test-device",
		ID:      1,
		File:    "/scripts/test.js",
		Append:  false,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Uploaded") {
		t.Errorf("output should contain 'Uploaded', got: %q", output)
	}
	if !strings.Contains(output, "bytes") {
		t.Errorf("output should contain 'bytes', got: %q", output)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_SuccessfulAppend(t *testing.T) {
	// Set up in-memory filesystem with script file
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	scriptContent := `// Additional code
function extra() { return 42; }`
	if err := afero.WriteFile(memFs, "/scripts/extra.js", []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "append-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"append-device": {},
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
		Device:  "append-device",
		ID:      1,
		File:    "/scripts/extra.js",
		Append:  true,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Appended") {
		t.Errorf("output should contain 'Appended', got: %q", output)
	}
	if !strings.Contains(output, "bytes") {
		t.Errorf("output should contain 'bytes', got: %q", output)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestRun_DeviceNotFound(t *testing.T) {
	// Set up in-memory filesystem with script file
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	scriptContent := `print("test");`
	if err := afero.WriteFile(memFs, "/scripts/test.js", []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
		File:    "/scripts/test.js",
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestNewCommand_Execute(t *testing.T) {
	// Set up in-memory filesystem with script file
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	scriptContent := `print("execute test");`
	if err := afero.WriteFile(memFs, "/scripts/exec.js", []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "exec-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"exec-device": {},
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
	cmd.SetArgs([]string{"exec-device", "1", "/scripts/exec.js"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Uploaded") {
		t.Errorf("output should contain 'Uploaded', got: %q", output)
	}
}

//nolint:paralleltest // Uses mock infrastructure with global state
func TestNewCommand_ExecuteWithAppend(t *testing.T) {
	// Set up in-memory filesystem with script file
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(nil)

	scriptContent := `// append content`
	if err := afero.WriteFile(memFs, "/scripts/append.js", []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fixtures := &mock.Fixtures{
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "append-exec-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Model:      "SNSW-001P16EU",
					Type:       "Plus1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"append-exec-device": {},
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
	cmd.SetArgs([]string{"append-exec-device", "2", "/scripts/append.js", "--append"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Appended") {
		t.Errorf("output should contain 'Appended', got: %q", output)
	}
}
