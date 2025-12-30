package importcmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import <file>")
	}

	// Test Aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}

	// Test that 'load' is in the aliases
	found := false
	for _, alias := range cmd.Aliases {
		if alias == "load" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'load' alias, got %v", cmd.Aliases)
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
		{"one arg valid", []string{"theme.yaml"}, false},
		{"two args", []string{"file1", "file2"}, true},
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

	// Test apply flag
	flag := cmd.Flags().Lookup("apply")
	if flag == nil {
		t.Fatal("--apply flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("--apply default = %q, want %q", flag.DefValue, "false")
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

func TestExecute_ValidThemeFileWithoutApply(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with valid name
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: dracula
colors:
  foreground: "#f8f8f2"
  background: "#282a36"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "validated") {
		t.Errorf("expected 'validated' in output, got %q", output)
	}
}

func TestExecute_ValidThemeFileWithApply(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with valid name
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: dracula
colors:
  foreground: "#f8f8f2"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile, "--apply"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "imported") {
		t.Errorf("expected 'imported' in output, got %q", output)
	}
}

func TestExecute_ValidThemeWithColorOverrides(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with custom colors
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "custom.yaml")
	content := `name: dracula
colors:
  foreground: "#ffffff"
  background: "#000000"
  green: "#00ff00"
  red: "#ff0000"
  blue: "#0000ff"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "5") {
		// Should mention 5 color overrides
		t.Logf("expected color override count in output, got %q", output)
	}
}

func TestExecute_InvalidThemeNotBuiltIn(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with non-existent theme
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "invalid.yaml")
	content := `name: nonexistent-theme-xyz
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for non-existent theme")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}
}

func TestExecute_FileNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"/nonexistent/path/theme.yaml"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' in error, got %q", err.Error())
	}
}

func TestExecute_InvalidYAML(t *testing.T) {
	t.Parallel()

	// Create a temporary file with invalid YAML
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "invalid.yaml")
	content := `{invalid: yaml: [structure`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("expected 'failed to parse' in error, got %q", err.Error())
	}
}

func TestExecute_MissingThemeAndColors(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with neither name nor colors
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "empty.yaml")
	content := `# Empty theme file with no name or colors
id: ""
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for missing name and colors")
	}

	if !strings.Contains(err.Error(), "invalid theme file") {
		t.Errorf("expected 'invalid theme file' in error, got %q", err.Error())
	}
}

func TestExecute_OldFormatWithID(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file using old format (id field instead of name)
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "old_format.yaml")
	content := `id: dracula
display_name: "Dracula Theme"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "validated") {
		t.Errorf("expected 'validated' in output, got %q", output)
	}
}

func TestExecute_ColorOnlyTheme(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with only colors, no theme name
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "colors_only.yaml")
	content := `colors:
  foreground: "#aabbcc"
  background: "#112233"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "validated") {
		t.Errorf("expected 'validated' in output, got %q", output)
	}
}

func TestExecute_ApplyColorOnlyTheme(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file with only colors and apply flag
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "colors_apply.yaml")
	content := `colors:
  green: "#00ff00"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile, "--apply"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "imported") {
		t.Errorf("expected 'imported' in output, got %q", output)
	}
}

func TestRun_ValidThemeWithoutApply(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: dracula
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, false)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "validated") {
		t.Errorf("expected 'validated' in output, got %q", output)
	}
}

func TestRun_ValidThemeWithApply(t *testing.T) {
	t.Parallel()

	// Create a temporary theme file
	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: dracula
colors:
  foreground: "#fff"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, true)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "imported") {
		t.Errorf("expected 'imported' in output, got %q", output)
	}
}

func TestRun_FileNotFound(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, "/nonexistent/file.yaml", false)
	if err == nil {
		t.Error("run() expected error for non-existent file")
	}
}

func TestRun_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "bad.yaml")
	content := `[invalid: yaml`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, false)
	if err == nil {
		t.Error("run() expected error for invalid YAML")
	}
}

func TestRun_InvalidThemeName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: this-theme-does-not-exist-xyz
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, false)
	if err == nil {
		t.Error("run() expected error for invalid theme name")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got %q", err.Error())
	}
}

func TestRun_MissingNameAndColors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "empty.yaml")
	content := `# No name or colors
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, false)
	if err == nil {
		t.Error("run() expected error for missing name and colors")
	}

	if !strings.Contains(err.Error(), "invalid theme file") {
		t.Errorf("expected 'invalid theme file' in error, got %q", err.Error())
	}
}

func TestRun_WithColorOverrides(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "colors.yaml")
	content := `name: dracula
colors:
  foreground: "#ffffff"
  background: "#000000"
  green: "#00ff00"
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	err := run(tf.Factory, themeFile, false)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	// Should mention color overrides
	if !strings.Contains(output, "Color overrides") && !strings.Contains(output, "validated") {
		t.Logf("expected color info in output, got %q", output)
	}
}

func TestExecute_ApplyWithInvalidTheme(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "invalid.yaml")
	content := `name: this-does-not-exist
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile, "--apply"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for apply with invalid theme")
	}
}

func TestExecute_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "empty.yaml")
	// Create empty file
	if err := os.WriteFile(themeFile, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() expected error for empty file")
	}

	if !strings.Contains(err.Error(), "invalid theme file") {
		t.Errorf("expected 'invalid theme file' in error, got %q", err.Error())
	}
}

func TestExecute_AlternateAlias(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	themeFile := filepath.Join(tmpDir, "theme.yaml")
	content := `name: dracula
`
	if err := os.WriteFile(themeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tf := factory.NewTestFactory(t)
	// Create the command through a parent that would use alias
	cmd := NewCommand(tf.Factory)

	// Verify alias exists
	aliasFound := false
	for _, alias := range cmd.Aliases {
		if alias == "load" {
			aliasFound = true
			break
		}
	}
	if !aliasFound {
		t.Fatal("'load' alias not found")
	}

	// Test with the command directly
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{themeFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
