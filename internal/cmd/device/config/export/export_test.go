package export

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "export <device> <file>" {
		t.Errorf("Use = %q, want 'export <device> <file>'", cmd.Use)
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

	if len(cmd.Aliases) < 2 {
		t.Errorf("Expected at least 2 aliases, got %d", len(cmd.Aliases))
	}

	expectedAliases := map[string]bool{"backup": true, "save": true}
	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("Unexpected alias: %s", alias)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("format shorthand = %q, want f", formatFlag.Shorthand)
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag.DefValue != "json" {
		t.Errorf("format default = %q, want json", formatFlag.DefValue)
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

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for device then file completion")
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
			name:    "format flag short json",
			args:    []string{"-f", "json"},
			wantErr: false,
		},
		{
			name:    "format flag long json",
			args:    []string{"--format", "json"},
			wantErr: false,
		},
		{
			name:    "format flag yaml",
			args:    []string{"-f", "yaml"},
			wantErr: false,
		},
		{
			name:    "format flag yml",
			args:    []string{"-f", "yml"},
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: "output.json",
		Format:   "json",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory:  tf.Factory,
		Device:   "test-device",
		FilePath: "output.json",
		Format:   "json",
	}

	err := run(ctx, opts)
	if err == nil {
		t.Error("Expected error with timed out context")
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
		{"stdout", "-", false},
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
	patterns := []string{"shelly", "config", "export"}
	for _, pattern := range patterns {
		found := false
		for i := 0; i <= len(example)-len(pattern); i++ {
			if example[i:i+len(pattern)] == pattern {
				found = true
				break
			}
		}
		if !found {
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

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "output.json"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_RunE_WithFormatFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"my-device", "output.yaml", "-f", "yaml"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_StdoutOutput(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify stdout marker "-" is accepted
	err := cmd.Args(cmd, []string{"device", "-"})
	if err != nil {
		t.Errorf("Command should accept '-' for stdout, got error: %v", err)
	}
}

func TestNewCommand_FormatValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"yaml format", "yaml"},
		{"yml format", "yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags([]string{"--format", tt.format})
			if err != nil {
				t.Errorf("Format %q should be valid, got error: %v", tt.format, err)
			}
		})
	}
}

//nolint:paralleltest // Uses shared mock server and filesystem
func TestRun_ExportToJSONFile(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "export-device",
					Address:    "192.168.1.200",
					MAC:        "AA:BB:CC:DD:EE:40",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"export-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Set up in-memory filesystem
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(afero.NewOsFs())

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"export-device", "/tmp/config.json"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify file was created
	exists, err := afero.Exists(memFs, "/tmp/config.json")
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("Expected config file to be created")
	}

	// Verify output contains success message
	out := tf.OutString()
	if !strings.Contains(out, "exported") && !strings.Contains(out, "Configuration") {
		t.Errorf("Output should contain success message, got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server and filesystem
func TestRun_ExportToYAMLFile(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "yaml-export-device",
					Address:    "192.168.1.201",
					MAC:        "AA:BB:CC:DD:EE:41",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"yaml-export-device": {
				"switch:0": map[string]any{"output": true},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Set up in-memory filesystem
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(afero.NewOsFs())

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"yaml-export-device", "/tmp/config.yaml", "-f", "yaml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify file was created
	exists, err := afero.Exists(memFs, "/tmp/config.yaml")
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("Expected YAML config file to be created")
	}

	// Read file content and verify it's YAML
	content, err := afero.ReadFile(memFs, "/tmp/config.yaml")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if strings.Contains(string(content), "{") {
		t.Error("YAML output should not contain JSON braces")
	}
}

//nolint:paralleltest // Uses shared mock server and filesystem
func TestRun_ExportToYMLFile(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "yml-export-device",
					Address:    "192.168.1.202",
					MAC:        "AA:BB:CC:DD:EE:42",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"yml-export-device": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Set up in-memory filesystem
	memFs := afero.NewMemMapFs()
	config.SetFs(memFs)
	defer config.SetFs(afero.NewOsFs())

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"yml-export-device", "/tmp/config.yml", "-f", "yml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify file was created
	exists, err := afero.Exists(memFs, "/tmp/config.yml")
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("Expected YML config file to be created")
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_ExportToStdout(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "stdout-export-device",
					Address:    "192.168.1.203",
					MAC:        "AA:BB:CC:DD:EE:43",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"stdout-export-device": {
				"switch:0": map[string]any{"output": false},
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
	cmd.SetArgs([]string{"stdout-export-device", "-"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify output contains JSON
	out := tf.OutString()
	if !strings.Contains(out, "{") && !strings.Contains(out, "sys") {
		t.Errorf("Output should contain JSON config, got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_GetConfigError(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "existing-device",
					Address:    "192.168.1.204",
					MAC:        "AA:BB:CC:DD:EE:44",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"existing-device": {},
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
	cmd.SetArgs([]string{"nonexistent-device", "/tmp/config.json"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent device")
	}
}
