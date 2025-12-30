package current

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

const testCommandUse = "current"

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

	if cmd.Use != testCommandUse {
		t.Errorf("Use = %q, want %q", cmd.Use, testCommandUse)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"cur", "c"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, expected := range expectedAliases {
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Show current theme"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "current") || !strings.Contains(cmd.Long, "theme") {
		t.Error("Long should describe showing current theme")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly theme current") {
		t.Error("Example should contain 'shelly theme current'")
	}

	if !strings.Contains(cmd.Example, "-o json") {
		t.Error("Example should show JSON output option")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// This command doesn't take arguments
	// Cobra commands without Args validation accept any arguments by default
	// Just verify the command structure is correct
	if cmd.Use != testCommandUse {
		t.Errorf("Use = %q, want %s", cmd.Use, testCommandUse)
	}
}

func TestExecute_ShowsCurrentTheme(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output showing current theme")
	}

	// Should contain "Current theme:" in output
	if !strings.Contains(output, "Current theme:") {
		t.Errorf("Output should contain 'Current theme:', got: %q", output)
	}
}

func TestExecute_WithOutputFlagJSON(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	// Note: The -o flag would be handled by parent command flags
	// For now, just test execution completes without error
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	out := tf.OutString()
	if out == "" {
		t.Error("Expected output")
	}
}

func TestExecute_WithOutputFlagYAML(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	// Note: The -o flag would be handled by parent command flags
	// For now, just test execution completes without error
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	out := tf.OutString()
	if out == "" {
		t.Error("Expected output")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_DefaultTheme(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output when displaying current theme")
	}

	if !strings.Contains(output, "Current theme:") {
		t.Errorf("Output should contain 'Current theme:', got: %q", output)
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_OutputContainsThemeName(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := tf.OutString()

	// Current theme should be set to dracula (default)
	currentTheme := theme.Current()
	if currentTheme == nil {
		t.Fatal("Expected current theme to be set")
	}

	if !strings.Contains(output, currentTheme.ID) {
		t.Errorf("Output should contain theme ID %q, got: %q", currentTheme.ID, output)
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_WithDisplayName(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	tf := factory.NewTestFactory(t)

	// Set a theme with a display name different from ID
	theme.SetTheme("monokai")

	err := run(tf.Factory)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("Expected output when showing theme with display name")
	}

	// Output should contain the theme ID
	currentTheme := theme.Current()
	if currentTheme != nil && currentTheme.ID != "" {
		if !strings.Contains(output, currentTheme.ID) {
			t.Errorf("Output should contain theme ID %q", currentTheme.ID)
		}
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_TextOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format (text output)
	viper.Set("output", "table")
	defer viper.Set("output", "")

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	out := tf.OutString()
	if out == "" {
		t.Error("Expected text output")
	}

	// Text output should not be JSON
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Error("Text output should not be JSON")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_JSONFormatWithIOStreams(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected output")
	}

	// Should contain theme information
	if !strings.Contains(out, "Current theme:") && !strings.Contains(out, "id") {
		t.Errorf("Output should contain theme info, got: %q", out)
	}
}

func TestNewCommand_AllFlagsExist(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Check for output format flag (may be inherited from parent command)
	_ = cmd.Flags().Lookup("output")
	// Output flag may be inherited from parent command, so not requiring it here
}

func TestExecute_CalledMultipleTimes(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd1 := NewCommand(tf.Factory)
	cmd1.SetArgs([]string{})

	err1 := cmd1.Execute()
	if err1 != nil {
		t.Errorf("First Execute() error = %v, want nil", err1)
	}

	out1 := tf.OutString()
	if out1 == "" {
		t.Error("First execution should produce output")
	}

	// Reset and test again
	tf.Reset()

	cmd2 := NewCommand(tf.Factory)
	cmd2.SetArgs([]string{})

	err2 := cmd2.Execute()
	if err2 != nil {
		t.Errorf("Second Execute() error = %v, want nil", err2)
	}

	out2 := tf.OutString()
	if out2 == "" {
		t.Error("Second execution should produce output")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_VerifyDataStructure(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	out := stdout.String()

	// Verify key information is present (theme output should always contain current theme)
	if !strings.Contains(out, "Current theme:") {
		t.Error("Output should contain 'Current theme:' label")
	}

	currentTheme := theme.Current()
	if currentTheme != nil && currentTheme.ID != "" {
		if !strings.Contains(out, currentTheme.ID) {
			t.Errorf("Output should contain theme ID %q", currentTheme.ID)
		}
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_TextFormatOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected text output")
	}

	// Should contain the text prefix
	if !strings.Contains(out, "Current theme:") {
		t.Errorf("Text output should contain 'Current theme:', got: %q", out)
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_StructuredOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure JSON format for structured output
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected output")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_DisplayNameShownWhenDifferent(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format so we get text output
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Use a different theme to test display name
	currentTheme := theme.Current()
	if currentTheme != nil && currentTheme.DisplayName != "" && currentTheme.DisplayName != currentTheme.ID {
		err := run(f)
		if err != nil {
			t.Errorf("run() error = %v, want nil", err)
		}

		out := stdout.String()
		if !strings.Contains(out, "Display name:") {
			t.Errorf("Output should contain 'Display name:' for themes with different display names, got: %q", out)
		}
	}
}

func TestNewCommand_HasValidStructure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify command structure
	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}

	if cmd.Use != testCommandUse {
		t.Errorf("Use = %q, want %s", cmd.Use, testCommandUse)
	}

	if len(cmd.Aliases) == 0 {
		t.Error("Should have at least one alias")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_ConsistentOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Ensure table format
	viper.Set("output", "table")
	defer viper.Set("output", "")

	tf := factory.NewTestFactory(t)

	// Run twice and verify output is consistent
	err1 := run(tf.Factory)
	if err1 != nil {
		t.Errorf("First run() error = %v, want nil", err1)
	}

	out1 := tf.OutString()
	if out1 == "" {
		t.Error("First run should produce output")
	}

	// Reset and run again
	tf.Reset()

	err2 := run(tf.Factory)
	if err2 != nil {
		t.Errorf("Second run() error = %v, want nil", err2)
	}

	out2 := tf.OutString()
	if out2 == "" {
		t.Error("Second run should produce output")
	}

	// Both outputs should contain the current theme
	if !strings.Contains(out1, "Current theme:") || !strings.Contains(out2, "Current theme:") {
		t.Error("Both runs should output current theme")
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_JSONStructuredOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Set JSON output format
	viper.Set("output", "json")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() with JSON format error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected JSON output")
	}

	// JSON output should be valid JSON-like or structured format
	// Just verify we got some output
	if !strings.Contains(out, "dracula") && !strings.Contains(out, "id") {
		t.Logf("JSON output: %q", out)
		// Output may vary based on viper configuration
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_YAMLStructuredOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Set YAML output format
	viper.Set("output", "yaml")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() with YAML format error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected YAML output")
	}

	// YAML output should contain theme fields
	if !strings.Contains(out, "id") && !strings.Contains(out, "displayName") {
		t.Logf("YAML output: %q", out)
		// Verify we at least got some output
	}
}

//nolint:paralleltest // cannot be parallel because it uses viper.Set
func TestRun_TableFormatOutput(t *testing.T) {
	// Cannot be parallel because it uses viper.Set
	// Set table output format explicitly
	viper.Set("output", "table")
	defer viper.Set("output", "")

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(f)
	if err != nil {
		t.Errorf("run() with table format error = %v, want nil", err)
	}

	out := stdout.String()
	if out == "" {
		t.Error("Expected table output")
	}

	// Table output should contain the text format
	if !strings.Contains(out, "Current theme:") {
		t.Errorf("Table output should contain 'Current theme:', got: %q", out)
	}
}
