package plugins

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewExecutor tests NewExecutor function.
func TestNewExecutor(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}
}

// TestExecutor_buildEnvironment tests buildEnvironment method.
func TestExecutor_buildEnvironment(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test-plugin",
		Dir:  "/home/user/.config/shelly/plugins/shelly-test",
	}

	env := executor.buildEnvironment(plugin)

	// Check that required environment variables are set
	var hasPluginDir, hasCliVersion, hasTheme, hasAPIMode bool
	for _, e := range env {
		if strings.HasPrefix(e, "SHELLY_PLUGIN_DIR=") {
			hasPluginDir = true
			if !strings.Contains(e, plugin.Dir) {
				t.Errorf("SHELLY_PLUGIN_DIR = %q, should contain %q", e, plugin.Dir)
			}
		}
		if strings.HasPrefix(e, "SHELLY_CLI_VERSION=") {
			hasCliVersion = true
		}
		if strings.HasPrefix(e, "SHELLY_THEME=") {
			hasTheme = true
		}
		if strings.HasPrefix(e, "SHELLY_API_MODE=") {
			hasAPIMode = true
		}
	}

	if !hasPluginDir {
		t.Error("SHELLY_PLUGIN_DIR not set in environment")
	}
	if !hasCliVersion {
		t.Error("SHELLY_CLI_VERSION not set in environment")
	}
	if !hasTheme {
		t.Error("SHELLY_THEME not set in environment")
	}
	if !hasAPIMode {
		t.Error("SHELLY_API_MODE not set in environment")
	}
}

// TestExecutor_buildEnvironment_NilPlugin tests buildEnvironment with nil plugin.
func TestExecutor_buildEnvironment_NilPlugin(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	env := executor.buildEnvironment(nil)

	// Should still have base environment variables
	var hasCliVersion bool
	for _, e := range env {
		if strings.HasPrefix(e, "SHELLY_CLI_VERSION=") {
			hasCliVersion = true
		}
	}

	if !hasCliVersion {
		t.Error("SHELLY_CLI_VERSION not set even with nil plugin")
	}
}

// TestExecutor_buildEnvironment_EmptyDir tests buildEnvironment with empty plugin dir.
func TestExecutor_buildEnvironment_EmptyDir(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test-plugin",
		Dir:  "", // Empty directory
	}

	env := executor.buildEnvironment(plugin)

	// SHELLY_PLUGIN_DIR should not be set when Dir is empty
	for _, e := range env {
		if strings.HasPrefix(e, "SHELLY_PLUGIN_DIR=") {
			t.Errorf("SHELLY_PLUGIN_DIR should not be set when Dir is empty: %s", e)
		}
	}
}

// TestExecutor_ExecuteContext_InvalidPath tests ExecuteContext with invalid plugin path.
func TestExecutor_ExecuteContext_InvalidPath(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "nonexistent",
		Path: "/nonexistent/path/to/plugin",
	}

	err := executor.ExecuteContext(context.Background(), plugin, []string{"--help"})
	if err == nil {
		t.Error("ExecuteContext() should fail with invalid plugin path")
	}
}

// TestExecutor_ExecuteCaptureContext_InvalidPath tests ExecuteCaptureContext with invalid path.
func TestExecutor_ExecuteCaptureContext_InvalidPath(t *testing.T) {
	t.Parallel()

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "nonexistent",
		Path: "/nonexistent/path/to/plugin",
	}

	_, err := executor.ExecuteCaptureContext(context.Background(), plugin, []string{"--version"})
	if err == nil {
		t.Error("ExecuteCaptureContext() should fail with invalid plugin path")
	}
}

// TestExecutor_ExecuteContext_WithRealScript tests ExecuteContext with a real script.
func TestExecutor_ExecuteContext_WithRealScript(t *testing.T) {
	t.Parallel()

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test script that exits successfully
	scriptPath := filepath.Join(tmpDir, "shelly-test")
	script := `#!/bin/bash
exit 0
`
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test",
		Path: scriptPath,
		Dir:  tmpDir,
	}

	err = executor.ExecuteContext(context.Background(), plugin, []string{})
	if err != nil {
		t.Errorf("ExecuteContext() error: %v", err)
	}
}

// TestExecutor_ExecuteCaptureContext_WithRealScript tests ExecuteCaptureContext with output.
func TestExecutor_ExecuteCaptureContext_WithRealScript(t *testing.T) {
	t.Parallel()

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test script that outputs version
	scriptPath := filepath.Join(tmpDir, "shelly-test")
	script := `#!/bin/bash
echo "test-plugin 1.0.0"
`
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test",
		Path: scriptPath,
		Dir:  tmpDir,
	}

	output, err := executor.ExecuteCaptureContext(context.Background(), plugin, []string{})
	if err != nil {
		t.Fatalf("ExecuteCaptureContext() error: %v", err)
	}

	if !strings.Contains(string(output), "test-plugin 1.0.0") {
		t.Errorf("output = %q, want to contain 'test-plugin 1.0.0'", string(output))
	}
}

// TestExecutor_Execute tests Execute method (wrapper for ExecuteContext).
func TestExecutor_Execute(t *testing.T) {
	t.Parallel()

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test script
	scriptPath := filepath.Join(tmpDir, "shelly-test")
	script := `#!/bin/bash
exit 0
`
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test",
		Path: scriptPath,
	}

	err = executor.Execute(plugin, []string{})
	if err != nil {
		t.Errorf("Execute() error: %v", err)
	}
}

// TestExecutor_ExecuteCapture tests ExecuteCapture method (wrapper for ExecuteCaptureContext).
func TestExecutor_ExecuteCapture(t *testing.T) {
	t.Parallel()

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test script
	scriptPath := filepath.Join(tmpDir, "shelly-test")
	script := `#!/bin/bash
echo "hello world"
`
	//nolint:gosec // Test file needs to be executable
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	executor := NewExecutor()
	plugin := &Plugin{
		Name: "test",
		Path: scriptPath,
	}

	output, err := executor.ExecuteCapture(plugin, []string{})
	if err != nil {
		t.Fatalf("ExecuteCapture() error: %v", err)
	}

	if !strings.Contains(string(output), "hello world") {
		t.Errorf("output = %q, want to contain 'hello world'", string(output))
	}
}

// TestRunPlugin_NotFound tests RunPlugin with non-existent plugin.
func TestRunPlugin_NotFound(t *testing.T) {
	t.Parallel()

	err := RunPlugin(context.Background(), "nonexistent-plugin-12345", []string{})
	if err == nil {
		t.Error("RunPlugin() should fail for non-existent plugin")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}
