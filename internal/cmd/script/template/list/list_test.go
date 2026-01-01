package list

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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
	if cmd.Use != "list" {
		t.Errorf("Use = %q, want %q", cmd.Use, "list")
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
		"shelly script template list",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestRun_ListsTemplates(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain at least one built-in template
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
	if !strings.Contains(output, "power-monitor") {
		t.Errorf("output should contain 'power-monitor', got: %q", output)
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_JSONOutput(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Set JSON output format
	viper.Set("output", "json")
	defer viper.Reset()

	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should be valid JSON containing template names
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
	// JSON output should have quotes and braces
	if !strings.Contains(output, `"name"`) && !strings.Contains(output, `"Name"`) {
		t.Errorf("expected JSON output with name field, got: %q", output)
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_YAMLOutput(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Set YAML output format
	viper.Set("output", "yaml")
	defer viper.Reset()

	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain template names
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
}

//nolint:paralleltest // Uses viper global state
func TestRun_TableOutput(t *testing.T) {
	tf := factory.NewTestFactory(t)

	// Ensure table format
	viper.Set("output", "table")
	defer viper.Reset()

	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain column headers for table output (uppercase)
	if !strings.Contains(output, "NAME") {
		t.Errorf("output should contain 'NAME' column header, got: %q", output)
	}
	// Should contain template data
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
}

func TestNewCommand_OutputFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Check -o flag exists
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("--output flag not found")
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	// Should contain built-in templates
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestRun_SortsTemplatesByName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Templates should be sorted alphabetically
	// energy-logger < motion-light < power-monitor < schedule-helper < toggle-sync
	energyPos := strings.Index(output, "energy-logger")
	motionPos := strings.Index(output, "motion-light")
	powerPos := strings.Index(output, "power-monitor")

	if energyPos >= motionPos || motionPos >= powerPos {
		t.Errorf("templates should be sorted alphabetically, got: %s", output)
	}
}
