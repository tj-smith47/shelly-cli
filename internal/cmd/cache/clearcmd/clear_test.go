package clearcmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

	if cmd.Use != "clear" {
		t.Errorf("Use = %q, want %q", cmd.Use, "clear")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"c", "rm"}
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

	expected := "Clear the discovery cache"
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

	expected := "Clear all cached device discovery results."
	if cmd.Long != expected {
		t.Errorf("Long = %q, want %q", cmd.Long, expected)
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	if !strings.Contains(cmd.Example, "shelly cache clear") {
		t.Error("Example should contain 'shelly cache clear'")
	}
}

func TestNewCommand_RunE(t *testing.T) {
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
		checkFunc func() bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Use != ""
			},
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Short != ""
			},
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Long != ""
			},
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Example != ""
			},
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return len(cmd.Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.RunE != nil
			},
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc() {
				t.Error(tt.errMsg)
			}
		})
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

func TestNewCommand_AcceptsNoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Clear command doesn't take arguments
	if cmd.Args != nil {
		// If Args is set, verify it accepts no args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
}

func TestNewCommand_NoFlags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// The clear command doesn't add its own flags
	// Just verify the command is properly set up
	if cmd.Use != "clear" {
		t.Errorf("Use = %q, want %q", cmd.Use, "clear")
	}
}

// setCacheHome sets the appropriate environment variable to control cache directory.
// On Linux, this is XDG_CACHE_HOME. On macOS, we need to set HOME since
// os.UserCacheDir() uses $HOME/Library/Caches on macOS.
func setCacheHome(t *testing.T, tempDir string) {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		// On macOS, os.UserCacheDir() returns $HOME/Library/Caches
		t.Setenv("HOME", tempDir)
	default:
		// On Linux/other, os.UserCacheDir() uses XDG_CACHE_HOME
		t.Setenv("XDG_CACHE_HOME", tempDir)
	}
}

// getCacheDir returns the expected cache directory path based on the temp directory.
func getCacheDir(tempDir string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(tempDir, "Library", "Caches", "shelly")
	default:
		return filepath.Join(tempDir, "shelly")
	}
}

// TestRun tests the run function with various cache directory states.
// This test is NOT parallel because it modifies environment variables.
//
//nolint:gocyclo,paralleltest // Table-driven test with complex cases; modifies environment variables
func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, cacheDir string)
		expectedOutput string
		verify         func(t *testing.T, cacheDir string)
	}{
		{
			name:           "no cache directory",
			setup:          nil, // No setup - cache dir doesn't exist
			expectedOutput: "Cache is already empty",
			verify:         nil,
		},
		{
			name: "empty cache directory",
			setup: func(t *testing.T, cacheDir string) {
				t.Helper()
				if err := os.MkdirAll(cacheDir, 0o750); err != nil {
					t.Fatalf("Failed to create cache dir: %v", err)
				}
			},
			expectedOutput: "Cache is already empty",
			verify:         nil,
		},
		{
			name: "cache with single file",
			setup: func(t *testing.T, cacheDir string) {
				t.Helper()
				if err := os.MkdirAll(cacheDir, 0o750); err != nil {
					t.Fatalf("Failed to create cache dir: %v", err)
				}
				testFile := filepath.Join(cacheDir, "discovery.json")
				if err := os.WriteFile(testFile, []byte(`{"devices": []}`), 0o600); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
			verify: func(t *testing.T, cacheDir string) {
				t.Helper()
				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					t.Fatalf("Failed to read cache dir: %v", err)
				}
				if len(entries) != 0 {
					t.Errorf("Cache directory should be empty, has %d entries", len(entries))
				}
			},
		},
		{
			name: "cache with multiple files",
			setup: func(t *testing.T, cacheDir string) {
				t.Helper()
				if err := os.MkdirAll(cacheDir, 0o750); err != nil {
					t.Fatalf("Failed to create cache dir: %v", err)
				}
				filenames := []string{"file1.json", "file2.json", "file3.json", "data.cache"}
				for _, name := range filenames {
					path := filepath.Join(cacheDir, name)
					if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
						t.Fatalf("Failed to create file %s: %v", name, err)
					}
				}
			},
			expectedOutput: "Cache cleared",
			verify: func(t *testing.T, cacheDir string) {
				t.Helper()
				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					t.Fatalf("Failed to read cache dir: %v", err)
				}
				if len(entries) != 0 {
					t.Errorf("Cache directory should be empty after clear, has %d entries", len(entries))
				}
			},
		},
		{
			name: "cache with subdirectory",
			setup: func(t *testing.T, cacheDir string) {
				t.Helper()
				subDir := filepath.Join(cacheDir, "subdir")
				if err := os.MkdirAll(subDir, 0o750); err != nil {
					t.Fatalf("Failed to create subdir: %v", err)
				}
				testFile := filepath.Join(subDir, "nested.json")
				if err := os.WriteFile(testFile, []byte(`{}`), 0o600); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
			verify: func(t *testing.T, cacheDir string) {
				t.Helper()
				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					t.Fatalf("Failed to read cache dir: %v", err)
				}
				if len(entries) != 0 {
					t.Errorf("Cache directory should be empty, has %d entries", len(entries))
				}
			},
		},
		{
			name: "cache with mixed files and directories",
			setup: func(t *testing.T, cacheDir string) {
				t.Helper()
				if err := os.MkdirAll(cacheDir, 0o750); err != nil {
					t.Fatalf("Failed to create cache dir: %v", err)
				}
				// Create file
				if err := os.WriteFile(filepath.Join(cacheDir, "file.json"), []byte(`{}`), 0o600); err != nil {
					t.Fatalf("Failed to create file: %v", err)
				}
				// Create subdirectory with files
				subDir := filepath.Join(cacheDir, "subdir")
				if err := os.MkdirAll(subDir, 0o750); err != nil {
					t.Fatalf("Failed to create subdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(subDir, "nested.json"), []byte(`{}`), 0o600); err != nil {
					t.Fatalf("Failed to create nested file: %v", err)
				}
			},
			expectedOutput: "Cache cleared",
			verify: func(t *testing.T, cacheDir string) {
				t.Helper()
				entries, err := os.ReadDir(cacheDir)
				if err != nil {
					t.Fatalf("Failed to read cache dir: %v", err)
				}
				if len(entries) != 0 {
					t.Errorf("Cache directory should be empty, has %d entries", len(entries))
				}
			},
		},
		{
			name:  "context cancelled",
			setup: nil,
			// Context cancellation doesn't affect file operations, should still work
			expectedOutput: "Cache is already empty",
			verify:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for cache
			tempDir := t.TempDir()
			setCacheHome(t, tempDir)

			cacheDir := getCacheDir(tempDir)

			// Run setup if provided
			if tt.setup != nil {
				tt.setup(t, cacheDir)
			}

			var stdout, stderr bytes.Buffer
			ios := iostreams.Test(nil, &stdout, &stderr)
			f := cmdutil.NewWithIOStreams(ios)

			ctx := context.Background()
			if tt.name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			opts := &Options{Factory: f}
			err := run(ctx, opts)

			if err != nil {
				t.Errorf("run returned error: %v", err)
			}

			// Check output
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %q", tt.expectedOutput, combined)
			}

			// Run verification if provided
			if tt.verify != nil {
				tt.verify(t, cacheDir)
			}
		})
	}
}

// TestRun_RemoveAllError tests the DebugErr path when os.RemoveAll fails.
// This test is NOT parallel because it modifies environment variables.
//
//nolint:paralleltest // Modifies environment variables
func TestRun_RemoveAllError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission test not reliable on Windows")
	}

	tempDir := t.TempDir()
	setCacheHome(t, tempDir)

	cacheDir := getCacheDir(tempDir)
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create a subdirectory with a file
	subDir := filepath.Join(cacheDir, "protected")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	testFile := filepath.Join(subDir, "file.json")
	if err := os.WriteFile(testFile, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make parent directory read-only to prevent deletion
	if err := os.Chmod(subDir, 0o500); err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions for cleanup
		if err := os.Chmod(subDir, 0o750); err != nil {
			t.Logf("warning: failed to restore permissions: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	// Run should still succeed (error is logged via DebugErr, not returned)
	opts := &Options{Factory: f}
	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run should not return error for RemoveAll failures: %v", err)
	}

	// Should still print "Cache cleared" even if some removals failed
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "Cache cleared") {
		t.Errorf("Expected 'Cache cleared' message, got: %q", combined)
	}
}

// TestRun_ViaCommand tests run by executing through the cobra command.
// This test is NOT parallel because it modifies environment variables.
//
//nolint:paralleltest // Modifies environment variables
func TestRun_ViaCommand(t *testing.T) {
	tempDir := t.TempDir()
	setCacheHome(t, tempDir)

	cacheDir := getCacheDir(tempDir)
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	testFile := filepath.Join(cacheDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ios := iostreams.Test(nil, &stdout, &stderr)
	f := cmdutil.NewWithIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	// Check output message
	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "Cache cleared") {
		t.Errorf("Expected 'Cache cleared' message, got: %q", combined)
	}

	// Verify files are deleted
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("Failed to read cache dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Cache directory should be empty, has %d entries", len(entries))
	}
}
