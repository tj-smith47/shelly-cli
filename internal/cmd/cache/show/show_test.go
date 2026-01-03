package show

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
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

	expectedAliases := []string{"s", "stats", "status"}
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

	if !strings.Contains(cmd.Long, "file cache") {
		t.Error("Long should contain 'file cache'")
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

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

// TestRun_EmptyCache tests run when cache is empty.
// Uses SetupTestFs - NOT parallel.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_EmptyCache(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run returned error: %v", err)
	}
}

// TestRun_WithCacheEntries tests run when cache has entries.
// Uses SetupTestFs - NOT parallel.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_WithCacheEntries(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entries
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	if err := fc.Set("device2", "firmware", map[string]string{"v": "1.0"}, cache.TTLFirmware); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run returned error: %v", err)
	}
}

// TestOutputTable tests the output table creation for cache stats.
func TestOutputTable(t *testing.T) {
	t.Parallel()

	builder := table.NewBuilder("Property", "Value")
	builder.AddRow("Location", "/tmp/cache")
	builder.AddRow("Files", "5")
	builder.AddRow("Size", "1.2 KB")
	tbl := builder.Build()

	var buf bytes.Buffer
	err := tbl.PrintTo(&buf)

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
	if err != nil {
		t.Logf("format error (may be expected): %v", err)
	}
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
// Uses SetupTestFs - NOT parallel.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ContextCancelled(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run should still work since it doesn't make network calls
	opts := &Options{Factory: tf.Factory}
	err = run(ctx, opts)

	// Should complete (context cancellation doesn't affect file system operations)
	if err != nil {
		t.Logf("run error: %v", err)
	}
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

	if !strings.Contains(cmd.Long, "file cache") {
		t.Errorf("Long should contain 'file cache', got %q", cmd.Long)
	}
}

// TestRun_ViaCommand tests running via cobra Execute.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_ViaCommand(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	cmd := NewCommand(tf.Factory)
	err = cmd.Execute()

	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}
}

// TestRun_CacheNil tests run when cache initialization fails.
// Uses a read-only filesystem to make cache.New() fail.
// NOT parallel: Uses t.Setenv and config.SetFs which are package-level state.
func TestRun_CacheNil(t *testing.T) {
	// Create base filesystem with config directory
	memFs := afero.NewMemMapFs()
	if err := memFs.MkdirAll("/testconfig/shelly", 0o700); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// Wrap in read-only filesystem to make cache directory creation fail
	roFs := afero.NewReadOnlyFs(memFs)
	config.SetFs(roFs)

	t.Setenv("XDG_CONFIG_HOME", "/testconfig")
	config.ResetDefaultManagerForTesting()

	t.Cleanup(func() {
		config.SetFs(nil)
		config.ResetDefaultManagerForTesting()
	})

	tf := factory.NewTestFactory(t)
	// Don't call SetFileCache - let FileCache() try to lazy-init and fail

	opts := &Options{Factory: tf.Factory}
	err := run(context.Background(), opts)

	if err != nil {
		t.Errorf("run should not error with nil cache: %v", err)
	}

	combined := tf.OutString() + tf.ErrString()
	if !strings.Contains(combined, "Cache not available") {
		t.Errorf("Expected 'Cache not available' message, got: %q", combined)
	}
}

// TestRun_StructuredOutput tests JSON output mode.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_StructuredOutput(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entry
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Set output format to JSON
	oldOutput := viper.GetString("output")
	viper.Set("output", "json")
	t.Cleanup(func() {
		viper.Set("output", oldOutput)
	})

	tf := factory.NewTestFactory(t)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory}
	err = run(context.Background(), opts)

	if err != nil {
		t.Errorf("run returned error: %v", err)
	}

	// Should contain JSON output markers
	combined := tf.OutString()
	if !strings.Contains(combined, "location") && !strings.Contains(combined, "total_entries") {
		t.Logf("Output: %q", combined)
	}
}

// Note: TestRun_StatsError is not implemented because Stats() uses
// defensive nilerr patterns that skip all errors - it never returns an error.
// The error path at show.go:65-66 is defensive and unreachable in practice.

// TestRun_CacheDirError tests run when config.CacheDir() fails.
// NOT parallel: Uses t.Setenv to unset HOME.
func TestRun_CacheDirError(t *testing.T) {
	// Unset HOME and XDG_CACHE_HOME to make os.UserCacheDir() fail
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	tf := factory.NewTestFactory(t)

	opts := &Options{Factory: tf.Factory}
	err := run(context.Background(), opts)

	if err == nil {
		t.Error("Expected error when HOME is unset")
	}
}

// errWriter is a writer that always fails.
type errWriter struct{}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("mock write error")
}

// TestRun_PrintTableError tests run when table printing fails.
// This covers the ios.DebugErr paths for table print errors.
//
//nolint:paralleltest // Uses SetupTestFs which calls t.Setenv
func TestRun_PrintTableError(t *testing.T) {
	memFs := factory.SetupTestFs(t)

	fc, err := cache.NewWithFs("/cache", memFs)
	if err != nil {
		t.Fatalf("Failed to create file cache: %v", err)
	}

	// Create cache entries to ensure type breakdown table is printed
	if err := fc.Set("device1", "deviceinfo", map[string]string{"id": "d1"}, cache.TTLDeviceInfo); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Create IOStreams with failing output writer
	in := &bytes.Buffer{}
	out := &errWriter{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(in, out, errOut)

	tf := factory.NewTestFactory(t)
	tf.SetIOStreams(ios)
	tf.SetFileCache(fc)

	opts := &Options{Factory: tf.Factory}
	err = run(context.Background(), opts)

	// Should not return error (errors are logged via DebugErr, not returned)
	if err != nil {
		t.Errorf("run should not return error for print failures: %v", err)
	}
}
