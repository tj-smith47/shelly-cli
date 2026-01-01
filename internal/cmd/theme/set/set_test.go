package set

import (
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "set <theme>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "set <theme>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Command has no aliases")
	}
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "use" || alias == "s" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected alias 'use' or 's', got %v", cmd.Aliases)
	}

	// Verify example
	if cmd.Example == "" {
		t.Error("Command has no example")
	}

	// Verify --save flag exists
	saveFlag := cmd.Flags().Lookup("save")
	if saveFlag == nil {
		t.Error("--save flag not found")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Help may go to stdout or use command's built-in help output
	// Cobra help output typically goes through the command's writer
	output := tf.OutString() + tf.ErrString()

	// Just verify the command executed without error - help output location varies
	if output == "" {
		// Try getting from the command directly
		helpText := cmd.Short + cmd.Long
		if helpText == "" {
			t.Error("Command should have help text")
		}
	}
}

func TestNewCommand_RequiresArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no argument provided")
	}
}

func TestRun_ValidTheme(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"dracula"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Theme set to 'dracula'") {
		t.Errorf("Output = %q, want to contain \"Theme set to 'dracula'\"", output)
	}

	// Verify theme was actually set
	if theme.CurrentThemeName() != "dracula" {
		t.Errorf("CurrentThemeName() = %q, want %q", theme.CurrentThemeName(), "dracula")
	}
}

func TestRun_ValidTheme_Nord(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nord"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Theme set to 'nord'") {
		t.Errorf("Output = %q, want to contain \"Theme set to 'nord'\"", output)
	}
}

func TestRun_InvalidTheme(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent-theme"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid theme")
	}

	if !strings.Contains(err.Error(), "theme not found") {
		t.Errorf("Error = %q, want to contain 'theme not found'", err.Error())
	}
	if !strings.Contains(err.Error(), "nonexistent-theme") {
		t.Errorf("Error = %q, want to contain theme name", err.Error())
	}
	if !strings.Contains(err.Error(), "shelly theme list") {
		t.Errorf("Error = %q, should suggest 'shelly theme list'", err.Error())
	}
}

//nolint:paralleltest // SetupTestFs uses t.Setenv which is incompatible with t.Parallel()
func TestRun_WithSaveFlag(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create config directory
	if err := memFs.MkdirAll("/testconfig/shelly", 0o700); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"dracula", "--save"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Theme set to 'dracula'") {
		t.Errorf("Output = %q, want to contain 'Theme set to'", output)
	}
	if !strings.Contains(output, "saved to config") {
		t.Errorf("Output = %q, want to contain 'saved to config'", output)
	}
}

//nolint:paralleltest // SetupTestFs uses t.Setenv which is incompatible with t.Parallel()
func TestRun_SaveFlag_DifferentTheme(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	// Create config directory
	if err := memFs.MkdirAll("/testconfig/shelly", 0o700); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nord", "--save"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Theme set to 'nord'") {
		t.Errorf("Output = %q, want to contain theme name", output)
	}
	if !strings.Contains(output, "saved to config") {
		t.Errorf("Output = %q, want to contain 'saved to config'", output)
	}
}

func TestRun_InvalidTheme_WithSaveFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"nonexistent", "--save"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid theme with --save")
	}

	// Should fail at theme validation, not at save
	if !strings.Contains(err.Error(), "theme not found") {
		t.Errorf("Error = %q, want to contain 'theme not found'", err.Error())
	}
}

func TestOptions_Structure(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Factory:   nil,
		Save:      true,
		ThemeName: "dracula",
	}

	if opts.ThemeName != "dracula" {
		t.Errorf("ThemeName = %q, want %q", opts.ThemeName, "dracula")
	}
	if !opts.Save {
		t.Error("Save should be true")
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for tab completion")
	}
}

func TestRun_MultipleThemes(t *testing.T) {
	t.Parallel()

	themes := []string{"dracula", "nord", "gruvbox", "solarized-dark", "monokai"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			t.Parallel()

			// Check if theme exists first
			if _, ok := theme.GetTheme(themeName); !ok {
				t.Skipf("Theme %q not available in registry", themeName)
			}

			tf := factory.NewTestFactory(t)
			cmd := NewCommand(tf.Factory)
			cmd.SetArgs([]string{themeName})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			output := tf.OutString()
			if !strings.Contains(output, "Theme set to '"+themeName+"'") {
				t.Errorf("Output = %q, want to contain theme name", output)
			}
		})
	}
}

func TestRun_SaveFlag_WriteError(t *testing.T) {
	// Cannot use t.Parallel() because we modify global config.Fs state
	// Create a read-only filesystem to trigger write error
	memFs := afero.NewMemMapFs()
	// Create config directory first
	if err := memFs.MkdirAll("/testconfig/shelly", 0o700); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	// Wrap in read-only filesystem to cause write failures
	roFs := afero.NewReadOnlyFs(memFs)
	config.SetFs(roFs)

	t.Setenv("XDG_CONFIG_HOME", "/testconfig")
	config.ResetDefaultManagerForTesting()

	t.Cleanup(func() {
		config.SetFs(nil)
		config.ResetDefaultManagerForTesting()
	})

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"dracula", "--save"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when filesystem is read-only")
	}

	if !strings.Contains(err.Error(), "failed to save theme") {
		t.Errorf("Error = %q, want to contain 'failed to save theme'", err.Error())
	}
}
