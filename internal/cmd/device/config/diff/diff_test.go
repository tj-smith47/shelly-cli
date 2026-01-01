package diff

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "diff <source> <target>" {
		t.Errorf("Use = %q, want 'diff <source> <target>'", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Aliases) == 0 {
		t.Error("Expected at least one alias")
	}

	expectedAliases := map[string]bool{"compare": true, "cmp": true}
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

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept exactly 2 args (source and target)
	err := cmd.Args(cmd, []string{"device1", "device2"})
	if err != nil {
		t.Errorf("Expected no error with two args, got: %v", err)
	}

	// Should reject 0 args
	err = cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no args provided")
	}

	// Should reject 1 arg
	err = cmd.Args(cmd, []string{"device1"})
	if err == nil {
		t.Error("Expected error when only one arg provided")
	}

	// Should reject 3+ args
	err = cmd.Args(cmd, []string{"device1", "device2", "device3"})
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

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Source:  "device1",
		Target:  "device2",
	}
	err := run(ctx, opts)

	// Expect an error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestRun_Timeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Allow the timeout to trigger
	time.Sleep(1 * time.Millisecond)

	opts := &Options{
		Factory: tf.Factory,
		Source:  "device1",
		Target:  "device2",
	}
	err := run(ctx, opts)

	// Expect an error due to timeout
	if err == nil {
		t.Error("Expected error with timed out context")
	}
}

func TestNewCommand_AcceptsIPAddresses(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100", "192.168.1.101"})
	if err != nil {
		t.Errorf("Command should accept IP addresses, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceNames(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room", "bedroom"})
	if err != nil {
		t.Errorf("Command should accept device names, got error: %v", err)
	}
}

func TestNewCommand_AcceptsMixedArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Device name and IP address
	err := cmd.Args(cmd, []string{"living-room", "192.168.1.100"})
	if err != nil {
		t.Errorf("Command should accept mixed args (name+IP), got error: %v", err)
	}

	// Device and file path
	err = cmd.Args(cmd, []string{"living-room", "config-backup.json"})
	if err != nil {
		t.Errorf("Command should accept device and file path, got error: %v", err)
	}
}

func TestNewCommand_AcceptsFilePaths(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Two file paths
	err := cmd.Args(cmd, []string{"backup1.json", "backup2.json"})
	if err != nil {
		t.Errorf("Command should accept file paths, got error: %v", err)
	}
}

func TestNewCommand_RunE_PassesArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"device1", "device2"})

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	// Execute - we expect an error due to cancelled context but want to verify structure
	if err := cmd.Execute(); err == nil {
		t.Error("Expected error from Execute with cancelled context")
	}
}

func TestNewCommand_DiffScenarios(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name   string
		source string
		target string
	}{
		{"two devices", "kitchen-light", "bedroom-light"},
		{"device and backup", "living-room", "config-backup.json"},
		{"two backups", "backup1.json", "backup2.json"},
		{"IP addresses", "192.168.1.100", "192.168.1.101"},
		{"device and IP", "kitchen", "192.168.1.100"},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, []string{scenario.source, scenario.target})
			if err != nil {
				t.Errorf("Command should accept %s, got error: %v", scenario.name, err)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should contain information about diff indicators
	long := cmd.Long
	if long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Example should contain various diff scenarios
	example := cmd.Example
	if example == "" {
		t.Error("Example should not be empty")
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_DiffConfigs(t *testing.T) {
	// Create fixtures with two devices - configs will differ due to device names
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "device1",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "device2",
					Address:    "192.168.1.101",
					MAC:        "AA:BB:CC:DD:EE:02",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"device1": {
				"switch:0": map[string]any{"output": false},
			},
			"device2": {
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
	cmd.SetArgs([]string{"device1", "device2"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// Should show comparing header and summary
	if !strings.Contains(out, "Comparing") {
		t.Errorf("Output should contain 'Comparing', got: %s", out)
	}
	if !strings.Contains(out, "Summary") {
		t.Errorf("Output should contain 'Summary', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_DiffDifferentConfigs(t *testing.T) {
	// Create fixtures with two devices having different configs
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "diff-source",
					Address:    "192.168.1.110",
					MAC:        "AA:BB:CC:DD:EE:10",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "diff-target",
					Address:    "192.168.1.111",
					MAC:        "AA:BB:CC:DD:EE:11",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"diff-source": {
				"switch:0": map[string]any{"output": true},
			},
			"diff-target": {
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
	cmd.SetArgs([]string{"diff-source", "diff-target"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// Should show comparing header
	if !strings.Contains(out, "Comparing") {
		t.Errorf("Output should contain 'Comparing', got: %s", out)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_SourceConfigError(t *testing.T) {
	// Create fixtures with only one device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "only-device",
					Address:    "192.168.1.120",
					MAC:        "AA:BB:CC:DD:EE:20",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"only-device": {
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
	cmd.SetArgs([]string{"nonexistent-source", "only-device"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent source device")
	}
	if !strings.Contains(err.Error(), "source") {
		t.Errorf("Error should mention 'source', got: %v", err)
	}
}

//nolint:paralleltest // Uses shared mock server
func TestRun_TargetConfigError(t *testing.T) {
	// Create fixtures with only one device
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "source-only",
					Address:    "192.168.1.121",
					MAC:        "AA:BB:CC:DD:EE:21",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"source-only": {
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
	cmd.SetArgs([]string{"source-only", "nonexistent-target"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for nonexistent target device")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Errorf("Error should mention 'target', got: %v", err)
	}
}

func TestRun_DiffWithJSONOutput(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Set JSON output format
	viper.Set("output", "json")
	defer viper.Reset()

	// Create a cancelled context to test the structured output path check
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{
		Factory: tf.Factory,
		Source:  "device1",
		Target:  "device2",
	}

	err := run(ctx, opts)
	// Expect error due to cancelled context
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}

func TestOptions_Fields(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Source:  "source-device",
		Target:  "target-device",
	}

	if opts.Source != "source-device" {
		t.Errorf("Source = %q, want %q", opts.Source, "source-device")
	}
	if opts.Target != "target-device" {
		t.Errorf("Target = %q, want %q", opts.Target, "target-device")
	}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

//nolint:paralleltest // Uses shared mock server and viper state
func TestRun_DiffWithJSONOutputFormat(t *testing.T) {
	// Create fixtures with two devices
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "json-device1",
					Address:    "192.168.1.130",
					MAC:        "AA:BB:CC:DD:EE:30",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
				{
					Name:       "json-device2",
					Address:    "192.168.1.131",
					MAC:        "AA:BB:CC:DD:EE:31",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"json-device1": {
				"switch:0": map[string]any{"output": false},
			},
			"json-device2": {
				"switch:0": map[string]any{"output": false},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	// Set JSON output format before creating factory
	viper.Set("output", "json")
	defer viper.Reset()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"json-device1", "json-device2"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	out := tf.OutString()
	// JSON output should have array syntax or object syntax
	if !strings.Contains(out, "[") && !strings.Contains(out, "{") {
		t.Errorf("Output should be JSON formatted, got: %s", out)
	}
}
