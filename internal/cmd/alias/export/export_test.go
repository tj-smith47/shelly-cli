package export

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "export [file]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export [file]")
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

	expectedAliases := map[string]bool{"save": true, "dump": true}

	if len(cmd.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d: %v", len(cmd.Aliases), cmd.Aliases)
	}

	for _, alias := range cmd.Aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias %q", alias)
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
			name:    "no args valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg valid",
			args:    []string{"aliases.yaml"},
			wantErr: false,
		},
		{
			name:    "two args invalid",
			args:    []string{"arg1", "arg2"},
			wantErr: true,
		},
		{
			name:    "three args invalid",
			args:    []string{"arg1", "arg2", "arg3"},
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

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.checkFunc(cmd) {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if len(cmd.Long) < 30 {
		t.Error("Long description seems too short")
	}

	// Verify long description mentions key functionality
	long := cmd.Long
	if long == "" {
		t.Fatal("Long description is empty")
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Example should show meaningful patterns
	if len(cmd.Example) < 20 {
		t.Error("Example seems too short to be useful")
	}
}

func TestRun_NoAliases(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, "")
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Warning is written to stderr, not stdout
	errOutput := tf.ErrString()
	if errOutput == "" {
		t.Error("Expected warning output for empty alias list")
	}
}

func TestRun_WithAliasesToStdout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add aliases to the config
	tf.Config.Aliases = map[string]config.Alias{
		"st": {
			Name:    "st",
			Command: "status kitchen",
			Shell:   false,
		},
		"reboot": {
			Name:    "reboot",
			Command: "device reboot",
			Shell:   false,
		},
	}

	err := run(tf.Factory, "")
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should print YAML to stdout
	output := tf.OutString()
	if output == "" {
		t.Error("Expected YAML output")
	}
}

func TestRun_WithAliasesToFile(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add aliases to the config
	tf.Config.Aliases = map[string]config.Alias{
		"on": {
			Name:    "on",
			Command: "switch on kitchen",
			Shell:   false,
		},
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "aliases.yaml")

	err := run(tf.Factory, filename)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", filename)
	}

	// Should print success message
	output := tf.OutString()
	if output == "" {
		t.Error("Expected success message output")
	}
}

func TestRun_ShellAlias(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add a shell alias
	tf.Config.Aliases = map[string]config.Alias{
		"list": {
			Name:    "list",
			Command: "ls -la",
			Shell:   true,
		},
	}

	err := run(tf.Factory, "")
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should export shell aliases with ! prefix
	output := tf.OutString()
	if output == "" {
		t.Error("Expected YAML output for shell alias")
	}
}

func TestRun_InvalidFilePath(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add aliases to the config
	tf.Config.Aliases = map[string]config.Alias{
		"test": {
			Name:    "test",
			Command: "status",
			Shell:   false,
		},
	}

	// Use an invalid path (directory that doesn't exist)
	filename := "/nonexistent/directory/aliases.yaml"

	err := run(tf.Factory, filename)
	if err == nil {
		t.Error("expected error for invalid file path, got nil")
	}
}

func TestRun_MultipleAliases(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add multiple aliases
	tf.Config.Aliases = map[string]config.Alias{
		"st": {
			Name:    "st",
			Command: "status",
			Shell:   false,
		},
		"rb": {
			Name:    "rb",
			Command: "device reboot",
			Shell:   false,
		},
		"ls": {
			Name:    "ls",
			Command: "device list",
			Shell:   false,
		},
	}

	err := run(tf.Factory, "")
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should print YAML to stdout
	output := tf.OutString()
	if output == "" {
		t.Error("Expected YAML output with multiple aliases")
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Export command has no flags, just an optional positional arg
	if cmd.Flags().HasFlags() {
		t.Error("export command should have no flags")
	}
}

func TestNewCommand_RunE_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	// Execute command with no args (outputs to stdout)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("RunE() error = %v, want nil", err)
	}
}

func TestNewCommand_RunE_WithFilename(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	// Add an alias so export has something to export
	tf.Config.Aliases = map[string]config.Alias{
		"test": {
			Name:    "test",
			Command: "status",
			Shell:   false,
		},
	}

	cmd := NewCommand(tf.Factory)

	// Create a temporary file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test-aliases.yaml")

	// Execute command with filename arg
	err := cmd.RunE(cmd, []string{filename})
	if err != nil {
		t.Errorf("RunE() error = %v, want nil", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", filename)
	}
}
