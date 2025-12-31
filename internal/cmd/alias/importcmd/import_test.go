package importcmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

	if cmd.Use != "import <file>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import <file>")
	}

	wantAliases := []string{"load"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	}
	for i, alias := range wantAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if cmd.RunE == nil {
		t.Error("RunE is nil")
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
		{"one arg valid", []string{"aliases.yaml"}, false},
		{"two args invalid", []string{"aliases.yaml", "extra"}, true},
		{"three args invalid", []string{"aliases.yaml", "extra", "more"}, true},
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

	flag := cmd.Flags().Lookup("merge")
	if flag == nil {
		t.Fatal("--merge flag not found")
	}
	if flag.Shorthand != "m" {
		t.Errorf("--merge shorthand = %q, want %q", flag.Shorthand, "m")
	}
	if flag.DefValue != "false" {
		t.Errorf("--merge default = %q, want %q", flag.DefValue, "false")
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
		"shelly alias import",
		".yaml",
		"--merge",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Long description should mention key aspects
	wantPatterns := []string{
		"aliases",
		"YAML",
		"--merge",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(strings.ToLower(cmd.Long), strings.ToLower(pattern)) {
			t.Errorf("Long description should mention %q", pattern)
		}
	}
}

// setupTestConfigEnv sets up an isolated config environment for testing.
// Returns a cleanup function that should be called via t.Cleanup().
func setupTestConfigEnv(t *testing.T) string {
	t.Helper()

	// Use in-memory filesystem for config operations
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	// Reset the singleton before changing HOME
	config.ResetDefaultManagerForTesting()

	tmpDir := t.TempDir()

	// Register cleanup to reset singleton after test
	t.Cleanup(config.ResetDefaultManagerForTesting)

	return tmpDir
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_Success(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a valid aliases YAML file
	aliasFile := filepath.Join(tmpDir, "aliases.yaml")
	content := `aliases:
  st: "status kitchen"
  rb: "device reboot"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, false)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Check success output
	output := tf.OutString()
	if !strings.Contains(output, "Imported 2 alias") {
		t.Errorf("expected success message with 2 aliases, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_FileNotFound(t *testing.T) {
	setupTestConfigEnv(t)

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, "/nonexistent/path/to/aliases.yaml", false)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to import aliases") {
		t.Errorf("error should mention 'failed to import aliases': %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_InvalidYAML(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create an invalid YAML file
	aliasFile := filepath.Join(tmpDir, "invalid.yaml")
	content := `this is not valid yaml: [unclosed bracket`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, false)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to import aliases") {
		t.Errorf("error should mention 'failed to import aliases': %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_ShellAliases(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a YAML file with shell aliases (prefixed with !)
	aliasFile := filepath.Join(tmpDir, "shell-aliases.yaml")
	content := `aliases:
  ls: "!ls -la"
  myecho: "!echo hello"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, false)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Check success output
	output := tf.OutString()
	if !strings.Contains(output, "Imported 2 alias") {
		t.Errorf("expected success message with 2 aliases, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_MergeMode(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// First, add an existing alias
	if err := config.AddAlias("st", "existing command", false); err != nil {
		t.Fatalf("failed to add existing alias: %v", err)
	}

	// Create a YAML file with overlapping and new aliases
	aliasFile := filepath.Join(tmpDir, "merge-aliases.yaml")
	content := `aliases:
  st: "status kitchen"
  newcmd: "device info"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, true)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Check output - Info() writes to stdout, not stderr
	output := tf.OutString()
	if !strings.Contains(output, "Skipped 1 existing alias") {
		t.Errorf("expected skip message in output, got: %s", output)
	}

	// Check output - should mention imported alias
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message with 1 alias, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_OverwriteMode(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// First, add an existing alias
	if err := config.AddAlias("st", "existing command", false); err != nil {
		t.Fatalf("failed to add existing alias: %v", err)
	}

	// Create a YAML file with overlapping alias
	aliasFile := filepath.Join(tmpDir, "overwrite-aliases.yaml")
	content := `aliases:
  st: "status kitchen"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	// merge=false means overwrite existing
	err := run(tf.Factory, aliasFile, false)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Check output - should import (overwrite)
	output := tf.OutString()
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message with 1 alias, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_InvalidAliasName(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a YAML file with invalid alias name (reserved command)
	aliasFile := filepath.Join(tmpDir, "invalid-name.yaml")
	content := `aliases:
  help: "some command"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, false)
	if err == nil {
		t.Fatal("expected error for reserved alias name")
	}

	if !strings.Contains(err.Error(), "failed to import aliases") {
		t.Errorf("error should mention 'failed to import aliases': %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestRun_EmptyFile(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create an empty YAML file (no aliases section)
	aliasFile := filepath.Join(tmpDir, "empty.yaml")
	content := `aliases:`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tf := factory.NewTestFactory(t)

	err := run(tf.Factory, aliasFile, false)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should succeed but import 0 aliases
	output := tf.OutString()
	if !strings.Contains(output, "Imported 0 alias") {
		t.Errorf("expected success message with 0 aliases, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_Success(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a valid aliases YAML file
	aliasFile := filepath.Join(tmpDir, "aliases.yaml")
	content := `aliases:
  st: "status kitchen"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success output
	output := out.String()
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_WithMergeFlag(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a valid aliases YAML file
	aliasFile := filepath.Join(tmpDir, "aliases.yaml")
	content := `aliases:
  newcmd: "device info"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile, "--merge"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success output
	output := out.String()
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_ShortMergeFlag(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a valid aliases YAML file
	aliasFile := filepath.Join(tmpDir, "aliases.yaml")
	content := `aliases:
  newcmd: "device info"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile, "-m"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success output
	output := out.String()
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_FileNotFound(t *testing.T) {
	setupTestConfigEnv(t)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"/nonexistent/path/aliases.yaml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to import aliases") {
		t.Errorf("error should mention 'failed to import aliases': %v", err)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_NoArgs(t *testing.T) {
	setupTestConfigEnv(t)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_MergeWithExisting(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Add an existing alias
	if err := config.AddAlias("existing", "existing command", false); err != nil {
		t.Fatalf("failed to add existing alias: %v", err)
	}

	// Create a YAML file with overlapping and new aliases
	aliasFile := filepath.Join(tmpDir, "merge-aliases.yaml")
	content := `aliases:
  existing: "new command for existing"
  brandnew: "brand new command"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile, "--merge"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check that skipped message is shown - Info() writes to stdout
	output := out.String()
	if !strings.Contains(output, "Skipped 1 existing alias") {
		t.Errorf("expected skip message in output, got: %s", output)
	}

	// Check that imported message is shown
	if !strings.Contains(output, "Imported 1 alias") {
		t.Errorf("expected success message with 1 alias, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_MultipleAliases(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a YAML file with multiple aliases
	aliasFile := filepath.Join(tmpDir, "multi-aliases.yaml")
	content := `aliases:
  st: "status kitchen"
  rb: "device reboot"
  ls: "device list"
  info: "device info"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success output with correct count
	output := out.String()
	if !strings.Contains(output, "Imported 4 alias") {
		t.Errorf("expected success message with 4 aliases, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_MixedAliases(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a YAML file with both regular and shell aliases
	aliasFile := filepath.Join(tmpDir, "mixed-aliases.yaml")
	content := `aliases:
  st: "status kitchen"
  myls: "!ls -la"
  rb: "device reboot"
  myecho: "!echo hello world"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success output with correct count
	output := out.String()
	if !strings.Contains(output, "Imported 4 alias") {
		t.Errorf("expected success message with 4 aliases, got: %s", output)
	}
}

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExecute_AliasWithWhitespace(t *testing.T) {
	tmpDir := setupTestConfigEnv(t)

	// Create a YAML file with alias name containing whitespace (should fail)
	aliasFile := filepath.Join(tmpDir, "whitespace-alias.yaml")
	content := `aliases:
  "my alias": "status kitchen"
`
	if err := os.WriteFile(aliasFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetContext(context.Background())
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{aliasFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for alias with whitespace")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"load"}

	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d: %v", len(expectedAliases), len(cmd.Aliases), cmd.Aliases)
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}
