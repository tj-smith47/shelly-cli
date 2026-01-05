package show

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

const testConfigDir = "/test/config"
const testShellyDir = "/test/config/shelly"

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
	if cmd.Use != "show" {
		t.Errorf("Use = %q, want %q", cmd.Use, "show")
	}

	// Test Aliases
	expectedAliases := []string{"view", "cat"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test RunE
	if cmd.RunE == nil {
		t.Error("RunE is nil")
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
		{"lines", "n", "50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if flag.Shorthand != tt.shorthand {
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
		"shelly log show",
		"-n 100",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

// setupLogTest sets up the test filesystem with XDG_CONFIG_HOME.
// Uses config.SetFs for filesystem isolation.
func setupLogTest(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	config.SetFs(fs)
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("XDG_CONFIG_HOME", testConfigDir)
	return fs
}

// TestExecute_NoLogFile tests when the log file does not exist.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_NoLogFile(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir but no log file
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "No log file found") {
		t.Errorf("expected 'No log file found' message, got: %q", combined)
	}
	if !strings.Contains(combined, "Debug logging may not be enabled") {
		t.Errorf("expected debug logging hint, got: %q", combined)
	}
}

// TestExecute_EmptyLogFile tests when the log file exists but is empty.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_EmptyLogFile(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and empty log file
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"
	if err := afero.WriteFile(fs, logPath, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "Log file is empty") {
		t.Errorf("expected 'Log file is empty' message, got: %q", combined)
	}
}

// TestExecute_WithLogContent tests when the log file has content.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_WithLogContent(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and log file with content
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"
	logContent := `2024-01-15 10:00:00 INFO Starting shelly CLI
2024-01-15 10:00:01 DEBUG Connecting to device 192.168.1.100
2024-01-15 10:00:02 INFO Device connected successfully
`
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "INFO Starting shelly CLI") {
		t.Errorf("expected log content in output, got: %q", output)
	}
	if !strings.Contains(output, "Device connected successfully") {
		t.Errorf("expected log content in output, got: %q", output)
	}
}

// TestExecute_WithLinesFlag tests the -n/--lines flag.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_WithLinesFlag(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and log file with many lines
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"

	// Create log with 10 lines
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, "Line "+strings.Repeat("x", i))
	}
	logContent := strings.Join(lines, "\n") + "\n"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"-n", "3"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should only have the last 3 lines
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	if len(outputLines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(outputLines), output)
	}
	// Last line should be "Line xxxxxxxxxx" (10 x's)
	if !strings.Contains(output, "Line xxxxxxxxxx") {
		t.Errorf("expected last line to contain 'Line xxxxxxxxxx', got: %q", output)
	}
	// First line (of the 3 shown) should be line 8
	if !strings.Contains(output, "Line xxxxxxxx") {
		t.Errorf("expected to contain 'Line xxxxxxxx', got: %q", output)
	}
}

// TestExecute_AllLines tests when requested lines exceeds file length.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_AllLines(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and log file with few lines
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"

	// Create log with only 3 lines
	logContent := "Line 1\nLine 2\nLine 3\n"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	// Request more lines than exist
	cmd.SetArgs([]string{"-n", "100"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	// Should have all 3 lines
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	if len(outputLines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(outputLines), output)
	}
	if !strings.Contains(output, "Line 1") {
		t.Errorf("expected 'Line 1' in output, got: %q", output)
	}
}

// TestExecute_LongLinesFlag tests using the long --lines flag.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_LongLinesFlag(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and log file
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"

	logContent := "Line A\nLine B\nLine C\nLine D\nLine E\n"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--lines", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	if len(outputLines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(outputLines), output)
	}
	// Should have last 2 lines: D and E
	if !strings.Contains(output, "Line D") || !strings.Contains(output, "Line E") {
		t.Errorf("expected 'Line D' and 'Line E' in output, got: %q", output)
	}
}

// TestExecute_InvalidArgs tests with unexpected arguments.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_InvalidArgs(t *testing.T) {
	_ = setupLogTest(t)

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"unexpected", "args"})
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid args, got nil")
	}
}

// TestExecute_Aliases tests command execution via aliases.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestExecute_Aliases(t *testing.T) {
	fs := setupLogTest(t)

	// Create config dir and log file
	if err := fs.MkdirAll(testShellyDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	logPath := testShellyDir + "/shelly.log"
	logContent := "Test log line\n"
	if err := afero.WriteFile(fs, logPath, []byte(logContent), 0o600); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)

	// Test that aliases are set correctly (execution through parent would use aliases)
	if len(cmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}
	if cmd.Aliases[0] != "view" {
		t.Errorf("expected first alias to be 'view', got %q", cmd.Aliases[0])
	}
	if cmd.Aliases[1] != "cat" {
		t.Errorf("expected second alias to be 'cat', got %q", cmd.Aliases[1])
	}

	// Execute through the command
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Test log line") {
		t.Errorf("expected log content in output, got: %q", output)
	}
}

// TestOptions_DefaultLines tests that default lines value is correct.
func TestOptions_DefaultLines(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	flag := cmd.Flags().Lookup("lines")
	if flag == nil {
		t.Fatal("lines flag not found")
	}
	if flag.DefValue != "50" {
		t.Errorf("default lines = %q, want %q", flag.DefValue, "50")
	}
}
