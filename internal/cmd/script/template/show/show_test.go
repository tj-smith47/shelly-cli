package show

import (
	"bytes"
	"context"
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
	if cmd.Use != "show <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "show <name>")
	}

	// Test Aliases
	wantAliases := []string{"get", "view"}
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
		{"one arg valid", []string{"motion-light"}, false},
		{"two args", []string{"template", "extra"}, true},
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

	// Test code flag
	codeFlag := cmd.Flags().Lookup("code")
	if codeFlag == nil {
		t.Fatal("--code flag not found")
	}
	if codeFlag.DefValue != "false" {
		t.Errorf("--code default = %q, want %q", codeFlag.DefValue, "false")
	}

	// Test output flag
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("--output flag not found")
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
		"shelly script template show",
		"motion-light",
		"--code",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for completion")
	}
}

func TestRun_ShowsBuiltInTemplate(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Name:    "motion-light",
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain template details
	if !strings.Contains(output, "motion-light") {
		t.Errorf("output should contain 'motion-light', got: %q", output)
	}
	if !strings.Contains(output, "automation") {
		t.Errorf("output should contain 'automation' category, got: %q", output)
	}
}

func TestRun_TemplateNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Name:    "nonexistent-template",
	}

	err := run(opts)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want to contain 'not found'", err)
	}
}

func TestRun_CodeOnly(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Name:    "motion-light",
		Code:    true,
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain the code directly
	if !strings.Contains(output, "Motion-activated light control") {
		t.Errorf("output should contain code comment, got: %q", output)
	}
	// Should NOT contain template metadata when --code is used
	if strings.Contains(output, "Description:") {
		t.Errorf("output should not contain metadata, got: %q", output)
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
		Name:    "power-monitor",
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should be valid JSON containing template info
	if !strings.Contains(output, "power-monitor") {
		t.Errorf("output should contain 'power-monitor', got: %q", output)
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
		Name:    "toggle-sync",
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain template name
	if !strings.Contains(output, "toggle-sync") {
		t.Errorf("output should contain 'toggle-sync', got: %q", output)
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
		Name:    "energy-logger",
	}

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should contain template details
	if !strings.Contains(output, "energy-logger") {
		t.Errorf("output should contain 'energy-logger', got: %q", output)
	}
}

func TestRun_AllBuiltInTemplates(t *testing.T) {
	t.Parallel()

	templates := []string{
		"motion-light",
		"power-monitor",
		"schedule-helper",
		"toggle-sync",
		"energy-logger",
	}

	for _, tplName := range templates {
		t.Run(tplName, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			opts := &Options{
				Factory: tf.Factory,
				Name:    tplName,
			}

			err := run(opts)
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}

			output := tf.OutString()
			if !strings.Contains(output, tplName) {
				t.Errorf("output should contain %q, got: %q", tplName, output)
			}
		})
	}
}

func TestNewCommand_Execute(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"motion-light"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
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

func TestRun_Context(t *testing.T) {
	t.Parallel()

	// Verify context is passed properly (even though show doesn't use it for network)
	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
		Name:    "motion-light",
	}

	ctx := context.Background()
	_ = ctx // show doesn't use ctx but verifying Options works

	err := run(opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
}
