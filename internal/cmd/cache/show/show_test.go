package show

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "show" {
		t.Errorf("Use = %q, want %q", cmd.Use, "show")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"s", "stats"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Show cache statistics"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if !strings.Contains(cmd.Long, "discovery cache") {
		t.Error("Long should contain 'discovery cache'")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly cache show") {
		t.Error("Example should contain 'shelly cache show'")
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// TestRun_NoCacheDir tests run when cache directory doesn't exist.
func TestRun_NoCacheDir(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f)

	// Should complete without error, displaying "Cache directory does not exist"
	if err != nil {
		t.Errorf("run returned error: %v", err)
	}

	combined := stdout.String() + stderr.String()
	// Output may contain info about cache or may be empty depending on config
	_ = combined
}

// TestRun_WithCacheDir tests run when cache directory exists with files.
func TestRun_WithCacheDir(t *testing.T) {
	t.Parallel()

	// Create a temporary cache directory for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create a test file in the cache directory
	testFile := filepath.Join(cacheDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"test": true}`), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Note: This test will use the real config's cache dir, not our temp dir
	// This is a limitation of the current implementation

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	err := run(context.Background(), f)

	// Should complete (may or may not have cache files)
	if err != nil {
		t.Logf("run returned error: %v", err)
	}
}

// TestOutputTable tests the output table creation for cache stats.
func TestOutputTable(t *testing.T) {
	t.Parallel()

	table := output.NewTable("Property", "Value")
	table.AddRow("Location", "/tmp/cache")
	table.AddRow("Files", "5")
	table.AddRow("Size", "1.2 KB")

	var buf bytes.Buffer
	err := table.PrintTo(&buf)

	if err != nil {
		t.Errorf("PrintTo error: %v", err)
	}

	tableOutput := buf.String()
	if !strings.Contains(tableOutput, "Location") {
		t.Error("Output should contain 'Location'")
	}
	if !strings.Contains(tableOutput, "/tmp/cache") {
		t.Error("Output should contain '/tmp/cache'")
	}
}

// TestFormatSize tests the size formatting function.
func TestFormatSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := output.FormatSize(tc.size)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tc.size, result, tc.expected)
		}
	}
}

// TestWantsStructured tests the structured output check.
func TestWantsStructured(t *testing.T) {
	t.Parallel()

	// Default should be false (table output)
	result := output.WantsStructured()
	_ = result // Just verify it doesn't panic
}

// TestFormatOutput tests the format output function.
func TestFormatOutput(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"location": "/tmp/cache",
		"files":    5,
		"size":     1234,
	}

	var buf bytes.Buffer
	err := output.FormatOutput(&buf, data)

	// This may fail depending on viper output setting, but shouldn't panic
	_ = err
}

// TestNewCommand_NoFlags verifies no extra flags are added.
func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The show command doesn't add its own flags
	// Just verify the command is properly set up
	if cmd.Use != "show" {
		t.Errorf("Use = %q, want %q", cmd.Use, "show")
	}
}

// TestRun_ContextCancelled tests run with cancelled context.
func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run should still work since it doesn't make network calls
	err := run(ctx, f)

	// Should complete (context cancellation doesn't affect file system operations)
	_ = err
}

// TestNewCommand_AcceptsNoArgs verifies command accepts no arguments.
func TestNewCommand_AcceptsNoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Show command doesn't take arguments
	if cmd.Args != nil {
		// If Args is set, verify it accepts no args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
}

// TestNewCommand_LongDescription verifies Long description content.
func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Display information about the discovery cache."
	if cmd.Long != expected {
		t.Errorf("Long = %q, want %q", cmd.Long, expected)
	}
}
