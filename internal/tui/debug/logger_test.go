package debug

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Note: Tests that modify environment variables cannot run in parallel.

//nolint:paralleltest // Tests modify environment variables
func TestNew_DisabledByDefault(t *testing.T) {
	// Ensure env is not set
	if err := os.Unsetenv(EnvKey); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	l := New()
	// Logger is now always returned (for Shift+D toggle), but should be disabled
	if l == nil {
		t.Fatal("expected non-nil logger even when SHELLY_TUI_DEBUG is not set")
	}
	if l.Enabled() {
		t.Error("expected logger to be disabled when SHELLY_TUI_DEBUG is not set")
	}
}

//nolint:paralleltest // Tests modify environment variables
func TestNew_EnabledWithEnv(t *testing.T) {
	if err := os.Setenv(EnvKey, "1"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(EnvKey); err != nil {
			t.Logf("warning: unsetenv: %v", err)
		}
	}()

	l := New()
	if l == nil {
		t.Fatal("expected non-nil logger when SHELLY_TUI_DEBUG=1")
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("warning: close: %v", err)
		}
	}()

	if !l.Enabled() {
		t.Error("expected logger to be enabled")
	}

	// Verify session directory was created
	if l.SessionDir() == "" {
		t.Error("expected session directory to be set")
	}
}

//nolint:paralleltest // Tests modify environment variables
func TestLogger_Log(t *testing.T) {
	// Create temp dir for test log
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Logf("warning: restore HOME: %v", err)
		}
	}()

	if err := os.Setenv(EnvKey, "1"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(EnvKey); err != nil {
			t.Logf("warning: unsetenv: %v", err)
		}
	}()

	l := New()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}

	// Log some entries
	l.Log("Dashboard", "Devices", 120, 40, "test view content")
	l.LogEvent("test event")

	// Get the log path before closing
	logPath := filepath.Join(l.SessionDir(), MainLogFile)

	// Close to flush
	if err := l.Close(); err != nil {
		t.Logf("warning: close: %v", err)
	}

	// Read log file
	content, err := os.ReadFile(logPath) //nolint:gosec // Path from SessionDir()
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	// Verify content
	contentStr := string(content)
	if !strings.Contains(contentStr, "Tab: Dashboard") {
		t.Error("expected log to contain tab name")
	}
	if !strings.Contains(contentStr, "Focus: Devices") {
		t.Error("expected log to contain focus name")
	}
	if !strings.Contains(contentStr, "Size: 120x40") {
		t.Error("expected log to contain size")
	}
	if !strings.Contains(contentStr, "test view content") {
		t.Error("expected log to contain view content")
	}
	if !strings.Contains(contentStr, "test event") {
		t.Error("expected log to contain event")
	}
	if !strings.Contains(contentStr, "Session Ended") {
		t.Error("expected log to contain session end footer")
	}
}

//nolint:paralleltest // Tests modify environment variables
func TestLogger_SkipDuplicateViews(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Logf("warning: restore HOME: %v", err)
		}
	}()

	if err := os.Setenv(EnvKey, "1"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(EnvKey); err != nil {
			t.Logf("warning: unsetenv: %v", err)
		}
	}()

	l := New()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}

	// Log same view twice, different view once
	l.Log("Dashboard", "Devices", 120, 40, "same content")
	l.Log("Dashboard", "Devices", 120, 40, "same content") // Should be skipped
	l.Log("Dashboard", "Devices", 120, 40, "different content")

	logPath := filepath.Join(l.SessionDir(), MainLogFile)

	if err := l.Close(); err != nil {
		t.Logf("warning: close: %v", err)
	}

	// Read and count occurrences
	content, err := os.ReadFile(logPath) //nolint:gosec // Path from SessionDir()
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	contentStr := string(content)

	// "same content" should appear only once (duplicate skipped)
	sameCount := strings.Count(contentStr, "same content")
	if sameCount != 1 {
		t.Errorf("expected 'same content' to appear once, got %d", sameCount)
	}

	// "different content" should appear once
	diffCount := strings.Count(contentStr, "different content")
	if diffCount != 1 {
		t.Errorf("expected 'different content' to appear once, got %d", diffCount)
	}
}

//nolint:paralleltest // Tests modify environment variables
func TestLogger_SessionDir(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Logf("warning: restore HOME: %v", err)
		}
	}()

	if err := os.Setenv(EnvKey, "1"); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	defer func() {
		if err := os.Unsetenv(EnvKey); err != nil {
			t.Logf("warning: unsetenv: %v", err)
		}
	}()

	l := New()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("warning: close: %v", err)
		}
	}()

	sessionDir := l.SessionDir()

	// Verify session directory exists
	info, err := os.Stat(sessionDir)
	if err != nil {
		t.Fatalf("session dir stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected session dir to be a directory")
	}

	// Verify it's under debug/
	if !strings.Contains(sessionDir, DebugDir) {
		t.Errorf("expected session dir to contain '%s', got: %s", DebugDir, sessionDir)
	}
}

func TestLogger_NilSafe(t *testing.T) {
	t.Parallel()

	var l *Logger

	// These should not panic
	l.Log("Dashboard", "Devices", 120, 40, "content")
	l.LogEvent("event")
	if err := l.Close(); err != nil {
		t.Errorf("expected nil Close to return nil, got: %v", err)
	}
	if l.Enabled() {
		t.Error("expected nil logger to not be enabled")
	}
	if l.SessionDir() != "" {
		t.Error("expected nil logger SessionDir to return empty string")
	}
}

func TestLogger_Writer(t *testing.T) {
	t.Parallel()

	var l *Logger

	// Nil logger returns discard
	w := l.Writer()
	if w == nil {
		t.Error("expected non-nil writer")
	}

	// Should be able to write without error
	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Errorf("write error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}
}

//nolint:paralleltest // Tests modify environment variables
func TestLogger_Toggle(t *testing.T) {
	// Use temp dir for HOME
	tmpDir := t.TempDir()
	homeDir := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("setenv HOME: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", homeDir); err != nil {
			t.Logf("warning: restore HOME: %v", err)
		}
	}()

	// Ensure env is not set - start disabled
	if err := os.Unsetenv(EnvKey); err != nil {
		t.Logf("warning: unsetenv: %v", err)
	}

	l := New()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}

	// Initially disabled
	if l.Enabled() {
		t.Error("expected logger to be disabled initially")
	}

	// Toggle ON
	enabled, sessionDir := l.Toggle()
	if !enabled {
		t.Error("expected Toggle to enable logging")
	}
	if sessionDir == "" {
		t.Error("expected Toggle to return session directory")
	}
	if !l.Enabled() {
		t.Error("expected logger to be enabled after Toggle")
	}

	// Log something
	l.Log("Test", "Panel", 80, 24, "test content")

	// Toggle OFF
	enabled, sessionDir = l.Toggle()
	if enabled {
		t.Error("expected Toggle to disable logging")
	}
	if sessionDir != "" {
		t.Error("expected Toggle to return empty session directory when disabling")
	}
	if l.Enabled() {
		t.Error("expected logger to be disabled after second Toggle")
	}

	// Toggle ON again - creates new session
	enabled, newSessionDir := l.Toggle()
	if !enabled {
		t.Error("expected Toggle to enable logging again")
	}
	if newSessionDir == "" {
		t.Error("expected new session directory")
	}

	// Clean up
	if err := l.Close(); err != nil {
		t.Logf("warning: close: %v", err)
	}
}
