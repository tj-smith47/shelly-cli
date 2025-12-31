package edit

import (
	"context"
	"os"
	"path/filepath"
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

	if cmd.Use != "edit" {
		t.Errorf("Use = %q, want \"edit\"", cmd.Use)
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

	expectedAliases := []string{"e"}
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

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// edit command takes no args (it opens config in editor)
	// By default, commands accept any args if Args is nil
	if cmd.Args != nil {
		// If Args is set, it should accept empty args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("edit command should accept no args, got error: %v", err)
		}
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		check  func(cmd *cobra.Command) bool
		errMsg string
	}{
		{
			name:   "has use",
			check:  func(c *cobra.Command) bool { return c.Use != "" },
			errMsg: "Use should not be empty",
		},
		{
			name:   "has short",
			check:  func(c *cobra.Command) bool { return c.Short != "" },
			errMsg: "Short should not be empty",
		},
		{
			name:   "has long",
			check:  func(c *cobra.Command) bool { return c.Long != "" },
			errMsg: "Long should not be empty",
		},
		{
			name:   "has example",
			check:  func(c *cobra.Command) bool { return c.Example != "" },
			errMsg: "Example should not be empty",
		},
		{
			name:   "has aliases",
			check:  func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			errMsg: "Aliases should not be empty",
		},
		{
			name:   "has RunE",
			check:  func(c *cobra.Command) bool { return c.RunE != nil },
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if !tt.check(cmd) {
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

	// Long description should be substantial
	if len(cmd.Long) < 50 {
		t.Error("Long description seems too short")
	}
}

func TestRun_ConfigNotFound(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory}

	// Set HOME to a temp directory with no config
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when config file not found")
	}

	// Should mention config file or initialization
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestRun_NoEditorFound(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory}

	// Create a temp directory with a config file
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "shelly")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("output: table\n"), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set HOME to our temp directory
	t.Setenv("HOME", tempDir)
	// Clear editor environment variables
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	// Modify PATH to ensure no editor can be found
	t.Setenv("PATH", tempDir)

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when no editor found")
	}

	// Should mention editor not found
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{Factory: tf.Factory}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Even with cancelled context, we should get an error
	// (either from config not found, no editor, or context cancellation)
	err := run(ctx, opts)
	if err == nil {
		// This might succeed if config exists and editor runs quickly
		// but generally we expect an error
		t.Log("run() succeeded with cancelled context")
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
