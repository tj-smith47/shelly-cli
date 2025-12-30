package configimport

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// testFilePerms is the file permission for test config files.
const testFilePerms = 0o600

// Flag default value constants for tests.
const (
	flagDefFalse = "false"
	flagDefTrue  = "true"
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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "import <device> <file>" {
		t.Errorf("Use = %q, want 'import <device> <file>'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) < 2 {
		t.Errorf("Expected at least 2 aliases, got %d", len(cmd.Aliases))
	}

	expectedAliases := map[string]bool{"restore": true, "load": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("Unexpected alias: %s", alias)
		}
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != flagDefFalse {
		t.Errorf("dry-run default = %q, want false", dryRunFlag.DefValue)
	}

	mergeFlag := cmd.Flags().Lookup("merge")
	if mergeFlag == nil {
		t.Fatal("merge flag not found")
	}
	if mergeFlag.DefValue != flagDefTrue {
		t.Errorf("merge default = %q, want true", mergeFlag.DefValue)
	}

	overwriteFlag := cmd.Flags().Lookup("overwrite")
	if overwriteFlag == nil {
		t.Fatal("overwrite flag not found")
	}
	if overwriteFlag.DefValue != flagDefFalse {
		t.Errorf("overwrite default = %q, want false", overwriteFlag.DefValue)
	}
}

func TestNewCommand_RequiresArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should require exactly 2 arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err == nil {
		t.Error("Expected error when only one arg provided")
	}

	err = cmd.Args(cmd, []string{"device1", "config.json"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "config.json", "extra"})
	if err == nil {
		t.Error("Expected error when too many args provided")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			wantOK:    true,
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "uses ExactArgs(2)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			wantOK:    true,
			errMsg:    "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "dry-run flag",
			args:    []string{"--dry-run"},
			wantErr: false,
		},
		{
			name:    "merge flag false",
			args:    []string{"--merge=false"},
			wantErr: false,
		},
		{
			name:    "overwrite flag",
			args:    []string{"--overwrite"},
			wantErr: false,
		},
		{
			name:    "all flags",
			args:    []string{"--dry-run", "--overwrite"},
			wantErr: false,
		},
		{
			name:    "no flags",
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"sys":{"name":"test"}}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"sys":{"name":"test"}}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestRun_FileNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: "/nonexistent/path/config.json",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Error should mention file read failure, got: %v", err)
	}
}

func TestRun_InvalidJSON(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file with invalid JSON
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{invalid json`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid JSON/YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse file") {
		t.Errorf("Error should mention parse failure, got: %v", err)
	}
}

func TestRun_InvalidYAML(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file with invalid YAML (not valid JSON either)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("invalid:\n  - @#$%^&*(\n  unclosed:"), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse file") {
		t.Errorf("Error should mention parse failure, got: %v", err)
	}
}

func TestNewCommand_AcceptsIPAddress(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100", "config.json"})
	if err != nil {
		t.Errorf("Command should accept IP address as device, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceName(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room", "config.json"})
	if err != nil {
		t.Errorf("Command should accept device name, got error: %v", err)
	}
}

func TestNewCommand_AcceptsFileFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"json file", "config.json", false},
		{"yaml file", "config.yaml", false},
		{"yml file", "config.yml", false},
		{"with path", "/tmp/config.json", false},
		{"no extension", "config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Args(cmd, []string{"device", tt.file})
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if example == "" {
		t.Fatal("Example should not be empty")
	}

	// Check for expected patterns
	patterns := []string{"shelly", "config", "import"}
	for _, pattern := range patterns {
		if !strings.Contains(example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description should not be empty")
	}

	// Long should be more descriptive than Short
	if len(cmd.Long) <= len(cmd.Short) {
		t.Error("Long description should be longer than Short description")
	}
}

func TestNewCommand_RunE_PassesArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"sys":{"name":"test"}}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", configFile})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_RunE_WithDryRun(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"sys":{"name":"test"}}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", configFile, "--dry-run"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
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

//nolint:paralleltest // Uses shared mock server
func TestRun_ImportJSONConfig(t *testing.T) {
	// Create fixtures with a Gen2 switch device
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
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with JSON config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configData := `{
		"sys": {
			"device": {"name": "New Name"}
		}
	}`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"test-device", configFile})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "Configuration imported") {
		t.Errorf("Output should contain 'Configuration imported', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_ImportYAMLConfig(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "yaml-device",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"yaml-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with YAML config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configData := `sys:
  device:
    name: New Name
`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"yaml-device", configFile})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "Configuration imported") {
		t.Errorf("Output should contain 'Configuration imported', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_DryRun(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "dryrun-device",
					Address:    "192.168.1.102",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"dryrun-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with JSON config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configData := `{"sys": {"device": {"name": "Updated Name"}}}`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command with --dry-run
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"dryrun-device", configFile, "--dry-run"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains dry run message
	out := tf.OutString()
	if !strings.Contains(out, "Dry run") {
		t.Errorf("Output should contain 'Dry run', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_DeviceNotFound(t *testing.T) {
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

	// Create temp file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{"sys":{"name":"test"}}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"nonexistent-device", configFile})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_WithOverwriteFlag(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "overwrite-device",
					Address:    "192.168.1.103",
					MAC:        "AA:BB:CC:DD:EE:03",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"overwrite-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with JSON config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configData := `{"sys": {"device": {"name": "Overwritten"}}}`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command with --overwrite
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"overwrite-device", configFile, "--overwrite"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "Configuration imported") {
		t.Errorf("Output should contain 'Configuration imported', got: %s", out)
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory:   f,
		Device:    "test-device",
		FilePath:  "/path/to/config.json",
		DryRun:    true,
		Merge:     true,
		Overwrite: false,
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}

	if opts.FilePath != "/path/to/config.json" {
		t.Errorf("FilePath = %q, want %q", opts.FilePath, "/path/to/config.json")
	}

	if !opts.DryRun {
		t.Error("DryRun should be true")
	}

	if !opts.Merge {
		t.Error("Merge should be true")
	}

	if opts.Overwrite {
		t.Error("Overwrite should be false")
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}
}

func TestRun_EmptyConfigFile(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file with empty JSON object
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configFile, []byte(`{}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create a cancelled context to just test file parsing
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	// File should parse successfully, but fail due to cancelled context
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
	// Should not be a parse error since {} is valid JSON
	if strings.Contains(err.Error(), "failed to parse file") {
		t.Errorf("Should not be a parse error for valid JSON, got: %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_ComplexJSONConfig(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "complex-device",
					Address:    "192.168.1.104",
					MAC:        "AA:BB:CC:DD:EE:04",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"complex-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with complex nested JSON config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configData := `{
		"sys": {
			"device": {
				"name": "Complex Device",
				"eco_mode": true
			}
		},
		"switch:0": {
			"name": "Main Switch",
			"initial_state": "on"
		},
		"wifi": {
			"sta": {
				"ssid": "HomeNetwork",
				"pass": "secret123"
			}
		}
	}`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"complex-device", configFile})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "Configuration imported") {
		t.Errorf("Output should contain 'Configuration imported', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_ComplexYAMLConfig(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "yaml-complex",
					Address:    "192.168.1.105",
					MAC:        "AA:BB:CC:DD:EE:05",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"yaml-complex": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with complex YAML config
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configData := `sys:
  device:
    name: YAML Complex Device
    eco_mode: true
switch:0:
  name: Main Switch
  initial_state: "on"
wifi:
  sta:
    ssid: HomeNetwork
    pass: secret123
`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"yaml-complex", configFile})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "Configuration imported") {
		t.Errorf("Output should contain 'Configuration imported', got: %s", out)
	}
}

func TestNewCommand_MergeFlagDefault(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse empty args to get default values
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	mergeFlag := cmd.Flags().Lookup("merge")
	if mergeFlag == nil {
		t.Fatal("merge flag not found")
	}

	// Merge should default to true
	if mergeFlag.DefValue != flagDefTrue {
		t.Errorf("merge default should be true, got %q", mergeFlag.DefValue)
	}
}

func TestNewCommand_OverwriteFlagDefault(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse empty args to get default values
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	overwriteFlag := cmd.Flags().Lookup("overwrite")
	if overwriteFlag == nil {
		t.Fatal("overwrite flag not found")
	}

	// Overwrite should default to false
	if overwriteFlag.DefValue != flagDefFalse {
		t.Errorf("overwrite default should be false, got %q", overwriteFlag.DefValue)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_DryRunWithDiffOutput(t *testing.T) {
	// Create fixtures with a Gen2 switch device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "diff-device",
					Address:    "192.168.1.106",
					MAC:        "AA:BB:CC:DD:EE:06",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"diff-device": {
				"switch:0": map[string]any{
					"output": false,
				},
			},
		},
	}

	// Start demo mode with fixtures
	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Create test factory and inject mock
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create temp file with config that has different values
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	configData := `{"sys": {"device": {"name": "Changed Name", "eco_mode": true}}}`
	if err := os.WriteFile(configFile, []byte(configData), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Create and execute command with --dry-run
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"diff-device", configFile, "--dry-run"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains dry run message
	out := tf.OutString()
	if !strings.Contains(out, "Dry run") {
		t.Errorf("Output should contain 'Dry run', got: %s", out)
	}
	// Should contain "changes that would be applied"
	if !strings.Contains(out, "changes that would be applied") {
		t.Errorf("Output should contain 'changes that would be applied', got: %s", out)
	}
}

func TestRun_JSONParsedFirstBeforeYAML(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create temp file with valid JSON (even though it has .yaml extension)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	// This is valid JSON, should be parsed as JSON first
	if err := os.WriteFile(configFile, []byte(`{"test": "value"}`), testFilePerms); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: configFile,
	}

	// File should parse successfully as JSON, fail due to cancelled context
	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
	// Should not be a parse error
	if strings.Contains(err.Error(), "failed to parse file") {
		t.Errorf("Should not be a parse error for valid JSON, got: %v", err)
	}
}
