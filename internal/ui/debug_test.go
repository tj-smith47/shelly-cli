package ui

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
)

// Note: Tests in this file modify global state (viper, stdout/stderr)
// and cannot run in parallel.

func resetViper() {
	viper.Reset()
}

// captureStderr captures stderr during test execution.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stderr = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("warning: failed to close pipe reader: %v", err)
	}

	return buf.String()
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestDebug_VerboseEnabled(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	output := captureStderr(t, func() {
		Debug("test message %d", 42)
	})

	if output != "debug: test message 42\n" {
		t.Errorf("Debug() output = %q, want %q", output, "debug: test message 42\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestDebug_VerboseDisabled(t *testing.T) {
	resetViper()
	viper.Set("verbose", false)

	output := captureStderr(t, func() {
		Debug("test message")
	})

	if output != "" {
		t.Errorf("Debug() should not output when verbose is false, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestDebugErr_VerboseEnabled(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	output := captureStderr(t, func() {
		DebugErr("operation", errors.New("test error"))
	})

	if output != "debug: operation: test error\n" {
		t.Errorf("DebugErr() output = %q, want %q", output, "debug: operation: test error\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestDebugErr_VerboseDisabled(t *testing.T) {
	resetViper()
	viper.Set("verbose", false)

	output := captureStderr(t, func() {
		DebugErr("operation", errors.New("test error"))
	})

	if output != "" {
		t.Errorf("DebugErr() should not output when verbose is false, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestDebugErr_NilError(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	output := captureStderr(t, func() {
		DebugErr("operation", nil)
	})

	if output != "" {
		t.Errorf("DebugErr() should not output for nil error, got %q", output)
	}
}

// mockCloser implements Close() for testing CloseWithDebug.
type mockCloser struct {
	closed bool
	err    error
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.err
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCloseWithDebug_Success(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	closer := &mockCloser{}

	output := captureStderr(t, func() {
		CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "" {
		t.Errorf("CloseWithDebug() should not output on success, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCloseWithDebug_Error(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	closer := &mockCloser{err: errors.New("close failed")}

	output := captureStderr(t, func() {
		CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "debug: test close: close failed\n" {
		t.Errorf("CloseWithDebug() output = %q, want %q", output, "debug: test close: close failed\n")
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCloseWithDebug_NilCloser(t *testing.T) {
	resetViper()
	viper.Set("verbose", true)

	// Should not panic with nil closer
	output := captureStderr(t, func() {
		CloseWithDebug("test close", nil)
	})

	if output != "" {
		t.Errorf("CloseWithDebug() should not output for nil closer, got %q", output)
	}
}

//nolint:paralleltest // Tests modify global state (viper, stdout/stderr)
func TestCloseWithDebug_ErrorVerboseDisabled(t *testing.T) {
	resetViper()
	viper.Set("verbose", false)

	closer := &mockCloser{err: errors.New("close failed")}

	output := captureStderr(t, func() {
		CloseWithDebug("test close", closer)
	})

	if !closer.closed {
		t.Error("Close() was not called")
	}
	if output != "" {
		t.Errorf("CloseWithDebug() should not output when verbose is false, got %q", output)
	}
}
