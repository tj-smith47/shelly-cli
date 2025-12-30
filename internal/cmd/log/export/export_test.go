package export

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// setupTestConfigDir creates a temporary config directory and sets XDG_CONFIG_HOME.
// Returns the config directory path (inside the temp directory).
func setupTestConfigDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "shelly")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	return configDir
}

// createLogFile creates a log file in the given config directory with the specified content.
func createLogFile(t *testing.T, configDir, content string) {
	t.Helper()
	logPath := filepath.Join(configDir, "shelly.log")
	if err := os.WriteFile(logPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}
}

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

	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "export" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export")
	}

	// Test Aliases
	wantAliases := []string{"save", "backup"}
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
		{"no args", []string{}, false},
		{"one arg", []string{"extra"}, true},
		{"two args", []string{"arg1", "arg2"}, true},
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

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"output", "o", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
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
		"shelly log export",
		"-o",
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

	wantPatterns := []string{
		"log",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestExecute_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Export") {
		t.Error("Help output should contain 'Export'")
	}
	if !strings.Contains(output, "--output") {
		t.Error("Help output should contain '--output'")
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestExecute_NoLogFile(t *testing.T) {
	// Set up a temporary config dir with no log file
	_ = setupTestConfigDir(t)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() should not error for missing log file: %v", err)
	}

	// Should show info message about no log file
	output := tf.OutString()
	if !strings.Contains(output, "No log file found") {
		t.Errorf("Expected 'No log file found' message, got: %s", output)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestExecute_ExportToStdout(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "2024-01-01 10:00:00 INFO Test log line 1\n2024-01-01 10:01:00 INFO Test log line 2\n"
	createLogFile(t, configDir, logContent)

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Test log line 1") {
		t.Errorf("Expected output to contain log content, got: %s", output)
	}
	if !strings.Contains(output, "Test log line 2") {
		t.Errorf("Expected output to contain log content, got: %s", output)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestExecute_ExportToFile(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "2024-01-01 10:00:00 INFO Test export content\n"
	createLogFile(t, configDir, logContent)

	// Output file in temp directory (parent of configDir)
	outputFile := filepath.Join(filepath.Dir(configDir), "exported.log")

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", outputFile})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Check success message
	output := tf.OutString()
	if !strings.Contains(output, "exported") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify the file was created with correct content
	exported, err := os.ReadFile(outputFile) //nolint:gosec // Test file path from TempDir
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}
	if string(exported) != logContent {
		t.Errorf("Exported content = %q, want %q", string(exported), logContent)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestExecute_ExportToInvalidPath(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "2024-01-01 10:00:00 INFO Test content\n"
	createLogFile(t, configDir, logContent)

	// Try to write to an invalid path (directory that doesn't exist)
	tmpDir := filepath.Dir(configDir)
	invalidPath := filepath.Join(tmpDir, "nonexistent", "subdir", "export.log")

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-o", invalidPath})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error for invalid output path")
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_NoLogFile(t *testing.T) {
	// Set up a temporary config dir with no log file
	_ = setupTestConfigDir(t)

	tf := factory.NewTestFactory(t)
	opts := &Options{}

	err := run(tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "No log file found") {
		t.Errorf("Expected 'No log file found' message, got: %s", output)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_ExportToStdout(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "Line 1\nLine 2\nLine 3\n"
	createLogFile(t, configDir, logContent)

	tf := factory.NewTestFactory(t)
	opts := &Options{} // No output file

	err := run(tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "Line 1") || !strings.Contains(output, "Line 2") {
		t.Errorf("Expected log content in output, got: %s", output)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_ExportToFile(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "Export this content\n"
	createLogFile(t, configDir, logContent)

	tmpDir := filepath.Dir(configDir)
	outputFile := filepath.Join(tmpDir, "output.log")

	tf := factory.NewTestFactory(t)
	opts := &Options{Output: outputFile}

	err := run(tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify success message
	output := tf.OutString()
	if !strings.Contains(output, "exported") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify file contents
	exported, err := os.ReadFile(outputFile) //nolint:gosec // Test file path from TempDir
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}
	if string(exported) != logContent {
		t.Errorf("Exported content = %q, want %q", string(exported), logContent)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_WriteError(t *testing.T) {
	// Set up a temporary config dir with a log file
	configDir := setupTestConfigDir(t)
	logContent := "Test content\n"
	createLogFile(t, configDir, logContent)

	// Try to write to a path that doesn't exist (parent dir doesn't exist)
	tmpDir := filepath.Dir(configDir)
	outputFile := filepath.Join(tmpDir, "nonexistent-dir", "output.log")

	tf := factory.NewTestFactory(t)
	opts := &Options{Output: outputFile}

	err := run(tf.Factory, opts)
	if err == nil {
		t.Error("run() should error when write fails")
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_EmptyLogFile(t *testing.T) {
	// Set up a temporary config dir with an empty log file
	configDir := setupTestConfigDir(t)
	createLogFile(t, configDir, "")

	tf := factory.NewTestFactory(t)
	opts := &Options{} // Export to stdout

	err := run(tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Empty file should still work (just print nothing)
	output := tf.OutString()
	if output != "" {
		t.Errorf("Expected empty output for empty log file, got: %s", output)
	}
}

//nolint:paralleltest // Modifies XDG_CONFIG_HOME environment variable
func TestRun_LargeLogFile(t *testing.T) {
	// Set up a temporary config dir with a large log file
	configDir := setupTestConfigDir(t)

	// Create a log file with many lines
	var builder strings.Builder
	for i := range 1000 {
		builder.WriteString("2024-01-01 10:00:00 INFO Log line ")
		builder.WriteString(strconv.Itoa(i % 10))
		builder.WriteString("\n")
	}
	logContent := builder.String()
	createLogFile(t, configDir, logContent)

	tmpDir := filepath.Dir(configDir)
	outputFile := filepath.Join(tmpDir, "large-export.log")

	tf := factory.NewTestFactory(t)
	opts := &Options{Output: outputFile}

	err := run(tf.Factory, opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify file was exported correctly
	exported, err := os.ReadFile(outputFile) //nolint:gosec // Test file path from TempDir
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}
	if string(exported) != logContent {
		t.Error("Exported content doesn't match original")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	opts := &Options{}

	if opts.Output != "" {
		t.Errorf("Default Output = %q, want empty string", opts.Output)
	}
}
