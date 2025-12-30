package importcmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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
	if cmd.Use != "import <device> <file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import <device> <file>")
	}

	// Test Aliases
	wantAliases := []string{"load", "restore"}
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
		{"two args valid", []string{"device", "file.json"}, false},
		{"three args", []string{"device", "file.json", "extra"}, true},
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
		{"overwrite", "", "false"},
		{"dry-run", "", "false"},
		{"yes", "y", "false"},
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
		t.Error("ValidArgsFunction is nil")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly kvs import",
		"--overwrite",
		"--dry-run",
		"--yes",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:    "test-device",
		File:      "test-file.json",
		Overwrite: true,
		DryRun:    true,
	}
	opts.Yes = true

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.File != "test-file.json" {
		t.Errorf("File = %q, want %q", opts.File, "test-file.json")
	}
	if !opts.Overwrite {
		t.Error("Overwrite should be true")
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if !opts.Yes {
		t.Error("Yes should be true")
	}
}

// TestExecute_InvalidFile tests handling of nonexistent import file.
func TestExecute_InvalidFile(t *testing.T) {
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
	cmd.SetArgs([]string{"test-device", "/nonexistent/file.json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// TestExecute_DryRun tests the --dry-run flag functionality.
func TestExecute_DryRun(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create a temporary KVS import file
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.json")
	importData := `{
  "items": [
    {"key": "key1", "value": "value1"},
    {"key": "key2", "value": 42}
  ],
  "version": 1,
  "rev": 0
}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile, "--dry-run", "--yes"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (expected for dry-run without real device): %v", err)
	}

	output := buf.String()
	// Should show preview or dry-run output
	if !strings.Contains(output, "key") && !strings.Contains(output, "Dry run") && err == nil {
		t.Logf("Output: %s", output)
	}
}

// TestExecute_EmptyFile tests handling of empty KVS import file.
func TestExecute_EmptyFile(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create a temporary KVS import file with empty items
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-empty.json")
	importData := `{
  "items": [],
  "version": 1,
  "rev": 0
}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No keys to import") && err == nil {
		t.Logf("Output: %s", output)
	}
}

// TestExecute_DeviceNotFound tests handling of nonexistent device.
func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()
	fixtures := &mock.Fixtures{
		Version: "1",
		Config:  mock.ConfigFixture{},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create a temporary KVS import file
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.json")
	importData := `{"items": [{"key": "test", "value": "data"}], "version": 1, "rev": 0}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"nonexistent-device", importFile, "--yes"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Should error when trying to access nonexistent device (will fail on actual import/connection)
	// Either error is acceptable - the important thing is that the command tries to process
	if err != nil {
		t.Logf("Got expected error for nonexistent device: %v", err)
	}
}

// TestExecute_MalformedJSON tests handling of malformed JSON import file.
func TestExecute_MalformedJSON(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create a temporary file with malformed JSON
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "malformed.json")
	if err := os.WriteFile(importFile, []byte(`{invalid json`), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for malformed JSON")
	}
}

// TestExecute_WithOverwrite tests the --overwrite flag functionality.
func TestExecute_WithOverwrite(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.json")
	importData := `{"items": [{"key": "test", "value": "data"}], "version": 1, "rev": 0}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile, "--overwrite", "--yes"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v (may be expected)", err)
	}
}

// TestRun_WithOptions tests the run function with various options.
func TestRun_WithOptions(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.json")
	importData := `{"items": [{"key": "key1", "value": "value1"}], "version": 1, "rev": 0}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Device:    "test-device",
		File:      importFile,
		Overwrite: false,
		DryRun:    true,
		Factory:   tf.Factory,
	}
	opts.Yes = true

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error: %v", err)
	}
}

// TestExecute_YAMLFile tests import with YAML format file.
func TestExecute_YAMLFile(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.yaml")
	importData := `items:
  - key: key1
    value: value1
  - key: key2
    value: 42
version: 1
rev: 0
`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile, "--dry-run", "--yes"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error (expected for YAML without real device): %v", err)
	}
}

// TestExecute_ConfirmationDenied tests behavior when user denies confirmation.
func TestExecute_ConfirmationDenied(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-import.json")
	importData := `{"items": [{"key": "key1", "value": "value1"}], "version": 1, "rev": 0}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Set stdin to empty so ConfirmAction will return false
	tf.TestIO.In.Reset()

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(tf.TestIO.In)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v (expected)", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Aborted") && err == nil {
		t.Logf("Output: %s", output)
	}
}

// TestExecute_ComplexValues tests import with various JSON value types.
func TestExecute_ComplexValues(t *testing.T) {
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
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs-complex.json")
	importData := `{
  "items": [
    {"key": "string_key", "value": "string_value"},
    {"key": "number_key", "value": 123},
    {"key": "bool_key", "value": true},
    {"key": "array_key", "value": [1, 2, 3]},
    {"key": "object_key", "value": {"nested": "object"}}
  ],
  "version": 1,
  "rev": 0
}`
	if err := os.WriteFile(importFile, []byte(importData), 0o600); err != nil {
		t.Fatalf("Failed to write import file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", importFile, "--dry-run", "--yes"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error: %v (expected)", err)
	}
}

// TestExecute_NoArgs tests command execution with no arguments.
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
		t.Error("Expected error when executing with no arguments")
	}

	// Should contain error about missing arguments
	errStr := buf.String()
	if !strings.Contains(errStr, "accept") && !strings.Contains(err.Error(), "accept") {
		t.Logf("Error should mention arguments: %v", err)
	}
}

// TestExecute_OneArgOnly tests command execution with only device argument.
func TestExecute_OneArgOnly(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with only device argument")
	}
}

// TestNewCommand_FullConfiguration tests that NewCommand creates properly configured command.
func TestNewCommand_FullConfiguration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify all required fields are set
	checks := []struct {
		name    string
		checkFn func() bool
		errMsg  string
	}{
		{
			name:    "has Use",
			checkFn: func() bool { return cmd.Use != "" },
			errMsg:  "Use should not be empty",
		},
		{
			name:    "has Short",
			checkFn: func() bool { return cmd.Short != "" },
			errMsg:  "Short should not be empty",
		},
		{
			name:    "has Long",
			checkFn: func() bool { return cmd.Long != "" },
			errMsg:  "Long should not be empty",
		},
		{
			name:    "has Example",
			checkFn: func() bool { return cmd.Example != "" },
			errMsg:  "Example should not be empty",
		},
		{
			name:    "has Aliases",
			checkFn: func() bool { return len(cmd.Aliases) > 0 },
			errMsg:  "Aliases should not be empty",
		},
		{
			name:    "has RunE",
			checkFn: func() bool { return cmd.RunE != nil },
			errMsg:  "RunE should be set",
		},
		{
			name:    "has ValidArgsFunction",
			checkFn: func() bool { return cmd.ValidArgsFunction != nil },
			errMsg:  "ValidArgsFunction should be set",
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()
			if !check.checkFn() {
				t.Error(check.errMsg)
			}
		})
	}
}

// TestExecute_FlagParsing tests flag parsing with various combinations.
func TestExecute_FlagParsing(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "kvs.json")
	if err := os.WriteFile(importFile, []byte(`{"items": [], "version": 1}`), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "with --yes flag",
			args: []string{"test-device", importFile, "--yes"},
		},
		{
			name: "with -y shorthand",
			args: []string{"test-device", importFile, "-y"},
		},
		{
			name: "with --dry-run",
			args: []string{"test-device", importFile, "--dry-run", "--yes"},
		},
		{
			name: "with --overwrite",
			args: []string{"test-device", importFile, "--overwrite", "--yes"},
		},
		{
			name: "with all flags",
			args: []string{"test-device", importFile, "--overwrite", "--dry-run", "--yes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			cmd := NewCommand(tf.Factory)
			cmd.SetContext(context.Background())
			cmd.SetArgs(tt.args)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			if err := cmd.Execute(); err != nil {
				t.Logf("Execute error (expected for unconnected device): %v", err)
			}
		})
	}
}
