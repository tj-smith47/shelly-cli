package config

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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

func TestNewCommand_UseField(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "config" {
		t.Errorf("Use = %q, want \"config\"", cmd.Use)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"params", "parameters"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("missing alias[%d] = %q", i, want)
			continue
		}
		if cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if !strings.Contains(cmd.Short, "configuration") {
		t.Error("Short should mention configuration")
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "Z-Wave") {
		t.Error("Long should mention Z-Wave")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly zwave config") {
		t.Error("Example should contain 'shelly zwave config'")
	}

	if !strings.Contains(cmd.Example, "-o json") {
		t.Error("Example should contain '-o json'")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept no arguments
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("expected no error with 0 args, got: %v", err)
	}

	// Should reject any arguments
	err = cmd.Args(cmd, []string{"arg1"})
	if err == nil {
		t.Error("expected error with 1 arg")
	}

	err = cmd.Args(cmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is not set")
	}
}

func TestNewCommand_OutputFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}

	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want \"o\"", outputFlag.Shorthand)
	}

	if outputFlag.DefValue != "table" {
		t.Errorf("output default = %q, want \"table\"", outputFlag.DefValue)
	}
}

func TestNewCommand_CanBeAddedToParent(t *testing.T) {
	t.Parallel()

	parent := &cobra.Command{Use: "zwave"}
	child := NewCommand(cmdutil.NewFactory())

	parent.AddCommand(child)

	found := false
	for _, cmd := range parent.Commands() {
		if cmd.Name() == "config" {
			found = true
			break
		}
	}

	if !found {
		t.Error("config command was not added to parent")
	}
}

func TestOptions_Factory(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Factory: f,
	}

	if opts.Factory == nil {
		t.Error("Factory is nil")
	}

	if opts.Factory.IOStreams() == nil {
		t.Error("Factory.IOStreams() returned nil")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output from command")
	}
}

func TestExecute_WithFixtures(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
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
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output, got empty string")
	}

	if !strings.Contains(output, "Parameter") {
		t.Error("output should contain 'Parameter'")
	}
}

func TestExecute_OutputTable(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", "table"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output, got empty string")
	}

	if !strings.Contains(output, "Parameter") {
		t.Error("table output should contain parameter information")
	}
}

func TestExecute_OutputJSON(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", "json"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output, got empty string")
	}

	// JSON output should contain brackets
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("JSON output should contain array brackets")
	}
}

func TestExecute_OutputYAML(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", "yaml"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output, got empty string")
	}

	// YAML output should contain some structure
	if len(output) < 10 {
		t.Error("YAML output seems too short")
	}
}

func TestNewCommand_HelpText(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("command should have Long help text")
	}

	if cmd.Short == "" {
		t.Error("command should have Short help text")
	}

	if cmd.Example == "" {
		t.Error("command should have Example text")
	}

	// Verify help text mentions Z-Wave
	if !strings.Contains(cmd.Long, "Z-Wave") {
		t.Error("Long help should mention Z-Wave")
	}
}

func TestExecute_InvalidFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--invalid-flag"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestRun_DirectCall(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()
	if output == "" {
		t.Error("expected output from run")
	}

	if !strings.Contains(output, "Parameters") {
		t.Error("output should mention Parameters")
	}
}

func TestRun_OutputContainsParameterInfo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	output := tf.OutString()

	// Should contain numbers for parameter IDs
	if !strings.Contains(output, "[") {
		t.Error("output should contain parameter number indicators")
	}

	// Should contain the warning about device models
	if !strings.Contains(output, "device model") {
		t.Error("output should contain device model warning")
	}
}

func TestExecute_RejectsArguments(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"extra-arg"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when providing arguments")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			errMsg:    "Use should not be empty",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			errMsg:    "RunE should be set",
		},
		{
			name:      "has Args",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			errMsg:    "Args should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc(cmd) {
				t.Error(tt.errMsg)
			}
		})
	}
}
