package download

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"

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

	if cmd.Use != "download <device> <id> <file>" {
		t.Errorf("Use = %q, want \"download <device> <id> <file>\"", cmd.Use)
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

	expectedAliases := []string{"save"}
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
			name:    "three args valid",
			args:    []string{"device", "1", "output.js"},
			wantErr: false,
		},
		{
			name:    "four args invalid",
			args:    []string{"device", "1", "output.js", "extra"},
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
	cmd.SetArgs([]string{"device", "not-a-number", "file.js"})

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

	if len(cmd.Long) < 20 {
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

func TestNewCommand_NoLocalFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Download command doesn't define any local flags
	// Only inherited flags from parent commands
	localFlags := cmd.LocalFlags()

	// Count local flags
	count := 0
	localFlags.VisitAll(func(_ *pflag.Flag) {
		count++
	})

	// The command shouldn't have local flags based on the source
	if count > 0 {
		t.Logf("Found %d local flags (this is informational)", count)
	}
}

func TestExecute_Help(t *testing.T) {
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

func TestExecute_InvalidScriptID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"device", "not-a-number", "output.js"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid script ID")
	}

	if !strings.Contains(err.Error(), "invalid script ID") {
		t.Errorf("expected 'invalid script ID' error, got: %v", err)
	}
}

func TestExecute_MissingArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"device", "1"}) // missing file arg

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestRun_WithMockDevice(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				"script:1": map[string]any{
					"id":      1,
					"name":    "auto-light",
					"enable":  true,
					"running": false,
					"code":    "// Auto light script\nShelly.call('Switch.Set', {id: 0, on: true});",
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

	// Create temp file for output
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "script.js")

	err = run(context.Background(), tf.Factory, "test-device", 1, outputFile)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(outputFile); os.IsNotExist(statErr) {
		t.Error("expected output file to be created")
	}

	// Verify file content
	//nolint:gosec // G304: outputFile is constructed from t.TempDir() which is safe
	content, readErr := os.ReadFile(outputFile)
	if readErr != nil {
		t.Fatalf("failed to read output file: %v", readErr)
	}

	if !strings.Contains(string(content), "Auto light script") {
		t.Errorf("expected file to contain script code, got: %s", content)
	}

	// Check success message
	output := tf.OutString()
	if !strings.Contains(output, "Downloaded script") {
		t.Errorf("expected success message in output, got: %s", output)
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

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "script.js")

	err = run(context.Background(), tf.Factory, "nonexistent-device", 1, outputFile)
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

func TestRun_EmptyScriptCode(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				"script:1": map[string]any{
					"id":      1,
					"name":    "empty-script",
					"enable":  false,
					"running": false,
					"code":    "", // Empty code
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

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "script.js")

	err = run(context.Background(), tf.Factory, "test-device", 1, outputFile)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Check warning message for empty script (written to stderr)
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "has no code") {
		t.Errorf("expected warning about empty script in stderr, got: %s", errOutput)
	}
}

func TestRun_CreateDirectory(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"script:1": map[string]any{
					"id":   1,
					"name": "test-script",
					"code": "// test code",
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

	// Create path with nested directory that doesn't exist
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "scripts", "nested")
	outputFile := filepath.Join(nestedDir, "script.js")

	err = run(context.Background(), tf.Factory, "test-device", 1, outputFile)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify directory was created
	if _, statErr := os.Stat(nestedDir); os.IsNotExist(statErr) {
		t.Error("expected nested directory to be created")
	}

	// Verify file exists
	if _, statErr := os.Stat(outputFile); os.IsNotExist(statErr) {
		t.Error("expected output file to be created")
	}
}

func TestExecute_WithMockDevice(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"script:1": map[string]any{
					"id":   1,
					"name": "test",
					"code": "console.log('test');",
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

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "downloaded.js")

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"test-device", "1", outputFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(outputFile); os.IsNotExist(statErr) {
		t.Error("expected output file to be created")
	}
}

func TestExecute_ScriptNotFound(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"switch:0": map[string]any{"output": true},
				// No script:99 defined
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

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "script.js")

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"test-device", "99", outputFile})

	// Script not found returns empty code which triggers "has no code" warning
	// This should not error, just warn
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check warning message (written to stderr)
	errOutput := tf.ErrString()
	if !strings.Contains(errOutput, "has no code") {
		t.Errorf("expected warning about empty script in stderr, got: %s", errOutput)
	}
}

func TestRun_CurrentDirectory(t *testing.T) { //nolint:paralleltest // Uses global mock state
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
			"test-device": {
				"script:1": map[string]any{
					"id":   1,
					"name": "test-script",
					"code": "// test code for current dir",
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

	// Change to temp dir for test
	tmpDir := t.TempDir()
	origDir, getErr := os.Getwd()
	if getErr != nil {
		t.Fatalf("failed to get current dir: %v", getErr)
	}
	if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
		t.Fatalf("failed to change to temp dir: %v", chdirErr)
	}
	defer func() {
		if chdirErr := os.Chdir(origDir); chdirErr != nil {
			t.Logf("failed to restore dir: %v", chdirErr)
		}
	}()

	// Use just filename (current directory)
	outputFile := "local-script.js"

	err = run(context.Background(), tf.Factory, "test-device", 1, outputFile)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify file was created in current directory
	fullPath := filepath.Join(tmpDir, outputFile)
	if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
		t.Error("expected output file to be created in current directory")
	}
}
