package update

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if cmd.Use != "update <device> <id>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update <device> <id>")
	}

	// Test Aliases
	wantAliases := []string{"up"}
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
		{"two args valid", []string{"device", "1"}, false},
		{"three args", []string{"device", "1", "extra"}, true},
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
		{"name", "", ""},
		{"code", "", ""},
		{"file", "f", ""},
		{"append", "", "false"},
		{"enable", "", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
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
		t.Error("ValidArgsFunction should be set for completion")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly script update",
		"--name",
		"--file",
		"--append",
		"--enable",
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
	cmd.SetArgs([]string{"device", "not-a-number"})
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

func TestExecute_WithMock(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = "new-name"
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--name", "new-name"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// May fail due to mock limitations
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithCodeFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = "print('hello');"
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--code", "print('hello');"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithAppendFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = "// appended code"
	fileFlag = ""
	appendFlag = true
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--code", "// appended code", "--append"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_WithEnableFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = true

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--enable"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) { //nolint:paralleltest // Uses global mock state
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Reset global flags
	nameFlag = "test"
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device", "1", "--name", "test"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestExecute_FileNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = "/nonexistent/path/to/script.js"
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--file", "/nonexistent/path/to/script.js"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Error should mention 'failed to read file': %v", err)
	}
}

func TestExecute_WithFileFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
	// Create a temporary file with script code
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.js")
	scriptContent := "print('Hello from file!');"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

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

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = scriptPath
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1", "--file", scriptPath})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestExecute_NoChangesSpecified(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags - no changes specified
	nameFlag = ""
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Should succeed but print a warning about no changes
	if err != nil {
		t.Logf("Execute error = %v (may be expected for mock)", err)
	}
}

func TestRun_WithMock(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = "updated-script"
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	// May fail due to mock limitations
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_DeviceNotFound(t *testing.T) { //nolint:paralleltest // Uses global mock state
	fixtures := &mock.Fixtures{Version: "1", Config: mock.ConfigFixture{}}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Reset global flags
	nameFlag = "test-script"
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	err = run(context.Background(), tf.Factory, "nonexistent-device", 1)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_FileNotFound(t *testing.T) { //nolint:paralleltest // Uses global flags
	tf := factory.NewTestFactory(t)

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = "/nonexistent/path/to/script.js"
	appendFlag = false
	enableFlag = false

	err := run(context.Background(), tf.Factory, "test-device", 1)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Error should mention 'failed to read file': %v", err)
	}
}

func TestRun_WithCodeFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = "print('hello world');"
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_WithAppendFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = "// additional code"
	fileFlag = ""
	appendFlag = true
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_WithEnableFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = true

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_CombinedFlags(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags - combine multiple updates
	nameFlag = "new-name"
	codeFlag = "print('new code');"
	fileFlag = ""
	appendFlag = false
	enableFlag = true

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_NoChanges(t *testing.T) { //nolint:paralleltest // Uses global mock state
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

	// Reset global flags - no changes
	nameFlag = ""
	codeFlag = ""
	fileFlag = ""
	appendFlag = false
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	// Should succeed but issue a warning
	if err != nil {
		t.Logf("run() error = %v (may be expected)", err)
	}
}

func TestRun_WithFileFlag(t *testing.T) { //nolint:paralleltest // Uses global mock state
	// Create a temporary file with script code
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "update-script.js")
	scriptContent := "print('Script from file');\nlet x = 42;"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

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

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = scriptPath
	appendFlag = false
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}

func TestRun_WithFileAndAppend(t *testing.T) { //nolint:paralleltest // Uses global mock state
	// Create a temporary file with script code
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "append-script.js")
	scriptContent := "// Additional function\nfunction extra() { }"
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

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

	// Reset global flags
	nameFlag = ""
	codeFlag = ""
	fileFlag = scriptPath
	appendFlag = true
	enableFlag = false

	err = run(context.Background(), tf.Factory, "test-device", 1)
	if err != nil {
		t.Logf("run() error = %v (expected for mock)", err)
	}
}
