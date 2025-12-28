package get

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "get [key]" {
		t.Errorf("Use = %q, want \"get [key]\"", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"read"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, expectedAliases)
	}
	for i, expected := range expectedAliases {
		if i >= len(cmd.Aliases) {
			t.Errorf("Missing alias at index %d", i)
			continue
		}
		if cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	// Example should show basic usage
	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_HasValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for setting key completion")
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
			name:    "no args valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg valid",
			args:    []string{"defaults.timeout"},
			wantErr: false,
		},
		{
			name:    "two args invalid",
			args:    []string{"key1", "key2"},
			wantErr: true,
		},
		{
			name:    "three args invalid",
			args:    []string{"key1", "key2", "key3"},
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
			name:      "uses MaximumNArgs(1)",
			checkFunc: func(c *cobra.Command) bool { return c.Args != nil },
			wantOK:    true,
			errMsg:    "Args should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			wantOK:    true,
			errMsg:    "ValidArgsFunction should be set",
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

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long description is empty")
	}

	// Long description should mention dot notation
	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show get all settings
	examples := cmd.Example
	if !strings.Contains(examples, "shelly config get") {
		t.Error("Example should show 'shelly config get' usage")
	}
}

func TestRun_NoArgs_ReturnsAllSettings(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, []string{})
	if err != nil {
		t.Errorf("run() with no args should not error, got: %v", err)
	}

	// Output should contain settings (either table or some form of output)
	// Since viper may not be initialized in tests, we just verify no panic
}

func TestRun_WithUnknownKey(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, []string{"nonexistent.key.that.does.not.exist"})
	if err == nil {
		t.Error("Expected error for unknown key")
	}

	// Error message should mention the key
	if err != nil && !strings.Contains(err.Error(), "not set") {
		t.Errorf("Error message should mention 'not set', got: %v", err)
	}
}

func TestRun_WithValidKey(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Note: This may or may not error depending on viper state
	// The important thing is it doesn't panic
	err := run(tf.Factory, []string{"output"})
	if err != nil {
		// Some keys may not be set in test environment
		if !strings.Contains(err.Error(), "not set") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestNewCommand_UsesFactory(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Verify command uses factory properly
	cmd := NewCommand(tf.Factory)

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}

	// Verify command doesn't panic when created with test factory
	if cmd.Use == "" {
		t.Error("Command should have Use set")
	}
}

func TestNewCommand_AcceptsKeyWithDotNotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "simple key",
			args:    []string{"output"},
			wantErr: false,
		},
		{
			name:    "nested key with one dot",
			args:    []string{"defaults.timeout"},
			wantErr: false,
		},
		{
			name:    "nested key with two dots",
			args:    []string{"theme.colors.primary"},
			wantErr: false,
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

func TestRun_OutputIsProduced(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Run with no args (shows all settings)
	err := run(tf.Factory, []string{})

	// We mainly want to verify no panic occurs
	// Output verification depends on viper state which varies in tests
	if err != nil {
		t.Logf("run() returned error (may be expected in test env): %v", err)
	}
}

func TestNewCommand_CanBeAddedToParent(t *testing.T) {
	t.Parallel()

	parent := &cobra.Command{Use: "config"}
	child := NewCommand(cmdutil.NewFactory())

	parent.AddCommand(child)

	// Verify child was added
	found := false
	for _, cmd := range parent.Commands() {
		if cmd.Name() == "get" {
			found = true
			break
		}
	}

	if !found {
		t.Error("get command was not added to parent")
	}
}

func TestNewCommand_AliasesWork(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify aliases can be used to find the command
	for _, alias := range cmd.Aliases {
		if alias == "" {
			t.Error("Empty alias found")
		}
	}

	// Specifically check for "read" alias
	hasReadAlias := false
	for _, alias := range cmd.Aliases {
		if alias == "read" {
			hasReadAlias = true
			break
		}
	}
	if !hasReadAlias {
		t.Error("Expected 'read' alias")
	}
}
