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
	if l != nil {
		t.Error("expected nil logger when SHELLY_TUI_DEBUG is not set")
		if err := l.Close(); err != nil {
			t.Logf("warning: close: %v", err)
		}
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
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("warning: close: %v", err)
		}
	}()

	// Log some entries
	l.Log("Dashboard", "Devices", 120, 40, "test view content")
	l.LogEvent("test event")

	// Close to flush
	if err := l.Close(); err != nil {
		t.Logf("warning: close: %v", err)
	}

	// Read log file - path is constructed from known tmpDir, not user input
	logPath := filepath.Join(tmpDir, ".config", "shelly", "tui-debug.log")
	content, err := os.ReadFile(logPath) //nolint:gosec // Path from t.TempDir()
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
	defer func() {
		if err := l.Close(); err != nil {
			t.Logf("warning: close: %v", err)
		}
	}()

	// Log same view twice
	l.Log("Dashboard", "Devices", 120, 40, "same content")
	initialSize := l.size
	l.Log("Dashboard", "Devices", 120, 40, "same content")

	// Size should not have changed
	if l.size != initialSize {
		t.Error("expected duplicate view to be skipped")
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
